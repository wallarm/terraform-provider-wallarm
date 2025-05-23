---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_parser_state"
subcategory: "Rule"
description: |-
  Provides the "Disable/Enable request parser" rule resource.
---

# wallarm_rule_parser_state

Provides the resource to manage rules with the "[Disable/Enable request parser][1]" action type. Allows disabling and enabling of parsers applied to the specified request point when analyzing it. By default, all parsers are applied to request points if a different configuration is not set in other rules.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_parser_state" "disable_htmljs_parsing" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
  parser = "htmljs"
  state = "disabled"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described below.
* `parser` - (**required**) parser to enable/disable. Possible values: `base64`, `cookie`, `form_urlencoded`, `gzip`, `grpc`, `json_doc`, `multipart`, `percent`, `protobuf`, `htmljs`, `viewstate`, `xml`.
* `state` - (**required**) desired state for the parser. Possible values: `enabled`, `disabled`.
* `point` - (**required**) request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

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
* `rule_type` - type of the created rule. For example, `rule_type = "parser_state"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_parser_state.disable_htmljs_parsing 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_parser_state` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_parser_state" "disable_htmljs_parsing" {
  action {
    point = {
      header = "HOST"
    }
    type = "iequal"
    value = "example.com"
  }
  point = [["get_all"]]
  parser = "htmljs"
  state = "disabled"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_parser_state.disable_htmljs_parsing
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

[1]: https://docs.wallarm.com/user-guides/rules/disable-request-parsers/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
