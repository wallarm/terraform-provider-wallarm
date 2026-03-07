---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_binary_data"
subcategory: "Rule"
description: |-
  Provides the "Allow binary data" rule resource.
---

# wallarm_rule_binary_data

Provides the resource to manage rules with the "[Allow binary data][1]" action type. Allows fine-tuning attack detection for request points containing binary data (e.g. archived or encrypted files). When analyzing the specified request point, the Wallarm node will ignore attack signs that cannot be explicitly passed in binary data.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
resource "wallarm_rule_binary_data" "allow_bin_in_body" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"]]
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `point` - (**required**) request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "binary_data"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_binary_data.allow_bin_in_body 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_binary_data` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_binary_data" "allow_bin_in_body" {
  action {
    point = {
      header = "HOST"
    }
    type = "iequal"
    value = "example.com"
  }
  point = [["post"]]
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_binary_data.allow_bin_in_body
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

[1]: https://docs.wallarm.com/user-guides/rules/ignore-attacks-in-binary-data/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
