---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_disable_stamp"
subcategory: "Rules"
description: |-
  Provides the "Disable specific attack stamp" rule resource.
---

# wallarm_rule_disable_stamp

Provides the resource to manage rules with the "Disable specific attack stamp" action type. Disables detection of a specific attack type stamp for selected request points.

## Example Usage

```hcl
resource "wallarm_rule_disable_stamp" "disable_stamp" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
  stamp = 1234
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][1].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `stamp` - (**required**) integer attack type stamp to disable. Must be greater than 0.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "disable_stamp"`.

## Import

```
$ terraform import wallarm_rule_disable_stamp.disable_stamp 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
