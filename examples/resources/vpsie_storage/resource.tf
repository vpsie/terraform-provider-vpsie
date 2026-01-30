resource "vpsie_storage" "example" {
  name          = "my-storage"
  dc_identifier = "dc-identifier"
  size          = 50
  storage_type  = "ssd"
  disk_format   = "raw"
  description   = "Example storage volume"
}
