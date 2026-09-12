resource "vpsie_managed_database" "example" {
  cluster_name          = "my-db-cluster"
  datacenter_identifier = "datacenter-uuid-identifier"
  resource_identifier   = "offer-uuid-identifier" # from the managed database offers
  vpc_id                = 1234                    # numeric VPC id
  project_identifier    = "project-uuid-identifier"
  node_count            = 1
}
