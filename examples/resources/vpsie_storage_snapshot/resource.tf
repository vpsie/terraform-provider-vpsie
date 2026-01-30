resource "vpsie_storage_snapshot" "example" {
  name               = "my-storage-snapshot"
  type               = "snapshot"
  storage_identifier = "storage-identifier"
}
