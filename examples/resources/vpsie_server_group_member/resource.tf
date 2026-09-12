resource "vpsie_server_group_member" "example" {
  group_identifier = vpsie_server_group.example.identifier
  vm_identifier    = "vm-identifier"
}
