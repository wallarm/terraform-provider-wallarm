---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_file_upload_size_limit"
subcategory: "Rules"
description: |-
  Provides the "File upload restriction policy" rule resource.
---

# wallarm_rule_file_upload_size_limit

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
* `size` - (**required**) maximum allowed size of uploading data (range `1..2^64`; 0 invalid).
* `point` - (optional, computed, force-new) request parts to apply the rules to. See the [Point Guide](../guides/point). API default scope applies when omitted.
* `size_unit` - (optional, computed, force-new) `b`, `kb`, `mb`, `gb`, or `tb`. API default `b` applies when omitted.
* `mode` - (optional, computed) `monitoring`, `block`, `off`, or `default`. API default `monitoring` applies when omitted.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `mitigation` - type of the created mitigation. For example, `mitigation = "file_upload_policy"`
* `rule_type` - type of the created rule. For example, `rule_type = "file_upload_size_limit"`.

## Import

```
$ terraform import wallarm_rule_file_upload_size_limit.file_upload_restriction 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

This resource is a **mitigation control**. For an overview of all mitigation controls and their parameter mapping, see the [Mitigation Controls Guide](../guides/mitigation_controls).

[1]: https://docs.wallarm.com/api-protection/file-upload-restriction/#rule-based-protection
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/