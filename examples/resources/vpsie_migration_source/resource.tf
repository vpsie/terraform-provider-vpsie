resource "vpsie_migration_source" "esxi" {
  type                  = "vmware-esxi"
  name                  = "lab-esxi"
  datacenter_identifier = "dc-identifier"

  settings = {
    host     = "10.0.0.5"
    username = "root"
  }

  credentials = {
    password = var.esxi_password
  }
}
