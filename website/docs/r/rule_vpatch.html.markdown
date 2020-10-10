---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_vpatch"
sidebar_current: "docs-wallarm-resource-rule-vpatch"
description: |-
  Provides the "Create a virtual patch" rule resource.
---

# wallarm_rule_vpatch

Provides the resource to manage rules with the "Create a virtual patch" action type. This rule type allows you to block malicious requests if the WAF node is working in the `monitoring` mode or if any known attack vector is not detected in the request but this request must be blocked.

Virtual patches are especially useful in cases when it is impossible to fix a critical vulnerability in the code or install the necessary security updates quickly.

If attack types are specified, the request will be blocked only if the WAF node detects an attack of one of the listed types in the corresponding parameter. If the setting **Any request** is specified, the WAF node blocks the requests with the defined parameter, even if it does not contain an attack vector.

## Example Usage

```hcl
# Creates the rule to block incoming requests
# containing the SQL Injection
# in the "query" GET parameter

resource "wallarm_rule_vpatch" "default" {
  attack_type =  ["sqli"]
  point = [["get", "query"]]
}

# Creates the rule to block incoming requests with the "HOST" header
# containing the SQL Injection or NoSQL Injection
# in any GET parameter

resource "wallarm_rule_vpatch" "splunk" {
  attack_type =  ["sqli", "nosqli"]

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

* `client_id` - (Optional) ID of the client to apply the rules to. The value is required for multi-tenant scenarios.
* `attack_type` - (Required) Attack type. The request with this attack will be blocked. Can be:
  * `any` to block the request with the specified `point` even if the attack is not detected.
  * One more names of attack types to block the requests with the specified `point` if these attack vectors are detected. Possible attack types: `sqli`, `rce`, `crlf`, `nosqli`, `ptrav`, `xxe`, `ptrav`, `xss`, `scanner`, `redir`, `ldapi`.
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

  [Examples](https://registry.terraform.io/providers/416e64726579/wallarm/latest/docs/examples/point)

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

  # ... omitted configurations
  ```

> **_NOTE:_**
See below what limitations apply

When `type` is `absent`
`point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string)

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - The action ID (The conditions to apply on request).
* `rule_type` - Type of the created rule. For example, `rule_type = "ignore_regex"`.