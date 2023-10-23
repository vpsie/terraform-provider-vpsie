# terraform {
#   required_providers {
#     vpsie = {
#         source = "registry.terraform.local/hashicorp/vpsie"
#     }
#   }
# }

# provider "vpsie" {
#   access_token = "REDACTED-ROTATED-TOKEN"
# }

# data "vpsie_storages" "all" {}

# output "storages" {
#     value = data.vpsie_storages.all
# }