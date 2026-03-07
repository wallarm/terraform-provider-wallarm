---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_credential_stuffing_mode"
subcategory: "Rule"
description: |-
  Provides "Credential Stuffing" rule resource.
---

# wallarm_rule_credential_stuffing_mode

Provides the resource to enable and disable credentials stuffing feature for specific endpoints.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_credential_stuffing_point" "mode1" {

}

resource "wallarm_rule_credential_stuffing_point" "mode2" {
  client_id = 123

  action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  mode = "custom"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `mode` - (optional) credential stuffing mode. Can be: `default`, `custom`, `disabled`. Default value: `default`.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "cred_stuff_mode"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_credential_stuffing_mode.mode2 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.
* `wallarm_rule_credential_stuffing_mode` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_credential_stuffing_mode" "mode2" {
  action {
    type = "iequal"
    point = {
      action_name = "login"
    }
  }
  mode = "custom"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_credential_stuffing_mode.mode2
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

[1]: https://docs.wallarm.com/about-wallarm/credential-stuffing/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
