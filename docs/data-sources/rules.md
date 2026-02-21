---
layout: "wallarm"
page_title: "Wallarm: wallarm_rules"
subcategory: "Rule"
description: |-
  Get a list of Wallarm rules (hints).
---

# wallarm_rules

Use this data source to retrieve all rules (hints) configured for a client. Supports optional filtering by rule type. Results are paginated automatically.

A common use case is discovering existing `disable_stamp` rules before adopting them into Terraform state via import blocks (Terraform 1.5+).

## Example Usage

```hcl
# Retrieve all rules for the authenticated client
data "wallarm_rules" "all" {}

output "total_rules" {
  value = length(data.wallarm_rules.all.rules)
}
```

```hcl
# Retrieve only disable_stamp rules (false positive suppressions)
data "wallarm_rules" "stamp_rules" {
  types = ["disable_stamp"]
}

output "stamp_rules" {
  value = data.wallarm_rules.stamp_rules.rules
}
```

```hcl
# Use with jsondecode() to inspect rule points
data "wallarm_rules" "stamps" {
  types = ["disable_stamp"]
}

output "stamp_points" {
  value = [for r in data.wallarm_rules.stamps.rules : jsondecode(r.point)]
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to retrieve rules for. The value is required for [multi-tenant scenarios][1]. Defaults to the client associated with the API token.
* `types` - (optional) List of rule types to filter by. If omitted, all rule types are returned. Common values: `disable_stamp`, `vpatch`, `regex`, `disable_attack_type`, `rule_mode`, `masking`, `parser_state`.

## Attributes Reference

`rules` - List of rule objects. Each object contains:

* `rule_id` - ID of the rule (hint).
* `action_id` - ID of the action (the set of matching conditions the rule is attached to).
* `client_id` - Client ID the rule belongs to.
* `type` - Rule type string (e.g. `disable_stamp`, `vpatch`).
* `enabled` - Whether the rule is currently enabled.
* `mode` - Rule mode, if applicable (e.g. `monitoring`, `block`).
* `regex` - Regular expression string, for regex-based rule types.
* `attack_type` - Attack type the rule targets, if applicable.
* `name` - Rule name, if set.
* `point` - JSON-encoded string of the request point the rule applies to. Use `jsondecode(rule.point)` in Terraform to work with it as a list.
* `action` - JSON-encoded string of the rule's matching conditions (action details). Use `jsondecode(rule.action)` to inspect individual conditions.
* `create_time` - Unix timestamp when the rule was created.
* `updated_at` - Unix timestamp when the rule was last updated.

[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
