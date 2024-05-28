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
		},
		Debug: true,
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
