package tfGuard

import (
	"encoding/json"

	tfresources "github.com/S7R4nG3/terraform-resources"
	"github.com/fatih/color"
)

func (d *Deployment) stringResultFormatter() {
	d.debugLogger("Aggregating string results...")
	body := color.New(color.FgCyan).Add(color.Bold).Sprint("\nTF Guard Rule Evaluation for Resources:")
	validCounter := 0
	for _, res := range d.Results {
		c := color.New(color.FgYellow)
		if res.Valid {
			c = color.New(color.FgGreen)
			validCounter++
			body += "\n\n[ ✅ ]"
		} else {
			body += "\n\n[ ❌ ]"
		}
		thisResource := res.Resource.Planned
		body += c.Sprintf("  Rule: %s\n\tValid: %v\n\tSeverity: %s\n\tResource: %s\n\tType: %s", res.Name, res.Valid, res.Severity, thisResource.Address, thisResource.Type)
		if (tfresources.Module{}) != res.Resource.Module {
			thisModule := res.Resource.Module
			body += c.Sprintf("\n\t      ⌙ Module: %s\n\t\tSource: %s", thisModule.Key, thisModule.Source)
			if thisModule.Version != "" {
				body += c.Sprintf("\n\t\tVersion: %v", thisModule.Version)
			}
		}
	}
	score := (float64(validCounter) / float64(len(d.Results))) * 100
	body += color.New(color.FgCyan).Add(color.Bold).Sprintf("\n\nOverall Resource Score: %.0f%%\n", score)
	d.ResultsStdOut = body
	d.debugLogger("String results aggregated.")
}

func (d *Deployment) jsonResultFormatter() {
	d.debugLogger("Aggregating JSON results...")
	j := jsonResponse{
		ByResource: make(map[string][]Result),
		ByRule:     make(map[string][]Result),
	}
	validCounter := 0
	for _, res := range d.Results {
		if res.Valid {
			validCounter++
		}
		thisResource := res.Resource.Planned
		if _, exists := j.ByResource[thisResource.Address]; !exists {
			j.ByResource[thisResource.Address] = []Result{res}
		} else {
			j.ByResource[thisResource.Address] = append(j.ByResource[thisResource.Address], res)
		}
	}

	for _, rule := range d.Results {
		if _, exists := j.ByRule[rule.Name]; !exists {
			j.ByRule[rule.Name] = []Result{rule}
		} else {
			j.ByRule[rule.Name] = append(j.ByRule[rule.Name], rule)
		}
	}
	j.ValidResults = validCounter
	j.TotalResults = len(d.Results)
	j.Score = (float64(validCounter) / float64(len(d.Results))) * 100
	out, err := json.Marshal(j)
	if err != nil {
		d.Logger.Error("error marshalling json body", err)
	}
	d.ResultsJson = out
	d.debugLogger("JSON results aggregated.")
}

func (d *Deployment) debugLogger(msg string) {
	if d.Debug {
		d.Logger.Info(msg)
	}
}
