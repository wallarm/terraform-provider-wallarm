---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_credential_stuffing_regex"
subcategory: "Rule"
description: |-
  Provides the "Authentication endpoints by regular expression in Credential Stuffing" rule resource.
---

# wallarm_rule_credential_stuffing_regex

Provides the resource to configure authentication endpoints for [Credential Stuffing](https://docs.wallarm.com/about-wallarm/credential-stuffing/) by using regular expression approach.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_credential_stuffing_regex" "regex1" {
  regex = "*abc*"
  login_regex = "user*"
  case_sensitive = false
}

resource "wallarm_rule_credential_stuffing_regex" "regex2" {
  client_id = 123

  action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  regex = "*abc*"
  login_regex = "user*"
  case_sensitive = true
  cred_stuff_type = "custom"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `cred_stuff_type` - (optional) defines which database of compromised credentials to use. Can be: `default`, `custom`. Default value: `default`.
* `regex` - (**required**) regular expression used for specifying password parameters. Fore more details about regexps, see wallarm [documentation][1].
* `login_regex` - (**required**) regular expression used for specifying login parameters. Fore more details about regexps, see wallarm [documentation][1].
* `case_sensitive` - (**required**) defines whether regex and login_regex are case sensitive.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "credentials_regex"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_credential_stuffing_regex.regex2 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.
* `wallarm_rule_credential_stuffing_regex` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_credential_stuffing_regex" "regex2" {
  action {
    type = "iequal"
    point = {
      action_name = "login"
    }
  }
  regex           = "*abc*"
  login_regex     = "user*"
  case_sensitive  = true
  cred_stuff_type = "custom"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_credential_stuffing_regex.regex2
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

[1]: https://docs.wallarm.com/user-guides/rules/rules/#condition-type-regex
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
