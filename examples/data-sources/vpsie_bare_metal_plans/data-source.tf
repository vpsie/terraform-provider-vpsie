data "vpsie_bare_metal_plans" "example" {
  dc_identifier = "dc-identifier"
  min_ram_gb    = 64
  min_disk_gb   = 1000
  gpu           = false
}

output "in_stock_plans" {
  value = [for p in data.vpsie_bare_metal_plans.example.plans : p.name if p.available_servers > 0]
}
