package tfGuard

import (
	"encoding/json"
	"testing"

	tfresources "github.com/S7R4nG3/terraform-resources"
	"github.com/sirupsen/logrus"
)

func TestDeployment(t *testing.T) {
	testRules := []Rule{
		func(t tfresources.Resource) Result {
			if t.Planned.Type != "" {
				return Result{
					Name:     "Test Rule",
					Severity: Severity.Major,
					Valid:    true,
				}
			}
			return Result{}
		},
	}
	// Rules that apply to no resource in the null_resource plan, leaving
	// the Deployment with an empty result set.
	notApplicableRules := []Rule{
		func(t tfresources.Resource) Result {
			if t.Planned.Type == "aws_s3_bucket" {
				return Result{
					Name:     "Test Rule",
					Severity: Severity.Major,
					Valid:    true,
				}
			}
			return Result{NotApplicable: true}
		},
	}
	tests := []struct {
		name       string
		deployment Deployment
		want       float64
	}{
		{
			name: "Simple deployment should execute cleanly.",
			deployment: Deployment{
				PlanFile: "./testdata/simple/plan.json",
				Rules:    testRules,
				Debug:    true,
				Logger:   logrus.New(),
			},
			want: 100.00,
		},
		{
			name: "Complex deployment should execute cleanly.",
			deployment: Deployment{
				PlanFile: "./testdata/complex/plan.json",
				Rules:    testRules,
				Debug:    true,
				Logger:   logrus.New(),
			},
			want: 100.00,
		},
		{
			name: "A deployment producing no applicable results should still marshal cleanly.",
			deployment: Deployment{
				PlanFile: "./testdata/null/plan.json",
				Rules:    notApplicableRules,
				Debug:    true,
				Logger:   logrus.New(),
			},
			want: 100.00,
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %v", tt.name)
		d := tt.deployment
		d.Scan()
		if len(d.ResultsJson) == 0 {
			t.Errorf("Test Error -- %s:\nResultsJson is empty, the results failed to marshal\n", tt.name)
			continue
		}
		var got jsonResponse
		err := json.Unmarshal(d.ResultsJson, &got)
		if err != nil {
			t.Errorf("Error unmarshalling test JSON data -- %s", tt.name)
		}
		if got.Score != tt.want {
			t.Errorf("Error with test:\ngot: %v\n::\nwant: %v\n", got, tt.want)
		}
	}
}
