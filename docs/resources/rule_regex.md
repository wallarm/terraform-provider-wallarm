---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_regex"
subcategory: "Rule"
description: |-
  Provides the "Create regexp-based attack indicator" rule resource.
---

# wallarm_rule_regex

Provides the resource to manage rules with the "[Create regexp-based attack indicator][1]" action type. This rule type allows you to detect the specified attack based on the specified regular expression in the request.

The rule is generated based on the following parameters:

* **If request is**: conditions to trigger the action.

* **Regex**: regular expression denoting an attack. If the value of the following parameter matches the expression, that request is detected as an attack. Regular expressions syntax is described in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#regex).

* **Attack**: type of attack that will be detected when the parameter value in the request matches the regular expression. Possible values are described below.

* **Experimental**: flag to safely check the triggering of a regular expression without blocking requests. The requests won't be blocked even when the Wallarm node is set to the blocking mode. These requests will be considered as attacks detected by the experimental method. They can be accessed using search query `experimental attacks`.

* **In this part of request**: the point in the request where the specified attack should be detected.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
# Creates the rule to mark the requests sent to front.example.com
# with the URI value matching the regex ".*curltool.*" as
# non-experimental "vpatch" attacks

resource "wallarm_rule_regex" "regex_curltool" {
  regex = ".*curltool.*"
  experimental = false
  attack_type =  "vpatch"

  action {
    type = "iequal"
    value = "front.example.com"
    point = {
      header = "HOST"
    }
  }

  point = [["uri"]]
}


resource "wallarm_rule_regex" "scanner_rule" {
  regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
  experimental = true
  attack_type =  "scanner"
  action {
    point = {
      instance = 5
    }
  }
  point = [["header", "X-AUTHENTICATION"]]
}

```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `attack_type` - (**required**) attack type that will be detected when the parameter value in the request matches the regular expression. Can be: `any`, `sqli`, `rce`, `crlf`, `nosqli`, `ptrav`, `xxe`, `ptrav`, `xss`, `scanner`, `redir`, `ldapi`.
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.


## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "regex"`.
* `regex_id` - ID of the specified regular expression.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

ID should end with a rule type, which depends on experimental mode. The values are: `experimental_regex` or `regex` respectively.

```
$ terraform import wallarm_rule_regex.regex_curltool 6039/563855/11086881/regex
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.
* `wallarm_rule_regex` - Terraform resource rule type.
* `regex` - Rule type.

### Import blocks

The rule can be imported using Terraform import blocks.

Resource block example:

```hcl
resource "wallarm_rule_regex" "regex_curltool" {
  action {
    point = {
      header = "HOST"
    }
    type = "iequal"
    value = "front.example.com"
  }
  point = [["uri"]]
  attack_type = "vpatch"
  regex = ".*curltool.*"
  experimental = false
}
```

Import block example:

```hcl
import {
  to = wallarm_rule_regex.regex_curltool
  id = "6039/563855/11086881/regex"
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

[1]: https://docs.wallarm.com/user-guides/rules/regex-rule/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
