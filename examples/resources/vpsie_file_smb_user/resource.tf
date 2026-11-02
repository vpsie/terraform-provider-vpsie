variable "smb_password" {
  type      = string
  sensitive = true
}

# Name the VPC with the datacenter and VPC id, or use share_identifier to take the VPC of a share.
resource "vpsie_file_smb_user" "alice" {
  username      = "alice"
  password      = var.smb_password
  dc_identifier = "replace-with-datacenter-identifier"
  vpc_id        = 42
}
