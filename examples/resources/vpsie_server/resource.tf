resource "vpsie_server" "example" {
  hostname            = "my-server"
  dc_identifier       = "dc-identifier"
  os_identifier       = "os-identifier"
  resource_identifier = "resource-identifier"
  project_id          = 1
  password            = "secure-password"
  delete_reason       = "no longer needed"
}
