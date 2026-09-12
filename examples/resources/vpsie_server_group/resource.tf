resource "vpsie_server_group" "example" {
  group_name        = "web-tier"
  group_description = "Front-end web servers"
  is_distributed    = true
}
