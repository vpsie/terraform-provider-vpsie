resource "vpsie_tag" "example" {
  entity              = "ssh_keys"
  resource_identifier = "ssh-key-identifier"
  tags                = ["production", "web"]
}
