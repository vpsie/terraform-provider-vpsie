data "vpsie_protected_servers" "paused" {
  state = "paused"
}

output "paused_servers" {
  value = [for s in data.vpsie_protected_servers.paused.servers : s.name]
}
