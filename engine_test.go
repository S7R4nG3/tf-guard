package tfGuard

import (
	"reflect"
	"testing"

	tfresources "github.com/S7R4nG3/terraform-resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/sirupsen/logrus"
)

func TestRuleEngine(t *testing.T) {
	testResource := tfresources.Resource{
		Planned: tfjson.StateResource{
			Type: "aws_s3_bucket",
		},
	}
	tests := []struct {
		name      string
		resources []tfresources.Resource
		rules     []Rule
		want      []Result
	}{
		{
			name:      "Simple test should execute cleanly.",
			resources: []tfresources.Resource{testResource},
			rules: []Rule{
				func(t tfresources.Resource) Result {
					if _, exists := t.Planned.AttributeValues["tags"]; !exists {
						return Result{
							Name:  "test",
							Valid: false,
						}
					}
					return Result{}
				},
			},
			want: []Result{
				{
					Name:     "test",
					Severity: "Minor",
					Valid:    false,
					Resource: testResource,
				},
			},
		},
		{
			name:      "Providng no resource tests should execute cleanly.",
			resources: []tfresources.Resource{testResource},
			rules:     []Rule{},
			want:      []Result{},
		},
		{
			name:      "Provide no resources should execute cleanly.",
			resources: []tfresources.Resource{},
			rules: []Rule{
				func(t tfresources.Resource) Result {
					if _, exists := t.Planned.AttributeValues["tags"]; !exists {
						return Result{
							Name:  "test",
							Valid: false,
						}
					}
					return Result{}
				},
			},
			want: []Result{},
		},
		{
			name:      "Providing no resources or tests should execute cleanly.",
			resources: []tfresources.Resource{},
			rules:     []Rule{},
			want:      []Result{},
		},
	}

	for _, tt := range tests {
		t.Logf("Starting test - %s", tt.name)
		d := Deployment{
			Debug:  true,
			Logger: logrus.New(),
		}
		d.ruleEngine(tt.resources, tt.rules)
		got := d.Results
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Test Error:\ngot: %v\n::\nwant: %v\n", got, tt.want)
		}
	}
}
