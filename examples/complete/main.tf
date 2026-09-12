# A complete, end-to-end example that provisions a small, tagged environment:
# a server group, a managed database, a container registry and a certificate.
#
#   export VPSIE_ACCESS_TOKEN="your-api-token"
#   terraform init
#   terraform apply
terraform {
  required_providers {
    vpsie = {
      source = "vpsie/vpsie"
    }
  }
}

provider "vpsie" {
  # Reads VPSIE_ACCESS_TOKEN from the environment when access_token is omitted.
  access_token = var.vpsie_access_token
}

# Look up available datacenters so you can pick a valid identifier.
data "vpsie_datacenters" "all" {}

# A reusable label applied to resources in this environment.
resource "vpsie_tag" "env" {
  name  = "${var.environment}-terraform"
  color = "#2563eb"
}

# A group to hold the environment's servers.
resource "vpsie_server_group" "app" {
  group_name        = "${var.environment}-app"
  group_description = "Application tier for the ${var.environment} environment"
  is_distributed    = true
}

# A managed MySQL database cluster.
resource "vpsie_managed_database" "app" {
  name                  = "${var.environment}-db"
  db_type               = "mysql"
  datacenter_identifier = var.datacenter_identifier
  plan_id               = var.database_plan_id
  node_count            = var.database_node_count
}

# A private container registry for the environment's images.
resource "vpsie_registry" "app" {
  name                  = "${var.environment}-registry"
  datacenter_identifier = var.datacenter_identifier
  plan_identifier       = var.registry_plan_identifier
}

# A TLS certificate issued for a domain you own in VPSie DNS.
resource "vpsie_certificate" "app" {
  count = var.certificate_domain_id == "" ? 0 : 1

  cert_name = "${var.environment}-cert"
  domain_id = var.certificate_domain_id
}
