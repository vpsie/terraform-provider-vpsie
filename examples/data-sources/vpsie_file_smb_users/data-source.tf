data "vpsie_file_smb_users" "all" {}

output "smb_user_names" {
  value = [for u in data.vpsie_file_smb_users.all.users : u.username]
}
