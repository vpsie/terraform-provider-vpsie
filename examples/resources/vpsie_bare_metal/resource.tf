variable "bare_metal_password" {
  type      = string
  sensitive = true
}

data "vpsie_bare_metal_plans" "gpu" {
  dc_identifier = "dc-identifier"
  gpu           = true
  min_ram_gb    = 128
}

data "vpsie_bare_metal_os" "ubuntu" {
  dc_identifier   = "dc-identifier"
  plan_identifier = data.vpsie_bare_metal_plans.gpu.plans[0].identifier
}

resource "vpsie_bare_metal" "example" {
  dc_identifier   = "dc-identifier"
  plan_identifier = data.vpsie_bare_metal_plans.gpu.plans[0].identifier
  os_identifier = one([
    for os in data.vpsie_bare_metal_os.ubuntu.operating_systems : os.identifier
    if os.distribution == "Ubuntu" && os.is_default
  ])
  hostname            = "gpu-01.example.com"
  password            = var.bare_metal_password
  ssh_key_identifiers = ["ssh-key-identifier"]
  contract_term       = "monthly"
  private_network_ids = [vpsie_vpc.example.id]

  remote_access = {
    enabled             = true
    firewall_identifier = "firewall-group-identifier"
  }

  user_data = <<-EOT
    #cloud-config
    packages:
      - nvidia-utils-550-server
  EOT

  release_reason = "Replaced by Terraform"

  timeouts = {
    create = "90m"
    delete = "45m"
  }
}
