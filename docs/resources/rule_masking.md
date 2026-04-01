---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_masking"
subcategory: "Rules"
description: |-
  Provides the "Mask sensitive data" rule resource.
---

# wallarm_rule_masking

Provides the resource to manage rules with the "[Mask sensitive data][1]" action type. This rule type is used to cut out sensitive information such as passwords or cookies from the uploading to the Wallarm Cloud making such data hidden.

The real values of the specified parameters will be replaced by `*` and will not be accessible either in the Wallarm Cloud or in the local post-analysis module. This method ensures that the protected data cannot leak outside the trusted environment.

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
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of created rule. For example, `rule_type = "sensitive_data"`.

## Import

```
$ terraform import wallarm_rule_masking.masking_header 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/user-guides/rules/sensitive-data-rule/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
