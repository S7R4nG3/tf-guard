resource "terraform_data" "outer" { input = "outer" }
module "inner" { source = "./inner" }
module "sibling" { source = "../sibling" }
