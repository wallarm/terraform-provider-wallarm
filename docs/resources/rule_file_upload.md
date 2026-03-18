---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_file_upload"
subcategory: "Rule"
description: |-
  Provides the "File upload restriction policy" rule resource.
---

# wallarm_rule_file_upload

Provides the resource to manage mitigation control with the "[File upload restriction policy][1]" action type. This control enforces strict limits on the total request size and/or the size of individual parameters (such as specific file upload fields or JSON payload elements). Additionally, you can configure this rule to limit the maximum size of any header. This capability reduces an attacker's potential to inject payloads or exploit Buffer Overflow vulnerabilities.

## Example Usage

```hcl
resource "wallarm_rule_file_upload_size_limit" "file_upload_restriction" {
  mode = "block"

  action {
    type  = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }

  point = [["post"],["multipart", "file"]]

  size      = 10
  size_unit = "mb"

}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.
* `size` - (**required**) maximum allowed size of uploading data.
* `size_unit` - (**required**) dimension of uploading data. Possible values (`b`, `kb`, `mb`, `gb`, `tb`).
* `mode` - (**required**) protection behaviour which will be applied to the detected attack. Possible values: `monitoring`, `block`, `off`, `default`.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `mitigation` - type of the created mitigation. For example, `mitigation = "file_upload_policy"`
* `rule_type` - type of the created rule. For example, `rule_type = "file_upload_size_limit"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_file_upload_size_limit.file_upload_restriction 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.
* `wallarm_rule_file_upload_size_limit` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_file_upload_size_limit" "file_upload_restriction" {
  action {
    type  = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point     = [["post"], ["multipart", "file"]]
  size      = 10
  size_unit = "mb"
  mode      = "block"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_file_upload_size_limit.file_upload_restriction
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

[1]: https://docs.wallarm.com/api-protection/file-upload-restriction/#rule-based-protection
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/