resource "vpsie_file_share_smb" "data" {
  share_identifier = vpsie_file_share.data.identifier
  enabled          = true

  users = [
    {
      identifier = vpsie_file_smb_user.alice.identifier
      mode       = "rw"
    },
  ]
}
