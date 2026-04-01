---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_rate_limit"
subcategory: "Rule"
description: |-
  Provides the "Rate Limit" rule resource.
---

# wallarm_rule_rate_limit

Wallarm provides the Set rate limit rule to help prevent excessive traffic to your API. This rule enables you to specify the maximum number of connections that can be made to a particular scope, while also ensuring that incoming requests are evenly distributed. If a request exceeds the defined limit, Wallarm rejects it and returns the code you selected in the rule.

## Example Usage

```hcl
resource "wallarm_rule_rate_limit" "rate_limit_api" {
  action {
    type = "equal"
    value = "api"
    point = {
      path = 0
    }
  }
  action {
    point = {
      instance = 1
    }
  }

  point = [["post"], ["json_doc"], ["hash", "email"]]

  delay      = 100
  burst      = 200
  rate       = 300
  rsp_status = 404
  time_unit  = "rps"
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][1].
* `delay` - (required) Specifies the delay.
* `burst` - (required) Specifies the burst size.
* `rate` - (required) Specifies the rate limit.
* `rsp_status` - (optional) Specifies the response status code when the rate limit is exceeded.
* `time_unit` - (required) Specifies the time unit for rate limiting. Can be rps (requests per second) or rpm (requests per minute).
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "rate_limit"`.

## Import

```
$ terraform import wallarm_rule_rate_limit.rate_limit_api 6039/563855/11086881
```

* `6039` - Client ID.
* `563855` - Action ID.
* `11086881` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
