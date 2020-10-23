---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_regex"
subcategory: "Rule"
description: |-
  Provides the "Define a request as an attack based on a regular expression" rule resource.
---

# wallarm_rule_regex

Provides the resource to manage rules with the "Define a request as an attack based on a regular expression" action type. This rule type allows you to detect the specified attack based on the specified regular expression in the request.

The rule is generated based on the following parameters:

* **If request is**: conditions to trigger the action.

* **Regex**: the regular expression denoting an attack. If the value of the following parameter matches the expression, that request is detected as an attack. Regular expressions syntax is described in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#regex).

* **Attack**: the type of attack that will be detected when the parameter value in the request matches the regular expression. Possible values are described below.

* **Experimental**: the flag to safely check the triggering of a regular expression without blocking requests. The requests won't be blocked even when the WAF node is set to the blocking mode. These requests will be considered as attacks detected by the experimental method. They can be accessed using search query `experimental attacks`.

* **in this part of request**: the point in the request where the specified attack should be detected.

## Example Usage

```hcl

# Creates the rule to mark the requests sent to front.example.com
# with the URI value matching the regex ".*curltool.*" as
# non-experimental "vpatch" attacks

resource "wallarm_rule_regex" "regex_curltool" {
  regex = ".*curltool.*"
  experimental = false
  attack_type =  "vpatch"

  action {
    type = "iequal"
    value = "front.example.com"
    point = {
      header = "HOST"
    }
  }

  point = [["uri"]]
}


resource "wallarm_rule_regex" "scanner_rule" {
  regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
  experimental = true
  attack_type =  "scanner"
  action {
    point = {
      instance = 5
    }
  }
  point = [["header", "X-AUTHENTICATION"]]
}

```

## Argument Reference

* `client_id` - (Optional) ID of the client to apply the rules to. The value is required for multi-tenant scenarios.
* `attack_type` - (Required) Attack type that will be detected when the parameter value in the request matches the regular expression. Can be: `any`, `sqli`, `rce`, `crlf`, `nosqli`, `ptrav`, `xxe`, `ptrav`, `xss`, `scanner`, `redir`, `ldapi`.
* `action` - (Optional) Rule conditions. Possible attributes are described below.
* `point` - (Required) Request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).
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
  |`instance`      | Integer ID of the application the request was sent to. |

  [Examples](https://registry.terraform.io/providers/416e64726579/wallarm/latest/docs/guides/point)

`action` argument shares the available
conditions which can be applied. The conditions are:

* `type` - (Optional) Condition type. Can be: `equal`, `iequal`, `regex`, `absent`. Must be omitted for the `instance` parameter in `point`.
  For more details, see the offical [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
  Example:
  `type = "absent"`
* `value` - (Optional) Value of the parameter to match with. Must be omitted for the `instance` parameter in `point` or if `type` is `absent`.
  Example:
  `value = "example.com"`
* `point` - (Optional) Request parameters that trigger the rule. Possible values are described below. For more details, see the official [Wallarm documentatioon](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

**point**

  * `header` - (Optional) Arbitrary HEADER parameter name.
  Example:
  `header = "HOST"`
  * `method` - (Optional) Request method. Can be: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
  Example:
  `method = "POST"`
  * `path` - (Optional) Array with URL parts separated by the `/` symbol (the last URL part is not included in the array). If there is only one part in the URL, the array will be empty.
  Example:
  `path = 0`
  * `action_name` - (Optional) The last part of the URL after the `/` symbol and before the first period (`.`). This part of the URL is always present in the request even if its value is an empty string.
  Example:
  `action_name = "login"`
  * `action_ext` - (Optional) The part of the URL after the last period (`.`). It may be missing in the request.
  Example:
  `action_ext = "php"`
  * `proto` - (Optional) Version of the HTTP Protocol.
  Example:
  `proto = "1.1"`
  * `scheme` - (Optional) `http`/`https`.
  Example:
  `scheme = "https"` 
  * `uri` - (Optional) Part of the request URL without domain.
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

When `type` is `absent`
`point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string)

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - The action ID (The conditions to apply on request).
* `rule_type` - Type of the created rule. For example, `rule_type = "ignore_regex"`.
* `regex_id` - ID of the specified regular expression.
