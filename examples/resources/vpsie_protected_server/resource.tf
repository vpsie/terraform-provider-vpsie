resource "vpsie_protected_server" "db01" {
  policy = vpsie_protect_policy.nightly.identifier
  source = vpsie_migration_source.vcenter.identifier
  vm_ref = "vm-1042"

  # Set to true to pause the copies; back to false to resume them.
  paused = false
}
