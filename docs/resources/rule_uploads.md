---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_uploads"
subcategory: "Rule"
description: |-
  Provides the "Allow certain file types" rule resource.
---

# wallarm_rule_uploads

Provides the resource to manage rules with the "[Allow certain file types][1]" action type. Allows fine-tuning attack detection for request points containing specific file types (e.g. PDF, JPG). When analyzing the specified request point, the Wallarm node will ignore attack signs that explicitly cannot be passed in the selected file types uploaded as binary data.

## Example Usage

```hcl
resource "wallarm_rule_uploads" "allow_markup_in_body" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"]]
  file_type = "html"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `file_type` - (**required**) file type to allow. Possible values: `docs`, `html`, `images`, `music`, `video`.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "uploads"`.

## Import

```
$ terraform import wallarm_rule_uploads.allow_markup_in_body 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/user-guides/rules/ignore-attacks-in-binary-data/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
