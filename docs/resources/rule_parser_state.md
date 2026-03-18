---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_parser_state"
subcategory: "Rule"
description: |-
  Provides the "Disable/Enable request parser" rule resource.
---

# wallarm_rule_parser_state

Provides the resource to manage rules with the "[Disable/Enable request parser][1]" action type. Allows disabling and enabling of parsers applied to the specified request point when analyzing it. By default, all parsers are applied to request points if a different configuration is not set in other rules.

## Example Usage

```hcl
resource "wallarm_rule_parser_state" "disable_htmljs_parsing" {
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
  parser = "htmljs"
  state = "disabled"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `parser` - (**required**) parser to enable/disable. Possible values: `base64`, `cookie`, `form_urlencoded`, `gzip`, `grpc`, `json_doc`, `multipart`, `percent`, `protobuf`, `htmljs`, `viewstate`, `xml`.
* `state` - (**required**) desired state for the parser. Possible values: `enabled`, `disabled`.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "parser_state"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_parser_state.disable_htmljs_parsing 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_parser_state` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_parser_state" "disable_htmljs_parsing" {
  action {
    point = {
      header = "HOST"
    }
    type = "iequal"
    value = "example.com"
  }
  point = [["get_all"]]
  parser = "htmljs"
  state = "disabled"
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_parser_state.disable_htmljs_parsing
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

[1]: https://docs.wallarm.com/user-guides/rules/disable-request-parsers/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
