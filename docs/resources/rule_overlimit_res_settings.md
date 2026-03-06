---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_overlimit_res_settings"
subcategory: "Rule"
description: |-
  Provides the "Overlimit Res Settings" rule resource.
---

# wallarm_rule_overlimit_res_settings

This rule enables you to set up a custom time limit for a single request processing and changing the default node behavior.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex). This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

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
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "overlimit_res_settings"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_overlimit_res_settings.uploads_overlimit 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_overlimit_res_settings` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_overlimit_res_settings" "uploads_overlimit" {
  action {
    point = {
      action_ext = ""
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
      path = 0
    }
    type = "absent"
  }
  mode = "blocking"
  overlimit_time = 2000
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_overlimit_res_settings.uploads_overlimit
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


[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
