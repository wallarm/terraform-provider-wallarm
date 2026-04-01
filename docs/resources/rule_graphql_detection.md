---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_graphql_detection"
subcategory: "Rules"
description: |-
  Provides the "Enumeration attack protection" mitigation control resource.
---

# wallarm_rule_graphql_detection

Provides the resource to manage mitigation control with the "[GraphQL API protection][1]" action type. They contain generic configuration to detect GraphQL API anomalies.

## Example Usage

```hcl
resource "wallarm_rule_graphql_detection" "graphql_detection" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  
  max_doc_size_kb = 100
  max_value_size_kb = 10
  max_depth = 10
  max_alias_size_kb = 5
  max_doc_per_batch = 10
  introspection = true
  debug_enabled = true

}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `max_doc_size_kb` - (optional) the limit for the size in kilobytes of an entire GraphQL query.
* `max_value_size_kb` - (optional) the limit for the size in kilobytes of an entire GraphQL query
* `max_depth` - (optional) the maximum allowed depth for a GraphQL query. By limiting query depth.
* `max_alias_size_kb` - (optional) the limit on the number of aliases that can be used in a single GraphQL query.
* `max_doc_per_batch` - (optional) the number of batched queries that can be sent in a single request.
* `introspection` - (optional) when enabled, the server will treat introspection requests—which can reveal the structure of your GraphQL schema—as potential attacks. Can be: `true`, `false`.
* `debug_enabled` - (optional) enabling this option means that requests containing the debug mode parameter will be considered potential attacks. Can be: `true`, `false`.
* `mode` - (**required**) protection behaviour which will be applied to the detected attack. Possible values: `monitoring`, `block`.

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "enum"`.

## Import

```
$ terraform import wallarm_rule_graphql_detection.graphql_detection 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.

For automated bulk import using the `wallarm_rules` data source, see the [Rules Import Guide](../guides/rules_import).

This resource is a **mitigation control**. For an overview of all mitigation controls and their parameter mapping, see the [Mitigation Controls Guide](../guides/mitigation_controls).

[1]: https://docs.wallarm.com/api-protection/graphql-rule/#mitigation-control-based-protection
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/