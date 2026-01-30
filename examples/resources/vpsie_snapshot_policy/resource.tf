resource "vpsie_snapshot_policy" "example" {
  name        = "daily-snapshot"
  backup_plan = "daily"
  plan_every  = "1"
  keep        = "7"
}
