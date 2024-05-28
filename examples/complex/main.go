package main

import (
	tfGuard "github.com/S7R4nG3/tf-guard"
)

func main() {
	g := tfGuard.Deployment{
		PlanFile:    "../../testdata/complex/plan.json",
		ModulesFile: "../../testdata/complex/modules.json",
		Rules: []tfGuard.Rule{
			RuleModulesMustBeSourcedFromRegistry,
			RuleS3BucketMustBeTagged,
			RuleS3BucketsShouldNotHaveForceDestroy,
		},
		// Debug: true,
	}
	g.Scan()
}
