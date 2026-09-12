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

- **`vpsie_registry`:** registry creation never worked — the request sent the
  wrong field names and omitted the project. The API expects `datacenterId`,
  `resourceIdentifier` (the plan), and a `projectId` (the project's UUID); a new
  required `project_identifier` argument supplies the project, and the create
  request now uses the correct fields. The list response was also mismodeled
  (`registry_id`/`registry_name`/`dc_name`), so reads returned blanks; the SDK
  types are corrected. Create, read, no-drift plan, and delete are verified
  against the live API. (On import, the plan and project are returned only as
  numeric ids, so those two identifiers must be set in config.)
- **`vpsie_server`:** server creation could never succeed. The project was sent
  as a numeric `project_id`, but the API requires the project's UUID identifier
  as a string (it rejects a number outright); a new required `project_identifier`
  argument carries it, and `project_id` is now a computed numeric attribute. The
  configured `password` was also silently dropped from the create request — it is
  now sent (and the API is asked to generate one only when it is omitted), and
  `initial_password` is marked sensitive. (Full provisioning could not be
  re-verified end-to-end on the test backend, whose hypervisor nodes were
  unreachable; the create request is accepted by the API with these fixes.)
- **`vpsie_storage`:** fixed two defects. Updating the volume (rename or resize)
  crashed with a value-conversion error because the Int64 `size` attribute was
  read as a string; and reads overwrote config-owned fields with the API's
  normalized values (notably `disk_format`, which the API stores as `MANUAL`),
  causing perpetual drift. Reads now keep the configured values and refresh only
  computed fields. The example's invalid `storage_type`/`disk_format` values were
  corrected, and an import example was added.
- **`vpsie_monitoring_rule`:** reworked to the API's actual shape. The resource
  sent a stale flat body (single metric, single action) the backend no longer
  accepts; it now takes one or more `rule` blocks, each with its own `action`
  blocks, plus `vms`. Reads, status toggling, VM attach/detach, delete and import
  all work, and the `vpsie_monitoring_rules` data source now returns each rule's
  metrics and actions.
- **`vpsie_firewall`:** the resource was non-functional — its schema and model
  were misaligned (a value-conversion crash), rules were marked read-only so they
  could never be set, and delete failed with HTTP 500 because the required
  `deleteStatistic` payload was missing. The resource is rewritten with a
  settable `rule` block (validated `action`/`type`), full create/read/update
  (rename plus in-place rule reconciliation)/delete, drift-safe reads, and import
  by group identifier.
- **`vpsie_firewalls` (data source):** never wrote its result to state and could
  not decode the rules payload (the API returns an object, not a list). Fixed the
  SDK types and the data source so it returns groups with their rules.
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
