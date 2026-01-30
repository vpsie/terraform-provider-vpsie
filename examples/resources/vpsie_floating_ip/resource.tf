resource "vpsie_floating_ip" "example" {
  vm_identifier = "vm-identifier"
  dc_identifier = "dc-identifier"
  ip_type       = "IPv4"
}
