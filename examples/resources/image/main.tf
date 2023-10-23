terraform {
  required_providers {
    vpsie = {
        source = "registry.terraform.local/hashicorp/vpsie"
    }
  }
}

provider "vpsie" {
}


# resource "vpsie_image" "image" {
#   fetched_from_url = "https://releases.ubuntu.com/22.04.3/ubuntu-22.04.3-desktop-amd64.iso?_ga=2.199773230.1401239764.1697809661-306526328.1697639674"
#   image_label = "ubuntu-22.04.3"
#   dc_identifier = "af876af7-f295-11e4-b45d-005056aadd24"
# }

# output "image_test" {
#     value = vpsie_image.image
# }



# resource "vpsie_script" "script" {
#   script = "echo hello"
#   script_name = "hello"
#   type = "bash"
# }

# output "script_test" {
#     value = vpsie_script.script
# }

resource "vpsie_server" "server" {
  project_id = "71128"
  hostname = "test-hostIP"
  dc_identifier = "55f06b85-c9ee-11e3-9845-005056aa8af7"
  os_identifier = "17e21ee5-37df-11ee-8bba-0050569c68dc"
  resource_identifier = "f161b479-c469-11eb-8ef8-0050569c68dc"
  password = "ExeMrq@y@6azxYQ"
  delete_reason = "no longer need"
}

output "server_test" {
    sensitive = true
    value = vpsie_server.server
}