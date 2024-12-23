---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "koyeb_secret Resource - terraform-provider-koyeb"
subcategory: ""
description: |-
  Secret resource in the Koyeb Terraform provider.
---

# koyeb_secret (Resource)

Secret resource in the Koyeb Terraform provider.

## Example Usage

```terraform
resource "koyeb_secret" "simple-secret" {
  name  = "secret-name"
  value = "secret-value"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The secret name

### Optional

- `azure_container_registry` (Block Set, Max: 1) The azure_container_registry configuration to use (see [below for nested schema](#nestedblock--azure_container_registry))
- `digital_ocean_container_registry` (Block Set, Max: 1) The digital_ocean_container_registry configuration to use (see [below for nested schema](#nestedblock--digital_ocean_container_registry))
- `docker_hub_registry` (Block Set, Max: 1) The docker_hub_registry configuration to use (see [below for nested schema](#nestedblock--docker_hub_registry))
- `github_registry` (Block Set, Max: 1) The github_registry configuration to use (see [below for nested schema](#nestedblock--github_registry))
- `gitlab_registry` (Block Set, Max: 1) The gitlab_registry configuration to use (see [below for nested schema](#nestedblock--gitlab_registry))
- `private_registry` (Block Set, Max: 1) The private_registry configuration to use (see [below for nested schema](#nestedblock--private_registry))
- `type` (String) The secret type
- `value` (String, Sensitive) The secret value

### Read-Only

- `created_at` (String) The date and time of when the secret was created
- `id` (String) The secret ID
- `organization_id` (String) The organization ID owning the secret
- `updated_at` (String) The date and time of when the secret was last updated

<a id="nestedblock--azure_container_registry"></a>
### Nested Schema for `azure_container_registry`

Required:

- `password` (String, Sensitive) The registry password
- `registry_name` (String) The registry name
- `username` (String) The registry username


<a id="nestedblock--digital_ocean_container_registry"></a>
### Nested Schema for `digital_ocean_container_registry`

Required:

- `password` (String, Sensitive) The registry password
- `username` (String) The registry username


<a id="nestedblock--docker_hub_registry"></a>
### Nested Schema for `docker_hub_registry`

Required:

- `password` (String, Sensitive) The registry password
- `username` (String) The registry username


<a id="nestedblock--github_registry"></a>
### Nested Schema for `github_registry`

Required:

- `password` (String, Sensitive) The registry password
- `username` (String) The registry username


<a id="nestedblock--gitlab_registry"></a>
### Nested Schema for `gitlab_registry`

Required:

- `password` (String, Sensitive) The registry password
- `username` (String) The registry username


<a id="nestedblock--private_registry"></a>
### Nested Schema for `private_registry`

Required:

- `password` (String, Sensitive) The registry password
- `url` (String) The registry URL
- `username` (String) The registry username


