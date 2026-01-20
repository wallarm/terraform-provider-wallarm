---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_enum"
subcategory: "Rule"
description: |-
  Provides the "Enumeration attack protection" mitigation control resource.
---

# wallarm_rule_enum

Provides the resource to manage mitigation control with the "[Enumeration attack protection][1]" action type. It can protect of your APIs to prevent the reveal of information highly valuable for malicious actors. By identifying valid usernames, email addresses, or system resources, attackers can significantly narrow their focus for subsequent attacks. Enum mitigation control counts the number of unique values seen for each parameter within a specified timeframe.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_enum" "wallarm_rule_enum_regexp" {
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
    additional_parameters = true
    plain_parameters      = true
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described below.
* `advanced_conditions` - (optional) built-in parameters of requests which the mitigation controll will be applied to.
* `arbitrary_conditions` - (optional) session context parameters which the mitigation controll will be applied to.
* `enumerated_parameters` - (**required**) parameters to be tracked for enumeration attempts
* `threshold` - (**required**) threshold number of unique endpoints accessed in a configured timeframe (in seconds).
* `mode` - (**required**) protection behaviour which will be applied to the detected attack. Possible values: `monitoring`, `block`.
* `reaction` - (**required**) action which will be performed once treshhold is breached.

**action**

`action` argument shares the available conditions which can be applied. The conditions are:

* `type` - (optional) condition type. Can be: `equal`, `iequal`, `regex`, `absent`. Must be omitted for the `instance` parameter in `point`.
  For more details, see the official [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
  Example:
  `type = "absent"`
* `value` - (optional) value of the parameter to match with. Must be omitted for the `instance` parameter in `point` or if `type` is `absent`.
  Example:
  `value = "example.com"`
* `point` - (optional) request parameters that trigger the rule. Possible values are described below. For more details, see the official [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

**advanced_conditions**

`advanced_conditions` - built-in parameters of requests - elements of meta information presented in each request handled by Wallarm filtering node.

* `field` - (**required**) value of the parameter you want to apply the mitigation control to. Possible values: `ip`, `method`, `user_agent`, `domain`, `uri`, `status_code`, `request_time`, `request_size`, `response_size`, `attack_type`, `blocked`.
* `value` - (**required**) expected field value (should be list of strings) 
* `operator` - (**required**) condition type. Can be: `eq`, `ne`, `imatch`, `notimatch`, `match`, `notmatch`, `lt`, `gt`, `le`, `ge`.

**arbitrary_conditions**

`arbitrary_conditions` - parameters from ones, that were defined as important in API Sessions.

* `point` - (**required**) parameter name 
* `value` - (**required**) value of the parameter you want to apply the mitigation control to.
* `operator` - (**required**) condition type. Can be: `eq`, `ne`, `imatch`, `notimatch`, `match`, `notmatch`, `lt`, `gt`, `le`, `ge`.

**enumerated_parameters**

`enumerated_parameters` - parameters that will be monitored for enumeration.

**Important:** `exact` is mode curruntly unavaliable in wallarm terraform provider.

* `mode` - parameters to be monitored via exact or or regex match (only one approach can be used within single mitigation control). Possible values: `regexp` , `exact`.
Parameters for mode `exact`:
* `points` - API Session parameters to be tracked for enumeration attempts.
Parameters for mode `regexp`:
* `name_regexps` - reqular expression for parameter name.
* `value_regexps` - reqular expression for parameter value.
* `additional_parameters` - session context parameters.
* `plain_parameters` - URL & query string parameters.

**threshold**

`threshold` - number of unique values for a parameter per timeframe, when exceeds the control will be triggered

* `period` - timeframe.
* `count` - number of unique values.

**reaction**

`reaction` - when the counter exceeds the threshold, the selected action is performed. You have to choose one of options below.

* `block_by_session` - block a session for the exacs period in seconds.
* `block_by_ip` - block an IP for the exacs period in seconds.
* `graylist_by_ip` - graylist an IP for the exacs period in seconds.

### Nested Objects

**point**

  * `header` - (optional) arbitrary HEADER parameter name.
  Example:
  `header = "HOST"`
  * `method` - (optional) request method. Can be: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
  Example:
  `method = "POST"`
  * `path` - (optional) array with URL parts separated by the `/` symbol (the last URL part is not included in the array). If there is only one part in the URL, the array will be empty.
  Example:
  `path = 0`
  * `action_name` - (optional) the last part of the URL after the `/` symbol and before the first period (`.`). This part of the URL is always present in the request even if its value is an empty string.
  Example:
  `action_name = "login"`
  * `action_ext` - (optional) the part of the URL after the last period (`.`). It may be missing in the request.
  Example:
  `action_ext = "php"`
  * `query` - (optional) the query parameter name.
  Example:
  `query = "user"`
  * `proto` - (optional) version of the HTTP Protocol.
  Example:
  `proto = "1.1"`
  * `scheme` - (optional) `http`/`https`.
  Example:
  `scheme = "https"`
  * `uri` - (optional) part of the request URL without domain.
  Example:
  `uri = "/api/login"`
  * `instance` - (optional) ID of the application.
  Example:
  `instance = 42`

Example:

  ```hcl
  # ... omitted

  action {
    type = "equal"
    point = {
      scheme = "https"
    }
  }

  action {
    point = {
      instance = 9
    }
  }

  action {
    type = "absent"
    point = {
      path = 0
     }
  }

  action {
    type = "regex"
    point = {
      action_name = "masking"
    }
  }

  action {
    type = "absent"
    point = {
      action_ext = ""
    }
  }

  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }

  action {
    type = "equal"
    value = "admin"
    point = {
      query = "user"
    }
  }

  # ... omitted
  ```

> **_NOTE:_**
See below what limitations apply

When `type` is `absent`, `point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string).



## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "enum"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_enum.wallarm_rule_enum_regexp 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.

[1]: https://docs.wallarm.com/api-protection/enumeration-attack-protection/#enumeration-attack-protection
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/