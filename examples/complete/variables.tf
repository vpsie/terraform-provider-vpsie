variable "vpsie_access_token" {
  description = "VPSie API access token. Prefer the VPSIE_ACCESS_TOKEN environment variable."
  type        = string
  default     = null
  sensitive   = true
}

variable "environment" {
  description = "Short environment name used as a prefix for resource names."
  type        = string
  default     = "dev"
}

variable "datacenter_identifier" {
  description = "Datacenter identifier to deploy into (see the vpsie_datacenters data source)."
  type        = string
}

variable "project_identifier" {
  description = "Project UUID identifier to create resources in."
  type        = string
}

variable "database_resource_identifier" {
  description = "Managed database offer (plan) identifier; see the datacenter's managed database offers."
  type        = string
}

variable "database_vpc_id" {
  description = "Numeric id of the VPC the managed database attaches to."
  type        = number
}

variable "database_node_count" {
  description = "Number of nodes in the managed database cluster."
  type        = number
  default     = 1
}

variable "registry_plan_identifier" {
  description = "Container registry resource plan identifier."
  type        = string
}

variable "certificate_domain_id" {
  description = "Domain identifier to issue a certificate for. Leave empty to skip the certificate."
  type        = string
  default     = ""
}
