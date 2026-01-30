resource "vpsie_backup" "example" {
  name          = "my-backup"
  backup_key    = "backup-key"
  vm_identifier = "vm-identifier"
}
