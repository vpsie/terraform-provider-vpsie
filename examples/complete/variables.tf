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

variable "database_plan_id" {
  description = "Managed database plan id."
  type        = number
  default     = 1
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
