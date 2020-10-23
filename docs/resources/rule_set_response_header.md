---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_set_response_header"
subcategory: "Rule"
description: |-
  Provides the "Specify a response header" rule resource.
---

# wallarm_rule_set_response_header

*THIS RESOURCE REQUIRES ADDITIONAL PERMISSIONS!*

Provides the resource to manage rules with the "Specify a response header" action type. This rule type is used to configure supplementary headers appended or replaced by a Wallarm WAF node.

## Example Usage

```hcl
# Append the "Server" header with the "Wallarm WAF" value
# and the "Blocked" header with the "Blocked by Wallarm" value
# to the requests sent to the application with ID 3

resource "wallarm_rule_set_response_header" "resp_headers" {
  mode = "append"

  action {
    point = {
      instance = 3
    }
  }

  headers = {
    Server = "Wallarm WAF"
    Blocked = "Blocked by Wallarm"
  }

}

```

## Argument Reference

* `mode` - (Required) Mode of header processing. Valid options: `append`, `replace`
* `headers` - (Required) The associative array of key/value headers. Might be defined as much headers as need at once. 
* `action` - (Optional) A series of conditions, see below for a
  a full list .

**action**

`action` argument shares the available
conditions which can be applied. The conditions are:

* `type` - (Optional) The type of comparison. Possible values: `equal`, `iequal`, `regex`, `absent`.
  For more information, see the [docs](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
  Example:
  `type = "absent"`
* `value` - (Optional) A value of the parameter to match with.
  Example:
  `value = "example.com"`
* `point` - (Optional) A series of arguments, see below for a a full list . See the [docs](https://docs.wallarm.com/user-guides/rules/request-processing/#parameter-parsing).

### Nested Objects

* `point`

The **point** attribute supports the following fields:
  * `header` - (Optional) A header name. It requres arbitrary value for the parameter.
  Example:
  `header = "HOST"`
  * `method` - (Optional) An HTTP method. It requires one of the values: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`
  Example:
  `method = "POST"`
  * `path` - (Optional) A part of the request URI.
  Example:
  `path = 0`
  * `action_name` - (Optional) The last part of the URL after the `/` symbol and before the first period `.`. This part of the URL is always present in the request even if its value is an empty string.
  Example:
  `action_name = "login"`
  * `action_ext` - (Optional) The part of the URL after the last period `.`. It may be missing in the request.
  Example:
  `action_ext = "php"`
  * `proto` - (Optional) Version of the HTTP Protocol.
  Example:
  `proto = "1.1"`
  * `scheme` - (Optional) http/https.
  Example:
  `scheme = "https"` 
  * `uri` - (Optional) String with the original URL value.
  Example:
  `uri = "/api/login"` 
  * `instance` - (Optional) ID of the application.
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
* `action_id` - The action ID (The conditions to apply on request).
* `rule_type` - Type of   created rule. For example, `rule_type = "ignore_regex"`.