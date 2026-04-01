---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_disable_attack_type"
subcategory: "Rules"
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

```
$ terraform import wallarm_rule_disable_attack_type.disable_sqli 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/user-guides/rules/ignore-attack-types/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
