---
layout: "wallarm"
page_title: "Wallarm: wallarm_actions"
subcategory: "Common"
description: |-
  Discovers all non-empty action scopes from the Wallarm API.
---

# wallarm_actions

Discovers all non-empty action scopes (rule groups) from the Wallarm API with pagination. Primarily used in advanced configurations for organizing rules by action scope.

## Example Usage

```hcl
data "wallarm_actions" "all" {}

output "action_scopes" {
  value = data.wallarm_actions.all.actions
}
```

## Argument Reference

* `client_id` - (Optional) Client ID. Defaults to the provider's client ID.

## Attributes Reference

* `actions` - List of action scopes, each containing:
  * `action_id` - (Int) Action ID.
  * `conditions` - List of action conditions, each with `type`, `point`, and `value`.
  * `conditions_hash` - (String) SHA256 hash of sorted conditions.
  * `dir_name` - (String) Computed directory name for file-based rule organization.
  * `endpoint_path` - (String) Endpoint path, if set.
  * `endpoint_domain` - (String) Endpoint domain, if set.
  * `endpoint_instance` - (String) Endpoint instance (pool ID), if set.
  * `updated_at` - (Int) Last update timestamp.
