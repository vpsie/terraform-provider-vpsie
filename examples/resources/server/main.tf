terraform {
  required_providers {
    vpsie = {
        source = "registry.terraform.local/hashicorp/vpsie"
    }
  }
}

provider "vpsie" {
}

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
