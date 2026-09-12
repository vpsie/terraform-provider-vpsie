# Complete example

Provisions a small, tagged environment on VPSie: a server group, a managed
MySQL database, a container registry and (optionally) a TLS certificate.

## Usage

```sh
export VPSIE_ACCESS_TOKEN="your-api-token"

terraform init
terraform apply \
  -var 'datacenter_identifier=REPLACE_ME' \
  -var 'registry_plan_identifier=REPLACE_ME'
```

Run `terraform apply` with no extra vars first to see the
`available_datacenters` output, then pick a datacenter identifier.

The managed database admin password is exposed as a `sensitive` output; read it
with `terraform output -raw managed_database_admin_password`.
