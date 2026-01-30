resource "vpsie_backup_policy" "example" {
  name        = "daily-backup"
  backup_plan = "daily"
  plan_every  = "1"
  keep        = "7"
}
