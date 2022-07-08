---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "koyeb_app Data Source - terraform-provider-koyeb"
subcategory: ""
description: |-
  
---

# koyeb_app (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The app name

### Read-Only

- `created_at` (String) The date and time of when the app was created
- `domains` (List of Object) The app domains (see [below for nested schema](#nestedatt--domains))
- `id` (String) The app id
- `organization_id` (String) The organization id owning the app
- `updated_at` (String) The date and time of when the app was last updated

<a id="nestedatt--domains"></a>
### Nested Schema for `domains`

Read-Only:

- `app_name` (String)
- `created_at` (String)
- `deployment_group` (String)
- `id` (String)
- `intended_cname` (String)
- `messages` (String)
- `name` (String)
- `organization_id` (String)
- `status` (String)
- `type` (String)
- `updated_at` (String)
- `verified_at` (String)
- `version` (String)

