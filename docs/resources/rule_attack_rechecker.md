---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_attack_rechecker"
subcategory: "Rule"
description: |-
  Provides the "Disable/Enable attack verification" rule resource.
---

# wallarm_rule_attack_rechecker

!> The resource requires additional permissions. Ask the support team to obtain them.

Provides the resource to manage rules with the "Disable/Enable attack verification" action type. This rule type is used to enable or disable checking attacks in specified request parts.

## Example Usage

```hcl
# Disables the attacks checking for requests sent to the application with ID 7

resource "wallarm_rule_attack_rechecker" "disable_rechecker" {
  enabled =  false

  action {
    point = {
      instance = 7
    }
  }
}

```

## Argument Reference

* `enabled` - (**required**) indicator to enable (`true`) or disable (`false`) attacks checking.
* `action` - (optional) rule conditions. Possible attributes are described below.

**action**

`action` argument shares the available
conditions which can be applied. The conditions are:

* `type` - (optional) condition type. Can be: `equal`, `iequal`, `regex`, `absent`. Must be omitted for the `instance` parameter in `point`.
  For more details, see the offical [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
  Example:
  `type = "absent"`
* `value` - (optional) value of the parameter to match with. Must be omitted for the `instance` parameter in `point` or if `type` is `absent`.
  Example:
  `value = "example.com"`
* `point` - (optional) request parameters that trigger the rule. Possible values are described below. For more details, see the official [Wallarm documentatioon](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

### Nested Objects

* `point`

The **point** attribute supports the following fields:
  * `header` - (optional) a header name. It requres arbitrary value for the parameter.
  Example:
  `header = "HOST"`
  * `method` - (optional) an HTTP method. It requires one of the values: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`
  Example:
  `method = "POST"`
  * `path` - (optional) a part of the request URI.
  Example:
  `path = 0`
  * `action_name` - (optional) the last part of the URL after the `/` symbol and before the first period `.`. This part of the URL is always present in the request even if its value is an empty string.
  Example:
  `action_name = "login"`
  * `action_ext` - (optional) the part of the URL after the last period `.`. It may be missing in the request.
  Example:
  `action_ext = "php"`
  * `proto` - (optional) version of the HTTP Protocol.
  Example:
  `proto = "1.1"`
  * `scheme` - (optional) http/https.
  Example:
  `scheme = "https"` 
  * `uri` - (optional) String with the original URL value.
  Example:
  `uri = "/api/login"` 
  * `instance` - (optional) ID of the application.
  Example:
  `instance = 42`

Example:

  ```hcl
  # ... other configuration

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

  # ... skipped
  ```

> **_NOTE:_**
See below what limitations apply

`type` must be omitted when:
- `point` is made up for `instance`

`value` must be omitted when: 
- `type` is `absent`
- `point` is made up for `instance`

When `type` is `absent`
`point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string)

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of created rule. For example, `rule_type = "ignore_regex"`.
