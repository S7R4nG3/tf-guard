package tfGuard

import (
	"encoding/json"
	"strings"
	"testing"

	tfresources "github.com/S7R4nG3/terraform-resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/sirupsen/logrus"
)

func TestStringResultFormatter(t *testing.T) {
	testResults := []Result{
		{
			Name:     "passing rule",
			Valid:    true,
			Severity: Severity.Minor,
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_bucket.default",
					Type:    "aws_s3_bucket",
				},
			},
		},
		{
			Name:     "failing rule",
			Valid:    false,
			Severity: Severity.Major,
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_object.obj",
					Type:    "aws_s3_object",
				},
			},
		},
	}
	remediationResults := []Result{
		{
			Name:               "failing rule with remediation",
			Valid:              false,
			Severity:           Severity.Major,
			RemediationMessage: "Add an Owner tag to this resource.",
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_bucket.default",
					Type:    "aws_s3_bucket",
				},
			},
		},
	}
	tests := []struct {
		name       string
		deployment Deployment
		want       []string
		notWant    []string
	}{
		{
			name: "Default output should only detail invalid rules.",
			deployment: Deployment{
				Results: testResults,
			},
			want:    []string{"failing rule", "Overall Resource Score: 50%"},
			notWant: []string{"passing rule"},
		},
		{
			name: "Verbose output should detail all rules.",
			deployment: Deployment{
				Results:       testResults,
				VerboseStdOut: true,
			},
			want: []string{"passing rule", "failing rule", "Overall Resource Score: 50%"},
		},
		{
			name: "All valid results should only output the score.",
			deployment: Deployment{
				Results: testResults[:1],
			},
			want:    []string{"Overall Resource Score: 100%"},
			notWant: []string{"passing rule"},
		},
		{
			name: "A provided remediation message should be included in the output.",
			deployment: Deployment{
				Results: remediationResults,
			},
			want: []string{"Remediation: Add an Owner tag to this resource."},
		},
		{
			name: "An empty remediation message should be omitted from the output.",
			deployment: Deployment{
				Results:       testResults,
				VerboseStdOut: true,
			},
			notWant: []string{"Remediation:"},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %v", tt.name)
		d := tt.deployment
		d.Debug = true
		d.Logger = logrus.New()
		d.stringResultFormatter()
		got := d.ResultsStdOut
		for _, want := range tt.want {
			if !strings.Contains(got, want) {
				t.Errorf("Test Error -- %s:\ngot: %v\n::\nexpected to contain: %v\n", tt.name, got, want)
			}
		}
		for _, notWant := range tt.notWant {
			if strings.Contains(got, notWant) {
				t.Errorf("Test Error -- %s:\ngot: %v\n::\nexpected NOT to contain: %v\n", tt.name, got, notWant)
			}
		}
	}
}

func TestJsonResultFormatter(t *testing.T) {
	testResults := []Result{
		{
			Name:     "passing rule",
			Valid:    true,
			Severity: Severity.Minor,
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_bucket.default",
					Type:    "aws_s3_bucket",
				},
			},
		},
		{
			Name:     "failing rule",
			Valid:    false,
			Severity: Severity.Major,
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_bucket.default",
					Type:    "aws_s3_bucket",
				},
			},
		},
		{
			Name:     "failing rule",
			Valid:    false,
			Severity: Severity.Major,
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_object.obj",
					Type:    "aws_s3_object",
				},
			},
		},
	}
	tests := []struct {
		name        string
		deployment  Deployment
		wantValid   []string
		wantInvalid []string
	}{
		{
			name: "Results should be grouped by validity.",
			deployment: Deployment{
				Results: testResults,
			},
			wantValid:   []string{"passing rule"},
			wantInvalid: []string{"failing rule", "failing rule"},
		},
		{
			name: "All valid results should produce an empty invalid grouping.",
			deployment: Deployment{
				Results: testResults[:1],
			},
			wantValid:   []string{"passing rule"},
			wantInvalid: []string{},
		},
		{
			name: "All invalid results should produce an empty valid grouping.",
			deployment: Deployment{
				Results: testResults[1:],
			},
			wantValid:   []string{},
			wantInvalid: []string{"failing rule", "failing rule"},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %v", tt.name)
		d := tt.deployment
		d.Debug = true
		d.Logger = logrus.New()
		d.jsonResultFormatter()
		var got jsonResponse
		if err := json.Unmarshal(d.ResultsJson, &got); err != nil {
			t.Errorf("Error unmarshalling test JSON data -- %s", tt.name)
			continue
		}
		if got.ByValid == nil || got.ByInvalid == nil {
			t.Errorf("Test Error -- %s:\nByValid and ByInvalid should always be present in the JSON output\n", tt.name)
		}
		if len(got.ByValid) != len(tt.wantValid) {
			t.Errorf("Test Error -- %s:\ngot: %v valid results\n::\nwant: %v\n", tt.name, len(got.ByValid), len(tt.wantValid))
		}
		for i, want := range tt.wantValid {
			if got.ByValid[i].Name != want || !got.ByValid[i].Valid {
				t.Errorf("Test Error -- %s:\ngot: %v\n::\nwant a valid result named: %v\n", tt.name, got.ByValid[i], want)
			}
		}
		if len(got.ByInvalid) != len(tt.wantInvalid) {
			t.Errorf("Test Error -- %s:\ngot: %v invalid results\n::\nwant: %v\n", tt.name, len(got.ByInvalid), len(tt.wantInvalid))
		}
		for i, want := range tt.wantInvalid {
			if got.ByInvalid[i].Name != want || got.ByInvalid[i].Valid {
				t.Errorf("Test Error -- %s:\ngot: %v\n::\nwant an invalid result named: %v\n", tt.name, got.ByInvalid[i], want)
			}
		}
		if len(got.ByValid)+len(got.ByInvalid) != got.TotalResults {
			t.Errorf("Test Error -- %s:\ngot: %v valid + %v invalid\n::\nwant them to total: %v\n", tt.name, len(got.ByValid), len(got.ByInvalid), got.TotalResults)
		}
	}
}
