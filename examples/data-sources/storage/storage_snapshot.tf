terraform {
  required_providers {
    vpsie = {
      source = "registry.terraform.local/hashicorp/vpsie"
    }
  }
}

provider "vpsie" {
}

data "vpsie_storage_snapshots" "all" {}

output "storages" {
  value = data.vpsie_storage_snapshots.all
}