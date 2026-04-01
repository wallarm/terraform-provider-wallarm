---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_credential_stuffing_regex"
subcategory: "Rules"
description: |-
  Provides the "Authentication endpoints by regular expression in Credential Stuffing" rule resource.
---

# wallarm_rule_credential_stuffing_regex

Provides the resource to configure authentication endpoints for [Credential Stuffing](https://docs.wallarm.com/about-wallarm/credential-stuffing/) by using regular expression approach.

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
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `cred_stuff_type` - (optional) defines which database of compromised credentials to use. Can be: `default`, `custom`. Default value: `default`.
* `regex` - (**required**) regular expression used for specifying password parameters. For more details about regexps, see wallarm [documentation][1].
* `login_regex` - (**required**) regular expression used for specifying login parameters. For more details about regexps, see wallarm [documentation][1].
* `case_sensitive` - (**required**) defines whether regex and login_regex are case sensitive.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "credentials_regex"`.

## Import

```
$ terraform import wallarm_rule_credential_stuffing_regex.regex1 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/user-guides/rules/rules/#condition-type-regex
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
