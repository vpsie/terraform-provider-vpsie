terraform {
  required_providers {
    vpsie = {
      source = "registry.terraform.local/hashicorp/vpsie"
    }
  }
}

provider "vpsie" {
}



resource "vpsie_image" "image" {
  fetched_from_url = "https://releases.ubuntu.com/22.04.3/ubuntu-22.04.3-desktop-amd64.iso?_ga=2.199773230.1401239764.1697809661-306526328.1697639674"
  image_label      = "ubuntu-22.04.3"
  dc_identifier    = "af876af7-f295-11e4-b45d-005056aadd24"
}

output "image_test" {
  value = vpsie_image.image
}
