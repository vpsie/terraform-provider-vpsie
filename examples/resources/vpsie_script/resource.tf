resource "vpsie_script" "example" {
  script_name = "setup-script"
  script      = "#!/bin/bash\napt-get update"
  type        = "bash"
}
