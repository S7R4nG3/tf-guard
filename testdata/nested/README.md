# nested

Copied from the [terraform-resources](https://github.com/S7R4nG3/terraform-resources/tree/main/testdata/nested) `testdata/nested` fixture.

Local and Git sourced modules nested within one another, with `count` and `for_each` instances and a deferred data source. The deployment uses only Terraform's built-in `terraform_data` resource and `terraform_remote_state` data source, so no provider credentials are needed.

`modules.json` and `plan.json` were generated with Terraform v1.15.9:

```sh
# The deferred data source reads an empty state file, which .gitignore excludes
echo '{"version":4,"terraform_version":"1.15.9","serial":1,"lineage":"fixture","outputs":{},"resources":[]}' > modules/app/bucket/empty.tfstate
terraform init
terraform plan -out=plan.tfplan
terraform show -json plan.tfplan > plan.json
cp .terraform/modules/modules.json .
```

The `remote` module was installed from a local Git repository containing the contents of `remote-repo/`, tagged `v1.0.0`. In the generated files, its `git::file://` path was then replaced with `git::https://example.com/org/terraform-modules.git` so they don't depend on a local path. To regenerate, point `main.tf` at a local clone in the same way.
