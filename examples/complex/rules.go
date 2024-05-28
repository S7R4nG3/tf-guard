package main

import (
	"strings"

	tfGuard "github.com/S7R4nG3/tf-guard"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

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
