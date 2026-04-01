---
layout: "wallarm"
page_title: "Wallarm: wallarm_applications"
subcategory: "Common"
description: |-
  Reads existing application pools from the Wallarm API.
---

# wallarm_applications

Reads all application pools configured for the specified client from the Wallarm API. Used by the import module to discover existing applications and generate Terraform import blocks.

## Example Usage

```hcl
data "wallarm_applications" "all" {}

output "app_list" {
  value = data.wallarm_applications.all.applications
}
```

### Generate Import Blocks

Use the data source to discover existing applications and generate import blocks for `terraform plan -generate-config-out`:

```hcl
data "wallarm_applications" "all" {}

locals {
  client_id = 8649
}

resource "local_file" "app_imports" {
  filename = "${path.module}/wallarm_application_imports.tf"
  content = join("\n", [
    for app in data.wallarm_applications.all.applications :
    <<-EOT
    import {
      to = wallarm_application.app_${app.app_id}
      id = "${local.client_id}/${app.app_id}"
    }
    EOT
    if app.app_id != -1  # skip default application
  ])
}
```

## Argument Reference

* `client_id` - (Optional) ID of the client to query. Defaults to the provider's default client ID.

## Attributes Reference

* `applications` - List of application objects, each containing:
  * `app_id` - (Int) Application ID. The default application has `app_id = -1`.
  * `name` - (String) Application name.
  * `client_id` - (Int) Client ID the application belongs to.
