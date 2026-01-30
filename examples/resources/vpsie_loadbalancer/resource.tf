resource "vpsie_loadbalancer" "example" {
  lb_name    = "my-loadbalancer"
  traffic    = 1000
  boxsize_id = 1
}
