resource "terraform_data" "bucket" { input = "bucket" }

// Depends on a value only known after apply, so Terraform defers the read and
// the data source is reported within the plan's planned values.
data "terraform_remote_state" "lookup" {
  backend  = "local"
  config   = { path = "${path.module}/empty.tfstate" }
  defaults = { bucket_id = terraform_data.bucket.id }
}
