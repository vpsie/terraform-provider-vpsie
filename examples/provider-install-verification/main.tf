terraform {
    required_providers {
        vpsie = {
            source = "registry.terraform.local/hashicorp/vpsie"
        }
    }
}

provider "vpsie" {}

data "vpsie_storages" "storage" {}