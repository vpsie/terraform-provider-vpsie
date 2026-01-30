resource "vpsie_image" "example" {
  fetched_from_url = "https://example.com/image.qcow2"
  image_label      = "my-custom-image"
  dc_identifier    = "dc-identifier"
}
