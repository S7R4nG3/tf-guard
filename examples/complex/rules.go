package main

import (
	"encoding/json"
	"fmt"
	"strings"

	tfGuard "github.com/S7R4nG3/tf-guard"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

// Where each resource type holds its IAM policy document.
var policyDocumentAttributes = map[string]string{
	"aws_iam_policy": "policy",
	"aws_iam_role":   "assume_role_policy",
}

// Validating a setting flagged as "known after apply".
//
// Policy documents assembled by an aws_iam_policy_document data source are
// only rendered during the apply, leaving nothing in the planned values to
// inspect. KnownAfterApply distinguishes that from a resource declaring no
// policy at all. Has suffices here - the document is a single string attribute.
func RuleIamPoliciesMustNotAllowWildcardActions(r tfresources.Resource) tfGuard.Result {
	name := "IAM policy documents must not grant wildcard actions."
	severity := tfGuard.Severity.Critical
	attribute, applies := policyDocumentAttributes[r.Planned.Type]
	if !applies {
		return tfGuard.Result{NotApplicable: true}
	}
	// A resource being destroyed leaves no policy behind to evaluate.
	if r.Change.Actions.Delete() {
		return tfGuard.Result{NotApplicable: true}
	}
	document, configured := r.Planned.AttributeValues[attribute]
	if !configured {
		// Rendered during the apply, so report the deferral rather than
		// failing on a value the rule was never given.
		if r.KnownAfterApply.Has(attribute) {
			return tfGuard.Result{
				Name:               name,
				Valid:              true,
				Severity:           severity,
				RemediationMessage: fmt.Sprintf("The %s document is known after apply - review the rendered policy once the plan has been applied. Unknown attributes: %v", attribute, r.KnownAfterApply.Paths()),
			}
		}
		return tfGuard.Result{
			Name:               name,
			Valid:              false,
			Severity:           severity,
			RemediationMessage: fmt.Sprintf("No %s document is declared for this resource.", attribute),
		}
	}
	if documentAllowsWildcardActions(document) {
		return tfGuard.Result{
			Name:               name,
			Valid:              false,
			Severity:           severity,
			RemediationMessage: "Replace the wildcard action with the specific actions this policy requires.",
		}
	}
	return tfGuard.Result{
		Name:     name,
		Valid:    true,
		Severity: severity,
	}
}

// A malformed document reports clean - rejecting it is the provider's job.
func documentAllowsWildcardActions(document any) bool {
	raw, isString := document.(string)
	if !isString {
		return false
	}
	var policy struct {
		Statement []struct {
			Action any `json:"Action"`
		} `json:"Statement"`
	}
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return false
	}
	for _, statement := range policy.Statement {
		switch action := statement.Action.(type) {
		case string:
			if action == "*" {
				return true
			}
		case []any:
			for _, entry := range action {
				if entry == "*" {
					return true
				}
			}
		}
	}
	return false
}

func RuleModulesMustBeSourcedFromRegistry(r tfresources.Resource) tfGuard.Result {
	name := "Modules must be sourced from the Terraform Registry."
	severity := tfGuard.Severity.Critical
	if r.Module.Source != "" {
		if !strings.Contains(r.Module.Source, "registry.terraform.io") {
			return tfGuard.Result{
				Name:     name,
				Severity: severity,
				Valid:    false,
			}
		} else {
			return tfGuard.Result{
				Name:     name,
				Severity: severity,
				Valid:    true,
			}
		}
	}
	return tfGuard.Result{NotApplicable: true}
}

func RuleS3BucketMustBeTagged(res tfresources.Resource) tfGuard.Result {
	name := "S3 buckets must be tagged at all times."
	if res.Planned.Type == "aws_s3_bucket" {
		if tags, exists := res.Planned.AttributeValues["tags"]; exists && tags == nil {
			return tfGuard.Result{
				Name:  name,
				Valid: true,
			}
		} else {
			return tfGuard.Result{
				Name:  name,
				Valid: false,
			}
		}
	}
	return tfGuard.Result{NotApplicable: true}
}

func RuleS3BucketsShouldNotHaveForceDestroy(r tfresources.Resource) tfGuard.Result {
	name := "S3 Buckets should not have Force Destroy enabled."
	severity := tfGuard.Severity.Major
	if r.Planned.Type == "aws_s3_bucket" {
		if forceDestroy, exists := r.Planned.AttributeValues["force_destroy"]; exists && forceDestroy.(bool) {
			return tfGuard.Result{
				Name:     name,
				Severity: severity,
				Valid:    false,
			}
		} else {
			return tfGuard.Result{
				Name:     name,
				Severity: severity,
				Valid:    true,
			}
		}
	}
	return tfGuard.Result{NotApplicable: true}
}
