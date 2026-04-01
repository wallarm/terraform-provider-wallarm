---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_mode"
subcategory: "Rules"
description: |-
  Provides the "Set filtration mode" rule resource.
---

# wallarm_rule_mode

Provides the resource to manage rules with the "[Set filtration mode][1]" action type. This rule type allows you to enable and disable the blocking of requests to various parts of a web application.

## Example Usage

```hcl
# Sets the `monitoring` mode for all the requests
# sent to the application with ID 9 via HTTPS protocol.

resource "wallarm_rule_mode" "tiredful_api_mode" {
  mode =  "monitoring"

  action {
    point = {
      instance = 9
    }
  }

  action {
    type = "equal"
    point = {
      scheme = "https"
    }
  }

  action {
    type = "equal"
    value = "admin"
    point = {
      query = "user"
    }
  }

}
```

## Argument Reference

* `mode` - (**required**) Wallarm node mode. Can be: `off`, `block`, `monitoring`, `default`. Aids to enable block mode granularly or turn off the Wallarm node for certain request parts.
* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - Type of the created rule. For example, `rule_type = "wallarm_mode"`.

## Import

```
$ terraform import wallarm_rule_mode.api_mode 6039/563855/11086881/monitoring
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `monitoring` - Wallarm mode value (`monitoring`, `block`, `off`, or `default`).

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

This resource is a **mitigation control**. For an overview of all mitigation controls and their parameter mapping, see the [Mitigation Controls Guide](../guides/mitigation_controls).

[1]: https://docs.wallarm.com/user-guides/rules/wallarm-mode-rule/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
