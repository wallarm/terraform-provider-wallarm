---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_bruteforce_counter"
subcategory: "Rule"
description: |-
  Provides the "Define brute-force attacks counter" rule resource.
---

# wallarm_rule_bruteforce_counter

Provides the resource to manage rules with the "Define brute-force attacks counter" action type. For detecting brute-force attacks, with every request, one of the statistical counters is incremented. By default, the counter name is automatically defined based on the domain name and the request path.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
# Sets a counter on the root `/` path

resource "wallarm_rule_bruteforce_counter" "root_counter" {
  action {
    type = "iequal"
    value = "/"
    point = {
      path = 0
    }
  }
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][1].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `counter` - Name of the counter. Randomly generated, but always starts with `b:`.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "brute_counter"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_bruteforce_counter.root_counter 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.
* `wallarm_rule_bruteforce_counter` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_bruteforce_counter" "root_counter" {
  action {
    type = "iequal"
    value = "/"
    point = {
      path = 0
    }
  }
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_bruteforce_counter.root_counter
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

[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
