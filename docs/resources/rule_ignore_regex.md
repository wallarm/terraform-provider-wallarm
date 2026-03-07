---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_ignore_regex"
subcategory: "Rule"
description: |-
  Provides the "Disable regexp-based attack detection" rule resource.
---

# wallarm_rule_ignore_regex

Provides the resource to manage rules with the "[Disable regexp-based attack detection][1]" action type. Ignoring the regular expression can be used when particular requests should NOT be defined as attacks based on the existing regular expression (the "Create regexp-based attack indicator" action type).

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
# Creates the rule to ignore an existing regular expression
# with ID 123 for the requests with the "X-LOGIN" header.

resource "wallarm_rule_ignore_regex" "ignore_regex" {
  regex_id = 123
  point = [["header", "X-LOGIN"]]
}
```

With newly created rule "Create regexp-based attack indicator":

```hcl
# Creates the rule to define requests with the "X-AUTHENTICATION" header value matching an expression
# "[^0-9a-f]|^.{33,}$|^.{0,31}$" as an attack

resource "wallarm_rule_regex" "scanner_rule" {
  regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
  experimental = true
  attack_type =  "scanner"
  point = [["header", "X-AUTHENTICATION"]]
}

# Creates the rule to ignore the regular expression above
# for the requests with the "X-AUTHENTICATION" header
# sent to the application with ID 5

resource "wallarm_rule_ignore_regex" "ignore_regex" {
  regex_id = wallarm_rule_regex.scanner_rule.regex_id

  action {
    point = {
      instance = 5
    }
  }

  point = [["header", "X-AUTHENTICATION"]]
  depends_on = [wallarm_rule_regex.scanner_rule]
}
```

## Argument Reference

* `regex_id` - (**required**) ID of the regular expression specified in the "Create regexp-based attack indicator" rule.
* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. Possible attributes are described in [action guide](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs/guides/action.md).
* `point` - (**required**) request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "ignore_regex"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_ignore_regex.ignore_regex 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_ignore_regex` - Terraform resource rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_ignore_regex" "ignore_regex" {
  action {
    point = {
      instance = 5
    }
  }
  point = [["uri"]]
  regex_id = 40671
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_ignore_regex.ignore_regex
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

[1]: https://docs.wallarm.com/user-guides/rules/regex-rule/#partial-disabling-of-a-new-detection-rule
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
