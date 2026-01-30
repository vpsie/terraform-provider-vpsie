resource "vpsie_bucket" "example" {
  bucket_name   = "my-bucket"
  project_id    = "project-identifier"
  datacenter_id = "dc-identifier"
}
