package tfGuard

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	tfresources "github.com/S7R4nG3/terraform-resources"
)

// The rule engine is used internally to orchestrate execution
// of each rule against the list of resources parsed out of the
// plan file.
//
// For speed, a goroutine is generated for each resource that
// executes all rules against the resource and later aggregates
// the results together into the overall Deployment.
//
// This allow for numerous complex rules to be evaluated in parallel
// against each resource so that large deployments do not bog the
// system or slow executions.
func (d *Deployment) ruleEngine(resources []tfresources.Resource, rules []Rule) {
	d.debugLogger("Begin Rule Engine execution...")
	wg := new(sync.WaitGroup)
	out := make(chan Result)
	for _, res := range resources {
		thisResource := res.Planned
		d.debugLogger(fmt.Sprintf("Starting rule execution for resource -- %s", thisResource.Address))
		wg.Add(1)
		go d.worker(rules, res, out, wg)
	}
	go wait(out, wg)
	results := []Result{}
	for o := range out {
		results = append(results, o)
	}
	d.debugLogger("Finished Rule Engine execution.")
	d.Results = results
}

func (d *Deployment) worker(rules []Rule, res tfresources.Resource, out chan<- Result, wg *sync.WaitGroup) {
	thisResource := res.Planned
	for _, rule := range rules {
		result := rule(res)
		if result.NotApplicable {
			fullName := strings.Split((runtime.FuncForPC(reflect.ValueOf(rule).Pointer()).Name()), ".")
			name := fullName[len(fullName)-1]
			d.debugLogger(fmt.Sprintf("Rule: **%s** execution for resource %s evaluated as NOT APPLICABLE", name, thisResource.Address))
		} else {
			result.Resource = res
			if result.Severity == "" {
				result.Severity = Severity.Minor
			}
			d.debugLogger(fmt.Sprintf("Rule: **%s** execution for resource %s evaluates as %v with severity %s", result.Name, thisResource.Address, result.Valid, result.Severity))
			out <- result
		}
	}
	wg.Done()
}

func wait(out chan<- Result, wg *sync.WaitGroup) {
	defer close(out)
	wg.Wait()
}
