---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_mode"
subcategory: "Rule"
description: |-
  Provides the "Set filtration mode" rule resource.
---

# wallarm_rule_mode

Provides the resource to manage rules with the "[Set filtration mode][1]" action type. This rule type allows you to enable and disable the blocking of requests to various parts of a web application.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

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

* `mode` - (**required**) Wallarm node mode. Can be: `off`, `block`, `monitoring`, `safe_blocking`, `default`. Aids to enable block mode granularly or turn off the Wallarm node for certain request parts.
* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - TType of the created rule. For example, `rule_type = "wallarm_mode"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

ID should end with a wallarm_mode value.

```
$ terraform import wallarm_rule_mode.api_mode 6039/563855/11086881/monitoring
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_mode` - Terraform resource rule type.
* `monitoring` - Wallarm mode.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_mode" "api_mode" {
  action {
    point = {
      path = 0
    }
    type = "equal"
    value = "api"
  }
  action {
    point = {
      instance = 9
    }
  }
  mode = "monitoring"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_mode.api_mode
  id = "6039/563855/11086881/monitoring"
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

[1]: https://docs.wallarm.com/user-guides/rules/wallarm-mode-rule/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
