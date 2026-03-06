---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_parser_state"
subcategory: "Rule"
description: |-
  Provides the "Disable/Enable request parser" rule resource.
---

# wallarm_rule_parser_state

Provides the resource to manage rules with the "[Disable/Enable request parser][1]" action type. Allows disabling and enabling of parsers applied to the specified request point when analyzing it. By default, all parsers are applied to request points if a different configuration is not set in other rules.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

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
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `parser` - (**required**) parser to enable/disable. Possible values: `base64`, `cookie`, `form_urlencoded`, `gzip`, `grpc`, `json_doc`, `multipart`, `percent`, `protobuf`, `htmljs`, `viewstate`, `xml`.
* `state` - (**required**) desired state for the parser. Possible values: `enabled`, `disabled`.
* `point` - (**required**) request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

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
