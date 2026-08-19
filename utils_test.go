package tfGuard

import (
	"strings"
	"testing"

	tfresources "github.com/S7R4nG3/terraform-resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/sirupsen/logrus"
)

func TestStringResultFormatter(t *testing.T) {
	testResults := []Result{
		{
			Name:     "passing rule",
			Valid:    true,
			Severity: Severity.Minor,
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_bucket.default",
					Type:    "aws_s3_bucket",
				},
			},
		},
		{
			Name:     "failing rule",
			Valid:    false,
			Severity: Severity.Major,
			Resource: tfresources.Resource{
				Planned: tfjson.StateResource{
					Address: "aws_s3_object.obj",
					Type:    "aws_s3_object",
				},
			},
		},
	}
	tests := []struct {
		name       string
		deployment Deployment
		want       []string
		notWant    []string
	}{
		{
			name: "Default output should only detail invalid rules.",
			deployment: Deployment{
				Results: testResults,
			},
			want:    []string{"failing rule", "Overall Resource Score: 50%"},
			notWant: []string{"passing rule"},
		},
		{
			name: "Verbose output should detail all rules.",
			deployment: Deployment{
				Results:       testResults,
				VerboseStdOut: true,
			},
			want: []string{"passing rule", "failing rule", "Overall Resource Score: 50%"},
		},
		{
			name: "All valid results should only output the score.",
			deployment: Deployment{
				Results: testResults[:1],
			},
			want:    []string{"Overall Resource Score: 100%"},
			notWant: []string{"passing rule"},
		},
	}

	for _, tt := range tests {
		t.Logf("Running test -- %v", tt.name)
		d := tt.deployment
		d.Debug = true
		d.Logger = logrus.New()
		d.stringResultFormatter()
		got := d.ResultsStdOut
		for _, want := range tt.want {
			if !strings.Contains(got, want) {
				t.Errorf("Test Error -- %s:\ngot: %v\n::\nexpected to contain: %v\n", tt.name, got, want)
			}
		}
		for _, notWant := range tt.notWant {
			if strings.Contains(got, notWant) {
				t.Errorf("Test Error -- %s:\ngot: %v\n::\nexpected NOT to contain: %v\n", tt.name, got, notWant)
			}
		}
	}
}
