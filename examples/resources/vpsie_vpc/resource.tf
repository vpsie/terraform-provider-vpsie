resource "vpsie_vpc" "example" {
  name          = "my-vpc"
  dc_identifier = "dc-identifier"
  auto_generate = 1
  network_size  = "24"
  network_range = "10.0.0.0"
}
