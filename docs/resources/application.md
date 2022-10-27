---
layout: "wallarm"
page_title: "Wallarm: wallarm_application"
subcategory: "Common"
description: |-
  Provides the resource to manage applications of the account.
---

# wallarm_application

Provides the resource to manage [applications][1] of the account.

## Example Usage

```hcl
# Creates a new application

resource "wallarm_application" "tf_app" {
  name = "New Terraform Application"
  app_id = 42
}
```

## Argument Reference

* `app_id` - (**required**) application ID. Unique ID identifying a logical part of the website.
* `name` - (**required**) application name.
* `client_id` - (optional) ID of the client to apply the rules to. The value is required for multi-tenant scenarios.

[1]: https://docs.wallarm.com/user-guides/settings/applications/
