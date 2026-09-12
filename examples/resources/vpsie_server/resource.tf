resource "vpsie_server" "example" {
  hostname            = "my-server"
  dc_identifier       = "dc-identifier"
  os_identifier       = "os-identifier"
  resource_identifier = "resource-identifier"
  project_identifier  = "project-uuid-identifier"
  password            = "secure-password"
  add_public_ip_v4    = 1
  backup_enabled      = 0
  delete_reason       = "no longer needed"
}
