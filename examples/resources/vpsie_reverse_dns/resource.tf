resource "vpsie_reverse_dns" "example" {
  vm_identifier     = "vm-identifier"
  ip                = "192.168.1.1"
  domain_identifier = "domain-identifier"
  hostname          = "server.example.com"
}
