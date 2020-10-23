---
layout: "wallarm"
page_title: "Wallarm: wallarm_application"
subcategory: "Common"
description: |-
  Provides the resource to manage applications of the account.
---

# wallarm_user

Provides the resource to manage applications of the account.

## Example Usage

```hcl
# Creates a new application

resource "wallarm_application" "tf_app" {
  name = "New Terraform Application"
  app_id = 42
}
```

## Argument Reference

* `app_id` - (Required) Application ID. Unique ID identifying a logical part of the website.
* `name` - (Required) Application name.
* `client_id` - (Optional) ID of the client to apply the rules to. The value is required for multi-tenant scenarios.
