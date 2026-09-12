resource "vpsie_managed_database" "example" {
  name                  = "app-db"
  db_type               = "mysql"
  datacenter_identifier = "datacenter-identifier"
  plan_id               = 1
  node_count            = 1
}
