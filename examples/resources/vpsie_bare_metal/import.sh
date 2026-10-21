# Bare-metal servers are imported by their identifier. The API does not return
# the plan, OS, password or other install-time inputs, so the first apply after
# the import takes them from the configuration without replacing the server.
terraform import vpsie_bare_metal.example "bm-identifier"
