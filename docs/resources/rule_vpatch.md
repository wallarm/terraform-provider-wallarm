---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_vpatch"
subcategory: "Rule"
description: |-
  Provides the "Create a virtual patch" rule resource.
---

# wallarm_rule_vpatch

Provides the resource to manage rules with the "[Create a virtual patch][1]" action type. This rule type allows you to block malicious requests if the Wallarm node is working in the `monitoring` mode or if any known malicious payload is not detected in the request but this request must be blocked.

Virtual patches are especially useful in cases when it is impossible to fix a critical vulnerability in the code or install the necessary security updates quickly.

If attack types are specified, the request will be blocked only if the Wallarm node detects an attack of one of the listed types in the corresponding parameter. If the attack type is set to `any`, the Wallarm node blocks the requests with the defined parameter, even if it does not contain a malicious payload.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
# Creates the rule to block incoming requests with the "HOST" header
# containing the SQL Injection
# in any GET parameter

resource "wallarm_rule_vpatch" "splunk" {
  attack_type = "sqli"

  action {
    type = "iequal"
    value = "app.example.com"

    point = {
      header = "HOST"
    }

  }

  point = [["get_all"]]
}

```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `attack_type` - (**required**) attack type. The request with this attack will be blocked. Can be:
  * `any` to block the request with the specified `point` even if the attack is not detected.
  * One of the names of attack types to block the requests with the specified `point` if these malicious payloads are detected. Possible attack types: `sqli`, `rce`, `crlf`, `nosqli`, `ptrav`, `xxe`, `ptrav`, `xss`, `scanner`, `redir`, `ldapi`.
* `action` - (optional) rule conditions. Possible attributes are described below.
* `point` - (**required**) request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).
  |     POINT      |POSSIBLE VALUES|
  |----------------|---------------|
  |`action_ext`    |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
  |`action_name`   |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
  |`get`           | Arbitrary GET parameter name.|
  |`get_all`       |`array`, `array_all`, `array_default`, `base64`, `gzip`, `json_doc`, `xml`, `hash`, `hash_all`, `hash_default`, `hash_name`, `htmljs`, `pollution`|
  |`get_default`   |`array`, `array_all`, `array_default`, `base64`, `gzip`, `json_doc`, `xml`, `hash`, `hash_all`, `hash_default`, `hash_name`, `htmljs`, `pollution`|
  |`get_name`      |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
  |`header`        | Arbitrary HEADER parameter name.|
  |`header_all`    |`array`, `array_all`, `array_default`, `base64`, `cookie`, `cookie_all`, `cookie_default`, `cookie_name`, `gzip`, `json_doc`, `xml`, `hash`, `htmljs`, `pollution`|
  |`header_default`|`array`, `array_all`, `array_default`, `base64`, `cookie`, `cookie_all`, `cookie_default`, `cookie_name`, `gzip`, `json_doc`, `xml`, `hash`, `htmljs`, `pollution`|
  |`path`          | Integer value (>= 0) indicating the number of the element in the path array. |
  |`path_all`      |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
  |`path_default`  |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
  |`post`          |`base64`, `form_urlencoded`, `form_urlencoded_all`, `form_urlencoded_default`, `form_urlencoded_name`, `grpc`, `grpc_all`, `grpc_default`, `gzip`, `htmljs`, `json_doc`, `multipart`, `multipart_all`, `multipart_default`, `multipart_name`, `xml`|
  |`uri`           |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`, `percent`|
  |`json_doc`   |`array`, `array_all`, `array_default`, `hash`, `hash_all`, `hash_default`, `hash_name`, `json_array`, `json_array_all`, `json_array_default`, `json_obj`, `json_obj_all`, `json_obj_default`, `json_obj_name`|
  |`instance`      | integer ID of the application the request was sent to. |

  [Examples](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/point)

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
  # ... omitted configurations

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

  # ... omitted configurations
  ```

> **_NOTE:_**
See below what limitations apply

When `type` is `absent`, `point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "vpatch"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_vpatch.vpatch_test 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_vpatch` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_vpatch" "vpatch_test" {
  action {
    point = {
      header = "HOST"
    }
    type = "iequal"
    value = "app.example.com"
  }
  point = [["get_all"]]
  attack_type = "sqli"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_vpatch.vpatch_test
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

[1]: https://docs.wallarm.com/user-guides/rules/vpatch-rule/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
