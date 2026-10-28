package tfGuard

import (
	"encoding/json"
	"testing"

	tfresources "github.com/S7R4nG3/terraform-resources"
	"github.com/sirupsen/logrus"
)

func TestDeployment(t *testing.T) {
	testRules := []Rule{
		func(t tfresources.Resource) Result {
			if t.Planned.Type != "" {
				return Result{
					Name:     "Test Rule",
					Severity: Severity.Major,
					Valid:    true,
				}
			}
			return Result{}
		},
	}
	// Rules that apply to no resource in the null_resource plan, leaving
	// the Deployment with an empty result set.
	notApplicableRules := []Rule{
		func(t tfresources.Resource) Result {
			if t.Planned.Type == "aws_s3_bucket" {
				return Result{
					Name:     "Test Rule",
					Severity: Severity.Major,
					Valid:    true,
				}
			}
			return Result{NotApplicable: true}
		},
	}
	tests := []struct {
		name       string
		deployment Deployment
		want       float64
	}{
		{
			name: "Simple deployment should execute cleanly.",
			deployment: Deployment{
				PlanFile: "./testdata/simple/plan.json",
				Rules:    testRules,
				Debug:    true,
				Logger:   logrus.New(),
			},
			want: 100.00,
		},
		{
			name: "Complex deployment should execute cleanly.",
			deployment: Deployment{
				PlanFile: "./testdata/complex/plan.json",
				Rules:    testRules,
				Debug:    true,
				Logger:   logrus.New(),
			},
			want: 100.00,
		},
		{
			name: "A deployment producing no applicable results should still marshal cleanly.",
			deployment: Deployment{
				PlanFile: "./testdata/null/plan.json",
				Rules:    notApplicableRules,
				Debug:    true,
				Logger:   logrus.New(),
			},
			want: 100.00,
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %v", tt.name)
		d := tt.deployment
		d.Scan()
		if len(d.ResultsJson) == 0 {
			t.Errorf("Test Error -- %s:\nResultsJson is empty, the results failed to marshal\n", tt.name)
			continue
		}
		var got jsonResponse
		err := json.Unmarshal(d.ResultsJson, &got)
		if err != nil {
			t.Errorf("Error unmarshalling test JSON data -- %s", tt.name)
		}
		if got.Score != tt.want {
			t.Errorf("Error with test:\ngot: %v\n::\nwant: %v\n", got, tt.want)
		}
	}
}

// The planned change and the known after apply attributes must reach the
// Rules, so a rule can tell a deferred attribute from an unconfigured one.
func TestKnownAfterApplyIsExposedToRules(t *testing.T) {
	// This bucket configures its tags but defers its encryption, exercising
	// both halves of that distinction.
	const (
		address    = "aws_s3_bucket.default"
		configured = "tags"
		deferred   = "server_side_encryption_configuration"
	)
	d := Deployment{
		PlanFile:      "./testdata/simple/plan.json",
		DisableStdOut: true,
		Rules: []Rule{
			func(r tfresources.Resource) Result {
				if r.Planned.Address != address {
					return Result{NotApplicable: true}
				}
				return Result{Name: "known after apply", Valid: true}
			},
		},
	}
	d.Scan()

	if len(d.Results) != 1 {
		t.Fatalf("Test Error -- expected a single result for %s, got %v", address, len(d.Results))
	}
	resource := d.Results[0].Resource

	if !resource.Change.Actions.Create() {
		t.Errorf("Test Error -- expected the planned change for %s to be a create, got %v", address, resource.Change.Actions)
	}
	if _, set := resource.Planned.AttributeValues[configured]; !set {
		t.Errorf("Test Error -- expected %s to be configured in the planned values for %s", configured, address)
	}
	if _, set := resource.Planned.AttributeValues[deferred]; set {
		t.Errorf("Test Error -- expected %s to be absent from the planned values for %s", deferred, address)
	}
	if !resource.KnownAfterApply.HasPrefix(deferred) {
		t.Errorf("Test Error -- expected %s to be known after apply for %s, got %v", deferred, address, resource.KnownAfterApply.Paths())
	}
	if resource.KnownAfterApply.Has(configured) {
		t.Errorf("Test Error -- expected the configured attribute %s not to be known after apply for %s", configured, address)
	}

	// The re-exported aliases must name the types the Resource carries.
	var (
		change     Change          = resource.Change
		unknowns   KnownAfterApply = resource.KnownAfterApply
		attributes []UnknownAttribute
	)
	attributes = append(attributes, unknowns...)
	if len(attributes) == 0 || len(change.Actions) == 0 {
		t.Errorf("Test Error -- expected the re-exported change types to carry the resource's planned change")
	}
}

// Resources within nested modules must be linked to the module that declares
// them, with local sources resolved against their parent modules so a Rule can
// identify exactly where a resource's module is sourced from.
func TestNestedModulesAreExposedToRules(t *testing.T) {
	const remote = "git::https://example.com/org/terraform-modules.git"
	tests := []struct {
		address        string
		key            string
		source         string
		resolvedSource string
		root           string
	}{
		{
			address:        `module.app["blue"].module.policy.terraform_data.policy`,
			key:            "app.policy",
			source:         "../policy",
			resolvedSource: "./modules/policy",
			root:           "app",
		},
		{
			address:        `module.app["green"].module.bucket[1].terraform_data.bucket`,
			key:            "app.bucket",
			source:         "./bucket",
			resolvedSource: "./modules/app/bucket",
			root:           "app",
		},
		{
			address:        "module.remote.module.sibling.terraform_data.sibling",
			key:            "remote.sibling",
			source:         "../sibling",
			resolvedSource: remote + "//modules/sibling?ref=v1.0.0",
			root:           "remote",
		},
		{
			address:        "module.remote.module.inner.terraform_data.inner",
			key:            "remote.inner",
			source:         "./inner",
			resolvedSource: remote + "//modules/outer/inner?ref=v1.0.0",
			root:           "remote",
		},
	}
	d := Deployment{
		PlanFile:      "./testdata/nested/plan.json",
		ModulesFile:   "./testdata/nested/modules.json",
		DisableStdOut: true,
		Rules: []Rule{
			func(r tfresources.Resource) Result {
				return Result{Name: "nested modules", Valid: true}
			},
		},
	}
	d.Scan()

	resources := make(map[string]tfresources.Resource)
	for _, res := range d.Results {
		resources[res.Resource.Planned.Address] = res.Resource
	}
	if root, found := resources["terraform_data.root"]; !found || root.Module.Key != "" {
		t.Errorf("Test Error -- expected terraform_data.root to be linked to no module, got %+v", root.Module)
	}
	for _, tt := range tests {
		r, found := resources[tt.address]
		if !found {
			t.Errorf("Test Error -- expected a result for %s", tt.address)
			continue
		}
		if r.Module.Key != tt.key || r.Module.Source != tt.source || r.Module.ResolvedSource != tt.resolvedSource {
			t.Errorf("Test Error -- %s:\ngot: key %q, source %q, resolved source %q\n::\nwant: key %q, source %q, resolved source %q\n",
				tt.address, r.Module.Key, r.Module.Source, r.Module.ResolvedSource, tt.key, tt.source, tt.resolvedSource)
		}
		if r.Module.Parent == nil || r.Module.Root().Key != tt.root {
			t.Errorf("Test Error -- expected %s to lead back to the %s module, got %+v", tt.address, tt.root, r.Module.Parent)
		}
	}
}
