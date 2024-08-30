---
layout: "wallarm"
page_title: "Wallarm: wallarm_rate_limit"
subcategory: "Rule"
description: |-
  Provides the "Rate Limit" rule resource.
---

# wallarm_rate_limit

Wallarm provides the Set rate limit rule to help prevent excessive traffic to your API. This rule enables you to specify the maximum number of connections that can be made to a particular scope, while also ensuring that incoming requests are evenly distributed. If a request exceeds the defined limit, Wallarm rejects it and returns the code you selected in the rule.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex). This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rate_limit" "example" {
  comment    = "Example rate limit rule"

  action = {
    type  = "equal"
    value = "example_value"
    point = {
      header       = ["X-Example-Header"]
      method       = "GET"
      path         = 10
      action_name  = "example_action"
      action_ext   = "example_extension"
      query        = "example_query"
      proto        = "HTTP/1.1"
      scheme       = "https"
      uri          = "/example_uri"
      instance     = 1
    }
  }

  point = ["example_point_1", "example_point_2"]

  delay      = 100
  burst      = 200
  rate       = 300
  rsp_status = 404
  time_unit  = "rps"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][1].
* `delay` - (required) Specifies the delay.
* `burst` - (required) Specifies the burst size.
* `rate` - (required) Specifies the rate limit.
* `rsp_status` - (optional) Specifies the response status code when the rate limit is exceeded.
* `time_unit` - (required) Specifies the time unit for rate limiting. Can be rps (requests per second) or rpm (requests per minute).
* `action` - (optional) rule conditions. Possible attributes are described below.

**action**

`action` argument shares the available conditions which can be applied. The conditions are:

* `type` - (optional) condition type. Can be: `equal`, `iequal`, `regex`, `absent`. Must be omitted for the `instance` parameter in `point`.
  For more details, see the official [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
  Example:
  `type = "absent"`
* `value` - (optional) value of the parameter to match with. Must be omitted for the `instance` parameter in `point` or if `type` is `absent`.
  Example:
  `value = "example.com"`
* `point` - (optional) request parameters that trigger the rule. Possible values are described below. For more details, see the official [Wallarm documentatioon](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

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
* `counter` - Name of the counter. Randomly generated, but always starts with `d:`.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "dirbust_counter"`.


[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
