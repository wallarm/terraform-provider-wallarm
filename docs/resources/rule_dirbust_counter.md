---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_dirbust_counter"
subcategory: "Rule"
description: |-
  Provides the "Define force browsing attacks counter" rule resource.
---

# wallarm_rule_dirbust_counter

Provides the resource to manage rules with the "Define force browsing attacks counter" action type. For detecting force browsing attacks, there is a counter that increments whenever a request hits 404 status code (resource not found). By default, every application has its own counter.

This rule should be used when an independent detection of force browsing attacks is required for different parts of the application.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_dirbust_counter" "login_counter" {
  action {
    type = "iequal"
    point = {
      action_name = "login"
    }
  }
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `counter` - Name of the counter. Randomly generated, but always starts with `d:`.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "dirbust_counter"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_dirbust_counter.login_counter 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.
* `wallarm_rule_dirbust_counter` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_dirbust_counter" "login_counter" {
  action {
    type = "iequal"
    point = {
      action_name = "login"
    }
  }
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_dirbust_counter.login_counter
  id = "6039/563854/11086884"
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

[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
