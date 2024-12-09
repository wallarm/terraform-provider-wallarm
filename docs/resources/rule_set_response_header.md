---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_set_response_header"
subcategory: "Rule"
description: |-
  Provides the "Change server response headers" rule resource.
---

# wallarm_rule_set_response_header

Provides the resource to manage rules with the "[Change server response headers][1]" action type. This rule type is used for adding or deleting server response headers and changing their values.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
# Append the "Server" header with the "Wallarm solution" value
# and the "Server" header with the "Blocked by Wallarm" value
# to the requests sent to the application with ID 3

resource "wallarm_rule_set_response_header" "resp_header" {
  mode = "append"

  action {
    point = {
      instance = 3
    }
  }

  name = "Server"
  values = ["Wallarm solution", "Blocked by Wallarm"]
}

```

```hcl
# Deletes the "Wallarm" header

resource "wallarm_rule_set_response_header" "delete_header" {
  mode = "replace"
  name =  "Wallarm"
  values = [""]
}

```

## Argument Reference

* `mode` - (**required**) mode of header processing. Valid options: `append`, `replace`
* `name` - (**required**) header name.
* `values` - (**required**) array of header values. Might be defined as much values as need at once.
* `action` - (optional) a series of conditions, see below for a
  a full list .

**action**

`action` argument shares the available conditions which can be applied. The conditions are:

* `type` - (optional) the type of comparison. Possible values: `equal`, `iequal`, `regex`, `absent`.
  For more information, see the [docs](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
  Example:
  `type = "absent"`
* `value` - (optional) a value of the parameter to match with.
  Example:
  `value = "example.com"`
* `point` - (optional) a series of arguments, see below for a a full list . See the [docs](https://docs.wallarm.com/user-guides/rules/request-processing/#parameter-parsing).

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
  * `query` - (optional) the query parameter name.
  Example:
  `query = "user"`
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

`type` must be omitted when:
- `point` is made up for `instance`

`value` must be omitted when:
- `type` is `absent`
- `point` is made up for `instance`

When `type` is `absent`, `point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of created rule. For example, `rule_type = "set_response_header"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_set_response_header.resp_header 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_set_response_header` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_set_response_header" "resp_header" {
  action {
    point = {
      instance = 3
    }
  }
  mode = "append"
  name = "Server"
  values = ["Blocked by Wallarm","Wallarm solution"]
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_set_response_header.resp_header
  id = "6039/563855/11086881"
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

[1]: https://docs.wallarm.com/user-guides/rules/add-replace-response-header/
