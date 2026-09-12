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

# A group to hold the environment's servers.
resource "vpsie_server_group" "app" {
  group_name        = "${var.environment}-app"
  group_description = "Application tier for the ${var.environment} environment"
  is_distributed    = true
}

# A managed database cluster (size/engine are selected by the offer).
resource "vpsie_managed_database" "app" {
  cluster_name          = "${var.environment}-db"
  datacenter_identifier = var.datacenter_identifier
  resource_identifier   = var.database_resource_identifier
  vpc_id                = var.database_vpc_id
  project_identifier    = var.project_identifier
  node_count            = var.database_node_count
}

# A private container registry for the environment's images.
resource "vpsie_registry" "app" {
  name                  = "${var.environment}-registry"
  datacenter_identifier = var.datacenter_identifier
  plan_identifier       = var.registry_plan_identifier
  project_identifier    = var.project_identifier
}

# Apply environment tags to the registry.
resource "vpsie_tag" "registry" {
  entity              = "container_registry"
  resource_identifier = vpsie_registry.app.identifier
  tags                = ["${var.environment}-terraform"]
}

# A TLS certificate issued for a domain you own in VPSie DNS.
resource "vpsie_certificate" "app" {
  count = var.certificate_domain_id == "" ? 0 : 1

  cert_name = "${var.environment}-cert"
  domain_id = var.certificate_domain_id
}
