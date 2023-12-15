---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vpsie_storage Resource - terraform-provider-vpsie"
subcategory: ""
description: |-
  
---

# vpsie_storage (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `dc_identifier` (String)
- `description` (String)
- `disk_format` (String)
- `name` (String)
- `size` (Number)
- `storage_type` (String)

### Optional

- `box_id` (Number)
- `hostname` (String)
- `is_automatic` (Number)
- `vm_identifier` (String)

### Read-Only

- `bus_device` (String)
- `bus_number` (Number)
- `created_on` (String)
- `disk_key` (String)
- `id` (Number) The ID of this resource.
- `identifier` (String)
- `os_identifier` (String)
- `state` (String)
- `storage_id` (Number)
- `user_id` (Number)
- `user_template_id` (Number)