terraform {
  required_providers {
    vpsie = {
        source = "registry.terraform.local/hashicorp/vpsie"
    }
  }
}

provider "vpsie" {
}

data "vpsie_images" "all" {}

output "images" {
    value = data.vpsie_images.all
}