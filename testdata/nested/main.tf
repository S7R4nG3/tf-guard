resource "terraform_data" "root" { input = "root" }
module "app" {
  source   = "./modules/app"
  for_each = toset(["blue", "green"])
}
module "remote" {
  source = "git::https://example.com/org/terraform-modules.git//modules/outer?ref=v1.0.0"
}
