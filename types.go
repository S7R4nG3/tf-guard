package tfGuard

import (
	tfresources "github.com/S7R4nG3/terraform-resources"
	"github.com/sirupsen/logrus"
)

// Some default values for Severities that can be easily
// leveraged if desired, if not, you can provide your
// own severity values/structures based on your needs.
var (
	Severity = struct {
		Minor    string
		Major    string
		Critical string
	}{
		Minor:    "Minor",
		Major:    "Major",
		Critical: "Critical",
	}
)

// A Rule function definition - this function type must be
// satisfied in order for the provided function to be executed
// against the planned resources.
//
// The input is a Resource type that extends the terraform-json
// project's StateResource struct by linking any resources
// created from modules back to their parent and children
// addresses. The output is a Result struct as defined below.
//
// Check out the [terraform-json](https://github.com/hashicorp/terraform-json/blob/main/state.go#L124) project for more
// details on how you can access different resource values.
//
// Additionally check out the [terraform-resources](https://pkg.go.dev/github.com/S7R4nG3/terraform-resources) project for more
// details on how you can utilize Module attributes to identify
// the exact source where a resources is defined.
type Rule func(tfresources.Resource) Result

// A Result is a data structure used to hold the evaluation
// results for a Rule (as defined above). A Rule type function
// must return a valid Result.
type Result struct {
	// A simple name for this result. This name will be provided
	// in the aggregated results to identify the rule executed.
	Name string

	// A validity boolean - ths is the primary mechanism to report
	// a rule execution as successful or unsuccessful.
	Valid bool

	// A simple Severity toggle - defaults to `Minor` and is fully
	// configurable if you have different Severity constructs.
	Severity string

	// The underlying resource attributes this result was executed
	// against. This is automatically populated at execution to
	// allow you to identify the resource being evaluated.
	Resource tfresources.Resource

	// A simple Not Applicable boolean toggle, used to flag
	// resources that a particular rule execution does not apply
	// to. Any results with this toggled to `true` will be ignored
	// in the overall results aggregation.
	NotApplicable bool
}

// A Deployment contains all the relevant attributes for a
// given Terraform plan execution. It contains references to
// Plan and Modules files, as well as the list of Rules to be
// evaluated against each resource, as lastly the results of
// the evaluation both in string-based StdOut format and JSON
// (to alloow for programmatic parsing).
type Deployment struct {
	// The filesystem path to your JSON formatting terraform plan
	// file.
	PlanFile string

	// The filesystem path to your JSON modules file. If not provided
	// a best effort is made to locate the file in its appropriate
	// directory within the execution directory.
	ModulesFile string

	// A Resource slice holding all resources parsed out of the
	// provided Terraform plan file, used to execute the rules.
	Resources []tfresources.Resource

	// A Rule slice used to contain all Rule type functions that
	// are to be executed against ALL resources within the Terraform
	// plan file.
	Rules []Rule

	// A Result slice containing all compiled results of the
	// execution of each Rule against each Resource.
	Results []Result

	// A simple string output of the results of the execution
	// outputted to StdOut by default but also accessible to
	// write to a file if desired.
	ResultsStdOut string

	// An organized JSON representation of the results of the
	// execution, organized by Resource and Rule and also providing
	// the overall score that can be used to programmatically take
	// actions based on results.
	ResultsJson []byte

	// A simple toggle allowing the default StdOut results output
	// to be disabled if the results are being assessed via other
	// means.
	DisableStdOut bool

	// A simple Debug logging toggle for troubleshooting.
	Debug bool

	// A Logger container used to output the debug logging (if
	// enabled).
	Logger *logrus.Logger
}

// An internal data type used to collect the results of the
// rule execution into JSON formatting.
type jsonResponse struct {
	ByResource   map[string][]Result
	ByRule       map[string][]Result
	TotalResults int
	ValidResults int
	Score        float64
}
