resource "terraform_data" "app" { input = "app" }
module "bucket" {
  source = "./bucket"
  count  = 2
}
module "policy" { source = "../policy" }
