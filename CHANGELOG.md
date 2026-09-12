## 0.1.0 (Unreleased)

FEATURES:

- **New Resource:** `vpsie_tag`
- **New Resource:** `vpsie_certificate`
- **New Resource:** `vpsie_server_group`
- **New Resource:** `vpsie_server_group_member`
- **New Resource:** `vpsie_registry`
- **New Resource:** `vpsie_managed_database`
- **New Data Source:** `vpsie_tags`
- **New Data Source:** `vpsie_certificates`
- **New Data Source:** `vpsie_server_groups`
- **New Data Source:** `vpsie_registries`
- **New Data Source:** `vpsie_managed_databases`

ENHANCEMENTS:

- Added an `endpoint` provider argument (and `VPSIE_ENDPOINT`) to target
  non-production API environments.

- Updated the HashiCorp Terraform Plugin Framework and companion modules to their
  current releases (`terraform-plugin-framework` v1.19.0, `terraform-plugin-go`
  v0.31.0, `terraform-plugin-testing` v1.16.0, `terraform-plugin-docs` v0.25.0,
  `terraform-plugin-framework-timeouts` v0.7.0).
- Modernized the `golangci-lint` configuration for current linter names and pinned
  the CI linter version for reproducible builds.
