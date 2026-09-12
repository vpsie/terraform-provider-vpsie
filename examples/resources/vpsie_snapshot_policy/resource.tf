resource "vpsie_snapshot_policy" "example" {
  name        = "daily-snapshot"
  backup_plan = "day" # one of: day, week, month
  plan_every  = "1"
  keep        = "5" # max 5
}
