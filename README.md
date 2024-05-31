![License: GPL v3](https://img.shields.io/badge/License-GPL_v3-blue.svg)
![latest build](https://github.com/S7R4nG3/terraform-resources/actions/workflows/test.yml/badge.svg)
![latest release](https://img.shields.io/github/release-date/S7R4nG3/tf-guard)
[![Go Reference](https://pkg.go.dev/badge/github.com/S7R4nG3/tf-guard.svg)](https://pkg.go.dev/github.com/S7R4nG3/tf-guard)
# tf-guard

A package for writing Terraform Resource evaluation rules in native Golang.

A problem I've encountered with other Terraform security tools is the lack of control caused (most often) by the abstractions imposed by Domain Specific Languages (DSLs) used to make writing the rules themselves easier.

This causes headaches when trying to write more complicated logic around resources that DSLs may not have support for, or may cause conflicts with other resources during evaluation. Additionally the overhead of having to learn an additional DSL to support these tools can be cumbersome and time consuming when you already know full coding languages.

This package serves as my solution to that problem, allowing the flexibility of the full Golang language to allow users to write their own custom logic and rules for their resources within a provided Terraform plan.

## Usage

The package requires that a JSON formatted Terraform plan file be collected containing the resources that are to be evaluated. This can be accomplished with 2 Terraform commands:

```
terraform plan -out=plan.tfplan
terraform show -json plan.tfplan > myplanfile.json
```

This JSON formatted plan file is fed into the package via the `Deployment` struct to parse its contents.

```go
d := Deployment{
    PlanFile: "./myplanfile.json",
    ...
}
```

Additionally, if utilizing modules, a `ModulesFile` can be provided as the path to the modules.json file for the deployment that is used to link the underlying resources being created from within a module, to its parent and children addresses. _If not provided, a best effort is made to identify the file in the execution directory._

```go
d := Deployment{
    PlanFile: "./myplanfile.json",
    ModulesFile: "./.terraform/modules.json",
    ...
}
```

Next, the functions to be executed against the resources must be provided. These functions must satisfy the `Rule` type and can be provided into the `Rules` parameter of the `Deployment`.

```go
d := Deployment{
    PlanFile: "./myplanfile.json",
    ModulesFile: "./.terraform/modules.json",
    Rules: []Rule{
        ...
    },
    ...
}
```

By providing a simple function type, the package allows custom functions to be written with full access to each Terraform resource's configuration values.

```go
type Rule func(tfresources.Resource) Result
```

An example Rule declaration to showing access to Terraform module attributes as well as resource values:

```go
func main(){
    d := Deployment{
        PlanFile: "./myplanfile.json",
        ModulesFile: "./.terraform/modules.json",
        Rules: []Rule{
            MyCustomTRule
        },
    }
    ...
}

func MyCustomTFRule(r tfresources.Resource) tfGuard.Result {
    moduleReference := r.Module.Version           // Immutable versioning for the module linked to this resource
    attributeValues := r.Planned.AttributeValues  // map[string]interface{} Containing resource attribute values
    resourceType := r.Planned.Type                // A Terraform resource type - like aws_s3_bucket
    ...
    return tfGuard.Result {               // A Result struct must be returned for each Rule type
        Name: "This is my custom rule.",  // A simple name for this rule
        Valid: true,                      // A boolean Valid argument for rule validity
        ...                               // Other attributes can be defined if required - see Result struct
    }
}
```

Finally the `Scan` method can be executed to parse the resources, evaluate each rule against them, and return the results as StdOut string text, or in JSON formatting for programmatic parsing.

```go
func main(){
    d := Deployment{
        ...
    }
    d.Scan()                    // Results are written to StdOut unless Disabled

    fmt.Println(d.ResultsJson)  // JSON output can be written to a file if desired, or parsed directly
}
```

```sh
TF Guard Rule Evaluation for Resources:

[ ✅ ]  Rule: S3 buckets must be tagged at all times.
	Valid: true
	Severity: Minor
	Resource: aws_s3_bucket.default
	Type: aws_s3_bucket

[ ❌ ]  Rule: All resources must include an Owner tag.
	Valid: false
	Severity: Major
	Resource: aws_s3_bucket.default
	Type: aws_s3_bucket

[ ✅ ]  Rule: S3 Bucket Objects must always be tagged.
	Valid: true
	Severity: Minor
	Resource: aws_s3_object.obj
	Type: aws_s3_object

[ ❌ ]  Rule: All resources must include an Owner tag.
	Valid: false
	Severity: Major
	Resource: aws_s3_object.obj
	Type: aws_s3_object

Overall Resource Score: 50%
```

Check out the [examples](./examples) on how you can integrate this package into your own codebase.

## Author

This codebase is created and maintained by [Dave Streng](https://www.linkedin.com/in/dave-streng).

## License

GNU General Public License v3.0 or later

See LICENSE to see the full text.