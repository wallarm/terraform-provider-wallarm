---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_variative_values"
subcategory: "Rule"
description: |-
  Provides the "Make a certain conditions point variative" rule resource.
---

# wallarm_rule_variative_values

!> The resource will be deprecated in the future versions.

Provides the resource to manage rules with the "Make a certain conditions point variative" action type. Specifies the condition point to group rules by. Notice that you may not have permissions to use this resource.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_variative_values" "action_name" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["action_name"]]
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `point` - (**required**) condition point to apply the rules to.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "variative_values"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_variative_values.action_name 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_variative_values` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_variative_values" "action_name" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["action_name"]]
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_variative_values.action_name
  id = "6039/563855/11086881"
}
```

Before importing resources run:

```
$ terraform plan
```

If import looks good apply the configuration:

```
$ terraform apply
```

[1]: https://docs.wallarm.com/user-guides/rules/rules/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
