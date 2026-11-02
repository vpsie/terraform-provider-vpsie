resource "vpsie_file_share" "data" {
  name          = "data"
  dc_identifier = "replace-with-datacenter-identifier"
  vpc_id        = 42
  size_gb       = 100

  access = [
    {
      cidr = "10.0.0.0/24"
      mode = "rw"
    },
    {
      vm_identifier = "replace-with-server-identifier"
      mode          = "ro"
    },
  ]
}

output "mount_command" {
  value = vpsie_file_share.data.mount_nfs4
}
