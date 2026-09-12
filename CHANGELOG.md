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

BUG FIXES:

- **`vpsie_dns_record`:** fixed deletion, which posted to a non-existent
  `/domain/dnsRecord/delete` path and always failed with "Not found"; it now uses
  the correct `DELETE /domain/dnsRecord` route. Read is now implemented against the
  parent domain (records are matched by name/type/content) so drift and out-of-band
  deletions are detected, and the resource can be imported via
  `domain_identifier/type/name/content`.
- **`vpsie_backup_policy`:** fixed reads, which queried the wrong (plural)
  `/backups/policy/:id` route and returned "Not found"; the singular
  `/backup/policy/:id` route is now used. Attached VMs are now parsed correctly
  (the API returns VM objects, not strings), and a policy deleted out of band is
  removed from state instead of erroring.
- **`vpsie_snapshot_policy`:** a policy deleted out of band is now removed from
  state instead of erroring.

ENHANCEMENTS:

- Added an `endpoint` provider argument (and `VPSIE_ENDPOINT`) to target
  non-production API environments.

- Updated the HashiCorp Terraform Plugin Framework and companion modules to their
  current releases (`terraform-plugin-framework` v1.19.0, `terraform-plugin-go`
  v0.31.0, `terraform-plugin-testing` v1.16.0, `terraform-plugin-docs` v0.25.0,
  `terraform-plugin-framework-timeouts` v0.7.0).
- Modernized the `golangci-lint` configuration for current linter names and pinned
  the CI linter version for reproducible builds.
