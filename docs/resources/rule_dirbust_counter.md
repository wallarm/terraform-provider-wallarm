---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_dirbust_counter"
subcategory: "Rules"
description: |-
  Provides the "Define force browsing attacks counter" rule resource.
---

# wallarm_rule_dirbust_counter

Provides the resource to manage rules with the "Define force browsing attacks counter" action type. For detecting force browsing attacks, there is a counter that increments whenever a request hits 404 status code (resource not found). By default, every application has its own counter.

The counter works in conjunction with `wallarm_trigger` to detect and mitigate force browsing attacks by tracking request patterns for different parts of the application.

## Example Usage

```hcl
resource "wallarm_rule_dirbust_counter" "login_counter" {
	action {
    	type = "iequal"
    	point = {
      		action_name = "login"
    	}
  	}
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `counter` - Name of the counter. Randomly generated, but always starts with `d:`.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "dirbust_counter"`.

## Import

```
$ terraform import wallarm_rule_dirbust_counter.login_counter 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
