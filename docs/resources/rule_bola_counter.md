---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_bola_counter"
subcategory: "Rule"
description: |-
  Provides the resource to manage BOLA (IDOR) counter rules.
---

# wallarm_rule_bola_counter

Provides the resource to manage BOLA (IDOR) counter rules in Wallarm. This rule defines a counter that increments when requests matching the specified action conditions are detected, enabling BOLA/IDOR attack detection.

The counter works in conjunction with `wallarm_trigger` to detect and mitigate BOLA attacks by tracking request patterns to endpoints with enumerable resource identifiers.

## Example Usage

```hcl
resource "wallarm_rule_bola_counter" "example" {
  action {
    type  = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  action {
    type  = "equal"
    value = "api"
    point = {
      path = "0"
    }
  }
}
```

## Argument Reference

* `client_id` - (Optional) ID of the client to apply the rules to. The value is required for multi-tenant scenarios.
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `counter` - Name of the counter (auto-generated).
* `action_id` - The action ID (computed).
* `rule_type` - The type of the created rule.

## Import

The resource can be imported using a composite ID formed of `client_id`/`action_id`/`rule_id`, e.g.:

```
$ terraform import wallarm_rule_bola_counter.example 12345/67890/12
```
