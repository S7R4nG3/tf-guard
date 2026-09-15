package main

import (
	tfGuard "github.com/S7R4nG3/tf-guard"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

func main() {
	g := tfGuard.Deployment{
		PlanFile: "../../testdata/simple/plan.json",
		Rules: []tfGuard.Rule{
			RuleS3BucketMustBeTagged,
			RuleBucketObjectsMustBeTagged,
			RuleResourcesMustHaveOwnerTag,
			RuleS3BucketsMustBeEncrypted,
		},
		Debug: true,
		// The encryption rule passes, and passing results are hidden by default.
		VerboseStdOut: true,
	}
	g.Scan()
}

func RuleResourcesMustHaveOwnerTag(res tfresources.Resource) tfGuard.Result {
	name := "All resources must include an Owner tag."
	severity := tfGuard.Severity.Major
	if _, exists := res.Planned.AttributeValues["owner"]; exists {
		return tfGuard.Result{
			Name:     name,
			Valid:    true,
			Severity: severity,
		}
	} else {
		return tfGuard.Result{
			Name:     name,
			Valid:    false,
			Severity: severity,
		}
	}
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

// Validating a setting flagged as "known after apply".
//
// Encryption is usually managed by a separate resource, leaving this attribute
// absent from the planned values just as it would be if never configured at
// all. KnownAfterApply tells the two apart. HasPrefix matches the block whether
// Terraform reports it as unknown in whole or only in part.
func RuleS3BucketsMustBeEncrypted(res tfresources.Resource) tfGuard.Result {
	name := "S3 buckets must have server-side encryption configured."
	encryption := "server_side_encryption_configuration"
	if res.Planned.Type != "aws_s3_bucket" {
		return tfGuard.Result{NotApplicable: true}
	}
	if _, configured := res.Planned.AttributeValues[encryption]; configured {
		return tfGuard.Result{
			Name:     name,
			Valid:    true,
			Severity: tfGuard.Severity.Critical,
		}
	}
	// Unresolvable until the apply, so it cannot be failed. Passed here with
	// the deferral noted - NotApplicable would drop it from the score instead.
	if res.KnownAfterApply.HasPrefix(encryption) {
		return tfGuard.Result{
			Name:               name,
			Valid:              true,
			Severity:           tfGuard.Severity.Critical,
			RemediationMessage: "Encryption is known after apply for this bucket - confirm it is enabled once the plan has been applied.",
		}
	}
	return tfGuard.Result{
		Name:               name,
		Valid:              false,
		Severity:           tfGuard.Severity.Critical,
		RemediationMessage: "Add a server_side_encryption_configuration block, or an aws_s3_bucket_server_side_encryption_configuration resource, for this bucket.",
	}
}

func RuleBucketObjectsMustBeTagged(res tfresources.Resource) tfGuard.Result {
	if res.Planned.Type == "aws_s3_object" {
		if _, exists := res.Planned.AttributeValues["tags"]; exists {
			return tfGuard.Result{
				Name:  "S3 Bucket Objects must always be tagged.",
				Valid: true,
			}
		}
	}
	return tfGuard.Result{NotApplicable: true}
}
