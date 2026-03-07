---
layout: "wallarm"
page_title: "Wallarm: action"
description: |-
  Describes the action argument shared across Wallarm rule resources.
---

# action 

`action` argument shares the available conditions which can be applied. The conditions are:

* `type` - (optional) condition type. Can be: `equal`, `iequal`, `regex`, `absent`. For the `header` HOST condition, must always be `iequal`.
  For more details, see the official [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
* `value` - (optional) value of the parameter to match with. Required for `header` and `path` conditions.
* `point` - (optional) request parameters that trigger the rule. Possible values are described below. For more details, see the official [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

  | POINT | POSSIBLE VALUES | EXAMPLE |
  |---|---|---|
  | `header` | Arbitrary HEADER parameter name. | `header = "HOST"` |
  | `method` | `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`. | `method = "POST"` |
  | `path` | Integer >= 0. | `path = 0` |
  | `action_name` | Any string. | `action_name = "login"` |
  | `action_ext` | Any string. | `action_ext = "php"` |
  | `query` | Any string. | `query = "user"` |
  | `proto` | Any string. | `proto = "1.1"` |
  | `scheme` | `http`, `https`. | `scheme = "https"` |
  | `uri` | Any string. | `uri = "/api/login"` |
  | `instance` | Integer. Only `point` is required; `type` and `value` must be omitted. | `instance = 42` |

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

  action {
    type = "equal"
    value = "admin"
    point = {
      query = "user"
    }
  }

  # ... skipped
  ```

> **_NOTE:_**
See below what limitations apply

When `type` is `absent`, `point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string).
