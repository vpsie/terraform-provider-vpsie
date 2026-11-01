resource "vpsie_protect_policy" "nightly" {
  name        = "nightly-backup"
  tier        = "backup"
  datacenter  = "replace-with-datacenter-identifier"
  rpo_minutes = 1440

  retention = {
    keep_recent_hours = 24
    keep_daily        = 7
    keep_weekly       = 4
  }

  immutability_days = 7
  verify_mode       = "latest-weekly"
}
