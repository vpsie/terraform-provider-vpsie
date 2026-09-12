output "available_datacenters" {
  description = "Datacenter identifiers available to this account."
  value       = [for dc in data.vpsie_datacenters.all.datacenters : dc.identifier]
}

output "server_group_identifier" {
  description = "Identifier of the created server group."
  value       = vpsie_server_group.app.identifier
}

output "managed_database_identifier" {
  description = "Identifier of the managed database cluster."
  value       = vpsie_managed_database.app.identifier
}

output "managed_database_admin_password" {
  description = "Generated admin password for the managed database."
  value       = vpsie_managed_database.app.admin_password
  sensitive   = true
}

output "registry_identifier" {
  description = "Identifier of the container registry."
  value       = vpsie_registry.app.identifier
}
