---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_bruteforce_counter"
subcategory: "Rule"
description: |-
  Provides the "Define brute-force attacks counter" rule resource.
---

# wallarm_rule_bruteforce_counter

Provides the resource to manage rules with the "Define brute-force attacks counter" action type. For detecting brute-force attacks, with every request, one of the statistical counters is incremented. By default, the counter name is automatically defined.

The counter works in conjunction with `wallarm_trigger` to detect and mitigate brute-force attacks by tracking request patterns to authentication and other endpoints.

## Example Usage

```hcl
# Sets a counter on the root `/` path

resource "wallarm_rule_bruteforce_counter" "root_counter" {
	action {
		type = "iequal"
		value = "/"
		point = {
			path = 0
		}
	}

}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `counter` - Name of the counter. Randomly generated, but always starts with `b:`.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "brute_counter"`.

## Import

```
$ terraform import wallarm_rule_bruteforce_counter.root_counter 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
