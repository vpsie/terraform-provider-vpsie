resource "vpsie_backup_policy" "example" {
  name        = "daily-backup"
  backup_plan = "day" # one of: day, week, month
  plan_every  = "1"
  keep        = "5" # max 5
}
