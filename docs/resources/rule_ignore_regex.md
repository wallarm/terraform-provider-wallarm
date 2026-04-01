---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_ignore_regex"
subcategory: "Rules"
description: |-
  Provides the "Disable regexp-based attack detection" rule resource.
---

# wallarm_rule_ignore_regex

Provides the resource to manage rules with the "[Disable regexp-based attack detection][1]" action type. Ignoring the regular expression can be used when particular requests should NOT be defined as attacks based on the existing regular expression (the "Create regexp-based attack indicator" action type).

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
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.


## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "ignore_regex"`.

## Import

```
$ terraform import wallarm_rule_ignore_regex.ignore_regex 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/user-guides/rules/regex-rule/#partial-disabling-of-a-new-detection-rule
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
