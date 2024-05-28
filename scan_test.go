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
	}

	for _, tt := range tests {
		t.Logf("Running test -- %v", tt.name)
		d := tt.deployment
		d.Scan()
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
