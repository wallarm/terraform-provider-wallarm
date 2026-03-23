---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_overlimit_res_settings"
subcategory: "Rule"
description: |-
  Provides the "Overlimit Res Settings" rule resource.
---

# wallarm_rule_overlimit_res_settings

This rule enables you with setting up a custom time limit for a single request processing and changing the default node behavior.


## Example Usage

```hcl
resource "wallarm_rule_overlimit_res_settings" "example_overlimit_res_settings" {
  action {
    point = {
      path = 0
    }
    type = "absent"
  }
  action {
    point = {
      action_name = "upload"
    }
    type = "equal"
  }
  action {
    point = {
      action_ext = ""
    }
    type = "absent"
  }
  mode = "blocking"
  overlimit_time = 2000
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][1].
* `overlimit_time` - (required) Specifies the overlimit time limit in ms.
* `mode` - (required) Specifies the overlimit res mode. Can be: `off`, `monitoring`, `blocking`.
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "overlimit_res_settings"`.

## Import

```
$ terraform import wallarm_rule_overlimit_res_settings.uploads_overlimit 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
