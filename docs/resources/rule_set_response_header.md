---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_set_response_header"
subcategory: "Rule"
description: |-
  Provides the "Change server response headers" rule resource.
---

# wallarm_rule_set_response_header

Provides the resource to manage rules with the "[Change server response headers][1]" action type. This rule type is used for adding or deleting server response headers and changing their values.

## Example Usage

```hcl
# Append the "Server" header with the "Wallarm solution" value
# and the "Server" header with the "Blocked by Wallarm" value
# to the requests sent to the application with ID 3

resource "wallarm_rule_set_response_header" "resp_header" {
  mode = "append"

  action {
    point = {
      instance = 3
    }
  }

  name = "Server"
  values = ["Wallarm solution", "Blocked by Wallarm"]
}

```

```hcl
# Deletes the "Wallarm" header

resource "wallarm_rule_set_response_header" "delete_header" {
  mode = "replace"
  name =  "Wallarm"
  values = [""]
}

```

## Argument Reference

* `mode` - (**required**) mode of header processing. Valid options: `append`, `replace`
* `name` - (**required**) header name.
* `values` - (**required**) array of header values. Might be defined as much values as need at once.
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of created rule. For example, `rule_type = "set_response_header"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_set_response_header.resp_header 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_set_response_header` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_set_response_header" "resp_header" {
  action {
    point = {
      instance = 3
    }
  }
  mode = "append"
  name = "Server"
  values = ["Blocked by Wallarm","Wallarm solution"]
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_set_response_header.resp_header
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

[1]: https://docs.wallarm.com/user-guides/rules/add-replace-response-header/
