---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_disable_attack_type"
subcategory: "Rule"
description: |-
  Provides the "Ignore certain attack types" rule resource.
---

# wallarm_rule_disable_attack_type

Provides the resource to manage rules with the "[Ignore certain attack types][1]" action type. Disables detection of all specified attack type stamps for selected request points.

## Example Usage

```hcl
resource "wallarm_rule_disable_attack_type" "disable_sqli" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
  attack_type = "sqli"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `attack_type` - (**required**) attack type to ignore. Possible values: `sqli`, `xss`, `rce`, `ptrav`, `crlf`, `nosqli`, `xxe`, `ldapi`, `scanner`, `ssti`, `ssi`, `mail_injection`, `vpatch`.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "disable_attack_type"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_disable_attack_type.disable_sqli 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_disable_attack_type` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_disable_attack_type" "disable_sqli" {
  action {
    point = {
      header = "HOST"
    }
    type = "iequal"
    value = "example.com"
  }
  point = [["get_all"]]
  attack_type = "sqli"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_disable_attack_type.disable_sqli
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

[1]: https://docs.wallarm.com/user-guides/rules/ignore-attack-types/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
