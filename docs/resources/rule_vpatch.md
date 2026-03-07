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
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `point` - (**required**) request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

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
