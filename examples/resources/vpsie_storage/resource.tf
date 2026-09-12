resource "vpsie_storage" "example" {
  name          = "my-storage"
  dc_identifier = "dc-identifier"
  size          = 50
  storage_type  = "SATA" # SATA, SSD, or LOCAL
  disk_format   = "XFS"  # XFS or REFS
  description   = "Example storage volume"
}
