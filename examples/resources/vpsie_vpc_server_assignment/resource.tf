resource "vpsie_vpc_server_assignment" "example" {
  vm_identifier = "vm-identifier"
  vpc_id        = 1
  dc_identifier = "dc-identifier"
}
