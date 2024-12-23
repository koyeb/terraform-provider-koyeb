---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "koyeb_volume Resource - terraform-provider-koyeb"
subcategory: ""
description: |-
  Volume resource in the Koyeb Terraform provider.
---

# koyeb_volume (Resource)

Volume resource in the Koyeb Terraform provider.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `max_size` (Number) The maximum size of the volume in GB
- `name` (String) The volume name
- `region` (String) The region where the volume is located

### Optional

- `read_only` (Boolean) If set to true, the volume will be mounted in read-only
- `volume_type` (String) The volume type

### Read-Only

- `backing_store` (String) The backing store of the volume
- `created_at` (String) The date and time of when the volume was created
- `cur_size` (Number) The current size of the volume in GB
- `id` (String) The volume ID
- `organization_id` (String) The organization ID owning the volume
- `service_id` (String) The service ID the volume is attached to
- `snapshot_id` (String) The snapshot ID the volume was created from
- `status` (String) The status of the volume
- `updated_at` (String) The date and time of when the volume was last updated


