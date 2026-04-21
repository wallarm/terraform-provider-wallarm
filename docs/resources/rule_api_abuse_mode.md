---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_api_abuse_mode"
subcategory: "Rules"
description: |-
  Enables or disables API Abuse Prevention for requests matching an action scope.
---

# wallarm_rule_api_abuse_mode

Provides the resource to toggle [API Abuse Prevention][1] for traffic matching a given action scope. Typical use: allowlist trusted crawlers (Pinterest, Google, monitoring agents) by applying `mode = "disabled"` to requests that carry their signature (User-Agent, path, method, etc.).

## Example Usage

```hcl
# Disable API Abuse Prevention for legitimate Pinterest crawler traffic.

resource "wallarm_rule_api_abuse_mode" "pinterest" {
  mode    = "disabled"
  title   = "Allow Pinterest"
  comment = "Allow Pinterest through API Abuse Prevention"

  action {
    type  = "regex"
    value = ".*(Pinterest|Pinterestbot)/(0.2|1.0);?\\s[(]?[+]https?://www[.]pinterest[.]com/bot[.]html[)].*"
    point = {
      header = "USER-AGENT"
    }
  }

  action {
    type  = "equal"
    value = "api"
    point = {
      path = 0
    }
  }

  action {
    type  = "regex"
    value = "v\\d"
    point = {
      path = 1
    }
  }

  action {
    type = "absent"
    point = {
      action_ext = ""
    }
  }
}
```

## Argument Reference

* `mode` - (optional) API abuse mode. One of: `enabled`, `disabled`. Default: `enabled`. Changing this value destroys and recreates the rule.
* `title` - (optional) human-readable rule title.
* `comment` - (optional) free-text rule comment. Defaults to `Managed by Terraform`.
* `active` - (optional) whether the rule is active. Defaults to `true`.
* `client_id` - (optional) ID of the client to apply the rule to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (the conditions to apply on request).
* `rule_type` - type of the created rule. For `wallarm_rule_api_abuse_mode` this is always `api_abuse_mode`.

## Import

```
$ terraform import wallarm_rule_api_abuse_mode.pinterest 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/api-abuse-prevention/overview/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
