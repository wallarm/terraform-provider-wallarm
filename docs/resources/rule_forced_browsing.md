---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_forced_browsing"
subcategory: "Rule"
description: |-
  Provides the "Forced browsing protection" mitigation control resource.
---

# wallarm_rule_forced_browsing

Provides the resource to manage mitigation control with the "[Forced browsing protection][1]" action type. For detecting force browsing attacks, there is a counter that increments whenever a request hits 404 status code (resource not found). By default, every application has its own counter.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_forced_browsing" "forced_browsing" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  
  advanced_conditions {
    field = "method"
    value = ["POST"]
    operator = "eq"
  }

  arbitrary_conditions {
    point = [["header","TEST"]]
    value = ["98676877"]
    operator = "eq"
  }
  
  reaction {
    block_by_ip = 600
  }

  threshold {
    count = 5
    period = 30
  }

}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `advanced_conditions` - (optional) built-in parameters of requests which the mitigation control will be applied to.
* `arbitrary_conditions` - (optional) session context parameters which the mitigation control will be applied to.
* `threshold` - (**required**) threshold number of unique endpoints accessed in a configured timeframe (in seconds).
* `mode` - (**required**) protection behaviour which will be applied to the detected attack. Possible values: `monitoring`, `block`.
* `reaction` - (**required**) action which will be performed once treshhold is breached.

**advanced_conditions**

`advanced_conditions` - built-in parameters of requests - elements of meta information presented in each request handled by Wallarm filtering node.

* `field` - (**required**) value of the parameter you want to apply the mitigation control to. Possible values: `ip`, `method`, `user_agent`, `domain`, `uri`, `status_code`, `request_time`, `request_size`, `response_size`, `attack_type`, `blocked`.
* `value` - (**required**) expected field value (should be list of strings) 
* `operator` - (**required**) condition type. Possible values: `eq`, `ne`, `imatch`, `notimatch`, `match`, `notmatch`, `lt`, `gt`, `le`, `ge`.

**arbitrary_conditions**

`arbitrary_conditions` - quickly select parameters from the list of ones, that were defined as important in API Sessions.

* `point` - (**required**) parameter name 
* `value` - (**required**) value of the parameter you want to apply the mitigation control to.
* `operator` - (**required**) condition type. Possible values: `eq`, `ne`, `imatch`, `notimatch`, `match`, `notmatch`, `lt`, `gt`, `le`, `ge`.

**threshold**

`threshold` - number of unique values for a parameter per timeframe, when exceeds the control will be triggered

* `period` - timeframe.
* `count` - number of unique values.

**reaction**

`reaction` - when the counter exceeds the threshold, the selected action is performed. You have to choose one of options below.

* `block_by_session` - block a session for the exact period in seconds.
* `block_by_ip` - block an IP for the exact period in seconds.
* `graylist_by_ip` - graylist an IP for the exact period in seconds.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `counter` - Name of the counter. Randomly generated, but always starts with `d:`.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "dirbust_counter"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_forced_browsing.forced_browsing 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.
* `wallarm_rule_forced_browsing` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_forced_browsing" "forced_browsing" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  mode = "block"
  reaction {
    block_by_ip = 600
  }
  threshold {
    count = 5
    period = 30
  }
  advanced_conditions {
    field    = "method"
    value    = ["POST"]
    operator = "eq"
  }
  arbitrary_conditions {
    point    = [["header", "TEST"]]
    value    = ["98676877"]
    operator = "eq"
  }
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_forced_browsing.forced_browsing
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

[1]: https://docs.wallarm.com/api-protection/enumeration-attack-protection/#mitigation-controls
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/