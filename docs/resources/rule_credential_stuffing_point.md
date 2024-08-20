---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_credential_stuffing_point"
subcategory: "Rule"
description: |-
  Provides the "Authentication endpoints by exact location of parameters in Credential Stuffing" rule resource.
---

# wallarm_rule_credential_stuffing_point

Provides the resource to configure authentication endpoints for [Credential Stuffing](https://docs.wallarm.com/about-wallarm/credential-stuffing/) by using request point approach.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_credential_stuffing_point" "point1" {
  point = [["HEADER", "HOST"]]
  login_point = ["HEADER", "SESSION-ID"]
}

resource "wallarm_rule_credential_stuffing_point" "point2" {
  client_id = 123

  action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  point = [["HEADER", "HOST"]]
  login_point = ["HEADER", "SESSION-ID"]
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described below.
* `cred_stuff_type` - (optional) defines which database of compromised credentials to use. Can be: `default`, `custom`. Default value: `default`.
* `point` - (**required**) request point used for specifying password parameters. Fore more details about request points, see wallarm [documentation][1].
* `login_point` - (**required**) request point used for specifying login parameters. Fore more details about request points, see wallarm [documentation][1].

**point**, **login_point**

Should be a correct point belonging to the request, that finishes by _all

Example:

Correct:

* [["post"],["form_urlencoded", "test"],["array_all"]]
* [["post"],["form_urlencoded_all"]]
* [["post"],["json_doc"],["array_all"]]
* [["header_all"]]

Incorrect:

* [["post"],["form_urlencoded", "test"]]
* [["post"]]
* [["path_all"]]
* [["header","HOST"]]

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

**action.point**

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
* `rule_type` - type of the created rule. For example, `rule_type = "credentials_point"`.

[1]: https://docs.wallarm.com/user-guides/rules/rules/#points
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
