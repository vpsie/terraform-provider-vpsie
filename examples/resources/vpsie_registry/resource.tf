resource "vpsie_registry" "example" {
  name                  = "my-registry"
  datacenter_identifier = "datacenter-uuid-identifier"
  plan_identifier       = "plan-uuid-identifier"
  project_identifier    = "project-uuid-identifier"
}
