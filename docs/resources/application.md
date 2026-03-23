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

* `app_id` - (**required**) application ID.
* `name` - (**required**) application name.
* `client_id` - (optional) ID of the client (tenant). The value is required for [multi-tenant scenarios][2].

## Import

```
$ terraform import wallarm_application.tf_app 8649/42
```

* `8649` - Client ID.
* `42` - Application ID.

Existing applications can be discovered using the [`wallarm_applications`](../data-sources/applications) data source, which returns `app_id` and `client_id` for each application.

~> **Note:** The default application (`app_id = -1`) is protected from deletion and should not be imported.

[1]: https://docs.wallarm.com/user-guides/settings/applications/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
