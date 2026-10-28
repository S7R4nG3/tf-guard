![License: GPL v3](https://img.shields.io/badge/License-GPL_v3-blue.svg)
![latest build](https://github.com/S7R4nG3/tf-guard/actions/workflows/test.yml/badge.svg)
![latest release](https://img.shields.io/github/release-date/S7R4nG3/tf-guard)
[![Go Reference](https://pkg.go.dev/badge/github.com/S7R4nG3/tf-guard.svg)](https://pkg.go.dev/github.com/S7R4nG3/tf-guard)
# tf-guard

A package for writing Terraform Resource evaluation rules in native Golang.

A problem I've encountered with other Terraform security tools is the lack of control caused (most often) by the abstractions imposed by Domain Specific Languages (DSLs) used to make writing the rules themselves easier.

This causes headaches when trying to write more complicated logic around resources that DSLs may not have support for, or may cause conflicts with other resources during evaluation. Additionally the overhead of having to learn an additional DSL to support these tools can be cumbersome and time consuming when you already know full coding languages.

This package serves as my solution to that problem, allowing the flexibility of the full Golang language to allow users to write their own custom logic and rules for their resources within a provided Terraform plan.

This package intakes your Terraform plan (in JSON formatting) and your own defined resource rules (as functions extending the `Rule` type) and returns a simple ordered evaluation of those rules against the resources defined within your plan.

By removing the overhead of having to dig through the resources and their attributes in your plan files, you can instead focus on writing rules to evaluate your resources and take appropriate actions based on the results.

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

Additionally, if utilizing modules, a `ModulesFile` can be provided as the path to the deployment's `modules.json` file - written by `terraform init` to `.terraform/modules/modules.json` - which is used to link each resource to the module that declares it on `r.Module` (see [Modules](#modules)). _If not provided, a best effort is made to identify the file in the execution directory._

```go
d := Deployment{
    PlanFile: "./myplanfile.json",
    ModulesFile: "./.terraform/modules/modules.json",
    ...
}
```

Next, the functions to be executed against the resources must be provided. These functions must satisfy the `Rule` type and can be provided into the `Rules` parameter of the `Deployment`.

```go
d := Deployment{
    PlanFile: "./myplanfile.json",
    ModulesFile: "./.terraform/modules/modules.json",
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
        ModulesFile: "./.terraform/modules/modules.json",
        Rules: []Rule{
            MyCustomTRule
        },
    }
    ...
}

func MyCustomTFRule(r tfresources.Resource) tfGuard.Result {
    moduleReference := r.Module.ResolvedVersion   // Immutable versioning for the module linked to this resource
    attributeValues := r.Planned.AttributeValues  // map[string]interface{} Containing resource attribute values
    resourceType := r.Planned.Type                // A Terraform resource type - like aws_s3_bucket
    plannedActions := r.Change.Actions            // What Terraform intends to do - create, update, delete, ...
    unknownAttributes := r.KnownAfterApply        // Attributes Terraform can only resolve during the apply
    ...
    return tfGuard.Result {               // A Result struct must be returned for each Rule type
        Name: "This is my custom rule.",  // A simple name for this rule
        Valid: true,                      // A boolean Valid argument for rule validity
        RemediationMessage: "Do X to fix this.",  // An optional message detailing how to resolve this result
        ...                               // Other attributes can be defined if required - see Result struct
    }
}
```

### Modules

`r.Module` is the module that directly declares a resource. For a resource within a nested module this is the innermost module, and `r.Module.Parent` leads back through each of the module calls above it. Resources declared directly in the root module are left with a zero valued `Module`.

`Source` is the source exactly as it was written in the module's declaration, so a module declared as `source = "./policy"` inside a Git or registry module records only `./policy` - which doesn't say where its contents came from. `ResolvedSource` and `ResolvedVersion` resolve such local sources against their parent module, the same way Terraform installs them, and are the ones to reach for in rules about where module code comes from:

```go
func RuleModulesMustBeSourcedFromRegistry(r tfresources.Resource) tfGuard.Result {
    name := "Modules must be sourced from the Terraform Registry."
    if r.Module.Key == "" {
        return tfGuard.Result{NotApplicable: true}  // Declared in the root module
    }
    return tfGuard.Result{
        Name:     name,
        Valid:    strings.Contains(r.Module.ResolvedSource, "registry.terraform.io"),
        Severity: tfGuard.Severity.Critical,
    }
}
```

`Root()` returns the top-level module - the module call in your own configuration that led to the resource - which suits rules about your configuration's module calls rather than about where the code comes from:

```go
r.Planned.Address              // module.app["blue"].module.policy.terraform_data.policy
r.Module.Key                   // app.policy
r.Module.Source                // ../policy
r.Module.ResolvedSource        // ./modules/policy
r.Module.Root().Key            // app

for m := r.Module.Parent; m != nil; m = m.Parent {
    fmt.Println(m.Key, m.ResolvedSource)  // app ./modules/app
}
```

The StdOut results show a resource's module by its `ResolvedSource` and `ResolvedVersion`, adding the `Declared Source` whenever the two differ:

```sh
[ ❌ ]  Rule: Modules must be sourced from the Terraform Registry.
	Valid: false
	Severity: Critical
	Resource: module.remote.module.sibling.terraform_data.sibling
	Type: terraform_data
	      ⌙ Module: remote.sibling
		Source: git::https://example.com/org/terraform-modules.git//modules/sibling?ref=v1.0.0
		Declared Source: ../sibling
```

See the terraform-resources [Nested Modules](https://github.com/S7R4nG3/terraform-resources#nested-modules) documentation for how each kind of source is resolved.

> **Upgrading from terraform-resources v0.4.x:** resources within nested modules, and data sources within any module, were previously left with a zero valued `Module`. They are now linked to the module that declares them, so rules that skip resources without a module will start evaluating them - which can change your overall score - and rules checking `Source` should move to `ResolvedSource`.

### Known After Apply

Terraform can only resolve some attribute values while it is applying a plan, and renders those attributes as `(known after apply)`. Those values are **absent** from `Planned.AttributeValues`, which makes an attribute that Terraform will compute during the apply indistinguishable from one that was never configured at all. A rule that only reads the planned values will therefore report a resource as invalid for a setting that is in fact going to be applied.

Every resource carries the change Terraform plans to make to it on `r.Change`, along with a flattened list of the attributes that will only be known once that change has been applied on `r.KnownAfterApply`.

```go
r.Change.Actions              // [create], [update], [delete replace], ...
r.KnownAfterApply.Paths()     // [arn id statement[0].resources[0] tags_all]
```

`KnownAfterApply` offers two lookups. `Has` matches an attribute address exactly, while `HasPrefix` also matches anything nested beneath it - which is the one to reach for when Terraform may report a block as unknown either in whole or in part. Terraform flags a nested block as unknown at its root when the entire block is computed, but flags only the individual unknown values within it when the block itself is configured.

The three states a rule needs to tell apart are the attribute being configured, the attribute being deferred to the apply, and the attribute genuinely being missing:

```go
func RuleS3BucketsMustBeEncrypted(r tfresources.Resource) tfGuard.Result {
    name := "S3 buckets must have server-side encryption configured."
    encryption := "server_side_encryption_configuration"
    if r.Planned.Type != "aws_s3_bucket" {
        return tfGuard.Result{NotApplicable: true}
    }
    if _, configured := r.Planned.AttributeValues[encryption]; configured {
        return tfGuard.Result{Name: name, Valid: true, Severity: tfGuard.Severity.Critical}
    }
    if r.KnownAfterApply.HasPrefix(encryption) {
        // Terraform cannot resolve this value until the apply, so the rule
        // cannot fail the resource on it.
        return tfGuard.Result{
            Name:               name,
            Valid:              true,
            Severity:           tfGuard.Severity.Critical,
            RemediationMessage: "Encryption is known after apply - confirm it is enabled once the plan has been applied.",
        }
    }
    return tfGuard.Result{
        Name:               name,
        Valid:              false,
        Severity:           tfGuard.Severity.Critical,
        RemediationMessage: "Add a server_side_encryption_configuration block for this bucket.",
    }
}
```

A `Result` is either valid or invalid, so the deferred case has to be reported as one of the two. Passing it and carrying the deferral in the `RemediationMessage` - as above - keeps the resource visible in the results, while flagging it `NotApplicable` leaves it out of the overall score entirely. Which of the two suits your pipeline depends on whether an unresolvable setting should count as a pass or shouldn't be scored at all.

Attributes are addressed relative to the resource they belong to. Nested attributes are joined with a `.` and elements of a list or set are addressed by their index, producing paths such as `arn`, `tags_all`, or `statement[0].resources[1]`. Each entry also keeps that address broken into its individual `Steps` - a `string` for every object attribute or map key, an `int` for every collection index - for cases where a map key contains a `.` or a `[` and the rendered path would be ambiguous.

```go
for _, attr := range r.KnownAfterApply {
    fmt.Println(attr.Path)   // statement[0].resources[1]
    fmt.Println(attr.Steps)  // [statement 0 resources 1]
}
```

The `Change`, `KnownAfterApply`, and `UnknownAttribute` types are re-exported by this package, so rules and any helpers you factor out of them can be written using `tfGuard` alone:

```go
func deferredAttributes(unknowns tfGuard.KnownAfterApply) []string { ... }
```

The full change is available on `r.Change` as a [terraform-json](https://github.com/hashicorp/terraform-json) `Change`, so the prior state (`Before`), the planned state (`After`), the sensitive value masks, and the attributes forcing a replacement (`ReplacePaths`) are all reachable as well. Resources that the plan declares no change for are left with a zero valued `Change` and an empty `KnownAfterApply`.

Both examples include a rule validating a setting flagged as known after apply - see [examples/simple](./examples/simple/main.go) for a bucket whose encryption is deferred, and [examples/complex](./examples/complex/rules.go) for an IAM policy document rendered during the apply.

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

By default the StdOut results only detail the rules that were evaluated as **invalid**, keeping the output focused on the resources that need attention. The overall score is always reported at the very end and is calculated across _all_ results, not just the ones displayed.

```sh
TF Guard Rule Evaluation for Resources:

[ ❌ ]  Rule: All resources must include an Owner tag.
	Valid: false
	Severity: Major
	Resource: aws_s3_bucket.default
	Type: aws_s3_bucket

[ ❌ ]  Rule: All resources must include an Owner tag.
	Valid: false
	Severity: Major
	Resource: aws_s3_object.obj
	Type: aws_s3_object

Overall Resource Score: 50%
```

### Remediation Messages

Any Result can carry an optional `RemediationMessage` detailing how to resolve the finding. When provided it is appended to the result's StdOut entry, and it is always present in the JSON output - as an empty string when it isn't set.

```go
return tfGuard.Result{
    Name:               "All resources must include an Owner tag.",
    Valid:              false,
    Severity:           tfGuard.Severity.Major,
    RemediationMessage: "Add an `owner` tag to this resource identifying the owning team.",
}
```

```sh
[ ❌ ]  Rule: All resources must include an Owner tag.
	Valid: false
	Severity: Major
	Resource: aws_s3_bucket.default
	Type: aws_s3_bucket
	Remediation: Add an `owner` tag to this resource identifying the owning team.
```

### Output Toggles

Two toggles are available on the `Deployment` to control this output:

| Toggle | Default | Description |
| --- | --- | --- |
| `VerboseStdOut` | `false` | Include _every_ rule result in the output, both valid and invalid, rather than only the invalid ones. |
| `DisableStdOut` | `false` | Suppress writing the results to StdOut entirely - useful when consuming `ResultsJson` programmatically. The `ResultsStdOut` string is still populated. |

```go
d := Deployment{
    PlanFile: "./myplanfile.json",
    Rules: []Rule{
        ...
    },
    VerboseStdOut: true,
}
d.Scan()
```

With `VerboseStdOut` enabled, valid results are included alongside the invalid ones:

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

### JSON Output

The `ResultsJson` output contains the same results grouped several different ways so they can be parsed however best suits your automation.

| Key | Description |
| --- | --- |
| `ByResource` | All results keyed by the address of the resource they were evaluated against. |
| `ByRule` | All results keyed by the name of the rule that produced them. |
| `ByValid` | Every result that evaluated as **valid**, as a flat list. |
| `ByInvalid` | Every result that evaluated as **invalid**, as a flat list. |
| `TotalResults` | The total count of all results, valid and invalid. |
| `ValidResults` | The count of results that evaluated as valid. |
| `Score` | The percentage of results that evaluated as valid. |

The `ByValid` and `ByInvalid` groupings let you act on passing or failing rules directly without having to filter the resource or rule groupings yourself - for example, gating a pipeline on the failures alone.

```go
d.Scan()

var results struct {
    ByInvalid []tfGuard.Result
    Score     float64
}
json.Unmarshal(d.ResultsJson, &results)

for _, res := range results.ByInvalid {
    fmt.Printf("%s failed %s\n", res.Resource.Planned.Address, res.Name)
}
```

Both keys are always present in the output - an empty list when no results fall into that grouping. See [resultsJson.json](./examples/simple/resultsJson.json) for a full example response.

Every result embeds the resource it was evaluated against, so each one also carries that resource's `Change` and `KnownAfterApply` alongside its `Module` and `Planned` values. A resource declared within a module also carries that module's `ResolvedSource`, `ResolvedVersion`, and its full chain of `Parent` modules. The planned change includes the resource's full before and after states, which makes the output considerably larger than the results alone - decode only the keys you need rather than the whole document if that matters to your automation.

A `Deployment` can legitimately finish with _no_ results at all - if every rule is flagged `NotApplicable` against every resource in the plan, for instance, which is common when a plan contains only resource types none of your rules target. In that case the groupings are empty, `TotalResults` is `0`, and the `Score` is reported as `100`. Check `TotalResults` rather than `Score` alone if you need to distinguish "everything passed" from "nothing was evaluated".

Check out the [examples](./examples) on how you can integrate this package into your own codebase.

## Author

This codebase is created and maintained by [Dave Streng](https://www.linkedin.com/in/dave-streng).

## License

GNU General Public License v3.0 or later

See [LICENSE](./LICENSE) to see the full text.