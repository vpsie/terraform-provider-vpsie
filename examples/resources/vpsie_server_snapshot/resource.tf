resource "vpsie_server_snapshot" "example" {
  name          = "my-server-snapshot"
  vm_identifier = "vm-identifier"
  note          = "Snapshot before upgrade"
}
