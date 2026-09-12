terraform {
  required_providers {
    vpsie = {
      source = "vpsie/vpsie"
    }
  }
}

provider "vpsie" {
  # Authenticates using the VPSIE_ACCESS_TOKEN environment variable.
  # Alternatively set access_token here (not recommended — keep tokens out of code).
}

data "vpsie_storages" "all" {}

output "storages" {
  value = data.vpsie_storages.all
}
