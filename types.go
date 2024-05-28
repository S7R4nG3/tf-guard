package tfGuard

import (
	tfresources "github.com/S7R4nG3/terraform-resources"
	"github.com/sirupsen/logrus"
)

// Some default values for Severities that can be easily
// leveraged if desired, if not, users can provide their
// own severity values/structures based on their own needs.
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
// satisfied in order for it to be executed against the
// planned resources.
//
// The input is a Resource type that extends the terraform-json
// project's StateResource struct by linking any resources
// created from modules back to their parent and children
// addresses. The output is a Result struct as defined below.
//
// Check out the [terraform-json](https://github.com/hashicorp/terraform-json/blob/main/state.go#L124) project for more
// details on how you can access different resource values.
type Rule func(tfresources.Resource) Result

// Result is a data structure used to hold the evaluation
// results for a Rule (as defined above). A Rule type function
// must return a valid Result.
type Result struct {
	Name          string
	Valid         bool
	Severity      string
	Resource      tfresources.Resource
	NotApplicable bool
}

// A Deployment contains all the relevant attributes for a
// given Terraform plan execution. It contains references to
// Plan and Modules files, as well as the list of Rules to be
// evaluated against each resource, as lastly the results of
// the evaluation both in string-based StdOut format and JSON
// (to alloow for programmatic parsing).
type Deployment struct {
	PlanFile      string
	ModulesFile   string
	Resources     []tfresources.Resource
	Rules         []Rule
	Results       []Result
	ResultsStdOut string
	ResultsJson   []byte
	DisableStdOut bool
	Debug         bool
	Logger        *logrus.Logger
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
