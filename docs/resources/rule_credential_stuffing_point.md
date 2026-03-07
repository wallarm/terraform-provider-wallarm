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
  login_point = [["HEADER", "SESSION-ID"]]
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
  login_point = [["HEADER", "SESSION-ID"]]
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
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

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "credentials_point"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_credential_stuffing_point.point2 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.
* `wallarm_rule_credential_stuffing_point` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_credential_stuffing_point" "point2" {
  action {
    type = "iequal"
    point = {
      action_name = "login"
    }
  }
  point       = [["HEADER", "HOST"]]
  login_point = ["HEADER", "SESSION-ID"]
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_credential_stuffing_point.point2
  id = "6039/563854/11086884"
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

[1]: https://docs.wallarm.com/user-guides/rules/rules/#points
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
