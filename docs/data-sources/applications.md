---
layout: "wallarm"
page_title: "Wallarm: wallarm_applications"
subcategory: "Data Source"
description: |-
  Reads existing application pools from the Wallarm API.
---

# wallarm_applications

Reads all application pools configured for the specified client from the Wallarm API. Used by the import module to discover existing applications and generate Terraform import blocks.

## Example Usage

```hcl
data "wallarm_applications" "all" {
  client_id = 12345
}

output "app_list" {
  value = data.wallarm_applications.all.applications
}
```

## Argument Reference

* `client_id` - (Optional) ID of the client to query. Defaults to the provider's default client ID.

## Attributes Reference

* `applications` - List of application objects, each containing:
  * `app_id` - (Int) Application ID. The default application has `app_id = -1`.
  * `name` - (String) Application name.
  * `client_id` - (Int) Client ID the application belongs to.
