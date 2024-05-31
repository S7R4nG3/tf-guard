package tfGuard

import (
	"fmt"

	tfresources "github.com/S7R4nG3/terraform-resources"
	"github.com/sirupsen/logrus"
)

// The primary Scan method intakes the specified Plan and
// Modules file paths and utilizes the tfresources project
// to parse out the underlying Terraform resources and link
// them to any declared modules.
//
// Once parsed, the resources are fed into the rule engine
// to begin the execution of each rule onto each resource
// and the results aggregated into StdOut and JSON formatting.
//
// Debug logging can be enabled if desired to assist with
// writing Rules. Additionall the default StdOut text can be
// disabled if using the JSON output for programmatic control
// over the results.
func (d *Deployment) Scan() {
	d.Logger = logrus.New()
	d.debugLogger("Begin resource scanning...")
	plan := tfresources.Plan{
		PlanFile:        d.PlanFile,
		ModulesFilePath: d.ModulesFile,
		Debug:           d.Debug,
		Logger:          d.Logger,
	}
	d.debugLogger(fmt.Sprintf("Parsing resources with plan file at path %s", plan.PlanFile))
	plan.GetResources()
	d.debugLogger(fmt.Sprintf("Parsed %v TF resources from provided plan", len(plan.Resources)))
	d.Resources = plan.Resources
	d.ruleEngine(d.Resources, d.Rules)
	d.stringResultFormatter()
	d.jsonResultFormatter()
	d.debugLogger(fmt.Sprintf("Finished rule evaluation for %v resources with %v results", len(plan.Resources), len(d.Results)))
	if !d.DisableStdOut {
		fmt.Println(d.ResultsStdOut)
	}
}
