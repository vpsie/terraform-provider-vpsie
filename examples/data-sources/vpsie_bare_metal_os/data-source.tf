data "vpsie_bare_metal_os" "example" {
  dc_identifier   = "dc-identifier"
  plan_identifier = "plan-identifier"
}

output "linux_images" {
  value = { for os in data.vpsie_bare_metal_os.example.operating_systems : os.identifier => os.name if os.os_family == "linux" }
}
