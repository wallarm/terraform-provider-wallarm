---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_masking"
subcategory: "Rule"
description: |-
  Provides the "Mask sensitive data" rule resource.
---

# wallarm_rule_masking

Provides the resource to manage rules with the "[Mask sensitive data][1]" action type. This rule type is used to cut out sensitive information such as passwords or cookies from the uploading to the Wallarm Cloud making such data hidden.

The real values of the specified parameters will be replaced by `*` and will not be accessible either in the Wallarm Cloud or in the local post-analysis module. This method ensures that the protected data cannot leak outside the trusted environment.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
# Masks the "field" value of the "hash" parameter
# in the JSON body for the requests sent to the `../masking` URL

resource "wallarm_rule_masking" "masking_json" {

  action {
    type = "equal"
    point = {
      action_name = "masking"
    }
  }

  action {
    type = "absent"
    point = {
      path = 0
     }
  }

  action {
    type = "absent"
    point = {
      action_ext = ""
    }
  }

  action {
    type = "equal"
    value = "admin"
    point = {
      query = "user"
    }
  }

  point = [["post"], ["json_doc"], ["hash", "field"]]
}

```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `point` - (**required**) request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of created rule. For example, `rule_type = "sensitive_data"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_masking.masking_header 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_masking` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_masking" "masking_header" {
  action {
    point = {
      instance = 1
    }
  }
  point = [["header","AUTHORIZATION"]]
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_masking.masking_header
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

[1]: https://docs.wallarm.com/user-guides/rules/sensitive-data-rule/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
