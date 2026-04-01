---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_bola"
subcategory: "Rule"
description: |-
  Provides the "BOLA protection" mitigation control resource.
---

# wallarm_rule_bola

Provides the resource to manage mitigation control with the "[BOLA protection][1]" action type. It provides protection against attempts of enumeration of user IDs, object IDs, and filenames. It counts the number of unique values seen for each parameter within a specified timeframe.


## Example Usage

```hcl
resource "wallarm_rule_bola" "wallarm_rule_bola_regexp" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  
  reaction {
    block_by_ip = 600
  }

  threshold {
    count = 5
    period = 30
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["bar"]
    additional_parameters = false
    plain_parameters      = false
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `advanced_conditions` - (optional) built-in parameters of requests which the mitigation control will be applied to.
* `arbitrary_conditions` - (optional) session context parameters which the mitigation control will be applied to.
* `enumerated_parameters` - (**required**) parameters to be tracked for enumeration attempts
* `threshold` - (**required**) threshold number of unique endpoints accessed in a configured timeframe (in seconds).
* `mode` - (**required**) protection behaviour which will be applied to the detected attack. Possible values: `monitoring`, `block`.
* `reaction` - (**required**) action which will be performed once threshold is breached.

**advanced_conditions**

`advanced_conditions` - built-in parameters of requests - elements of meta information presented in each request handled by Wallarm filtering node.

* `field` - (**required**) value of the parameter you want to apply the mitigation control to. Possible values: `ip`, `method`, `user_agent`, `domain`, `uri`, `status_code`, `request_time`, `request_size`, `response_size`, `attack_type`, `blocked`.
* `value` - (**required**) expected field value (should be list of strings) 
* `operator` - (**required**) condition type. Can be: `eq`, `ne`, `imatch`, `notimatch`, `match`, `notmatch`, `lt`, `gt`, `le`, `ge`.

**arbitrary_conditions**

`arbitrary_conditions` - parameters from ones, that were defined as important in API Sessions.

* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.
* `value` - (**required**) value of the parameter you want to apply the mitigation control to.
* `operator` - (**required**) condition type. Can be: `eq`, `ne`, `imatch`, `notimatch`, `match`, `notmatch`, `lt`, `gt`, `le`, `ge`.

**enumerated_parameters**

`enumerated_parameters` - parameters that will be monitored for enumeration.

* `mode` - parameters to be monitored via exact or or regex match (only one approach can be used within single mitigation control). Possible values: `regexp` , `exact`.
Parameters for mode `exact`:
* `points` - API Session parameters to be tracked for enumeration attempts.
Parameters for mode `regexp`:
* `name_regexps` - regular expression for parameter name.
* `value_regexps` - regular expression for parameter value.
* `additional_parameters` - session context parameters.
* `plain_parameters` - URL & query string parameters.

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
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "enum"`.

## Import

```
$ terraform import wallarm_rule_bola.wallarm_rule_bola_regexp 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

This resource is a **mitigation control**. For an overview of all mitigation controls and their parameter mapping, see the [Mitigation Controls Guide](../guides/mitigation_controls).

[1]: https://docs.wallarm.com/api-protection/enumeration-attack-protection/#mitigation-controls
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/