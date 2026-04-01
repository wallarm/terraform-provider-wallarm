---
layout: "wallarm"
page_title: "Mitigation Controls"
subcategory: "Guide"
description: |-
  Overview of Wallarm mitigation controls available as Terraform resources.
---

# Mitigation Controls

Mitigation controls are session-based security rules that extend Wallarm's attack protection. Unlike request-level rules, they operate on traffic patterns across multiple requests — counting thresholds, tracking sessions, and applying rate limits.

For detailed documentation, see [Mitigation Controls Overview](https://docs.wallarm.com/about-wallarm/mitigation-controls-overview/) in Wallarm docs.

## Terraform Resources

| Wallarm UI name | Terraform resource | Key parameters |
|-----------------|-------------------|----------------|
| Real-time blocking mode | [`wallarm_rule_mode`](../resources/rule_mode) | `mode` |
| GraphQL API protection | [`wallarm_rule_graphql_detection`](../resources/rule_graphql_detection) | `mode`, `max_depth`, `max_value_size_kb`, `max_doc_size_kb`, `max_alias_size_kb`, `max_doc_per_batch` |
| Enumeration attack protection | [`wallarm_rule_enum`](../resources/rule_enum) | `mode`, `threshold`, `reaction`, `advanced_conditions`, `arbitrary_conditions` |
| BOLA protection | [`wallarm_rule_bola`](../resources/rule_bola) | `mode`, `threshold`, `reaction`, `advanced_conditions`, `arbitrary_conditions` |
| Forced browsing protection | [`wallarm_rule_forced_browsing`](../resources/rule_forced_browsing) | `mode`, `threshold`, `reaction`, `advanced_conditions`, `arbitrary_conditions` |
| Brute force protection | [`wallarm_rule_brute`](../resources/rule_brute) | `mode`, `threshold`, `reaction`, `advanced_conditions`, `arbitrary_conditions` |
| DoS protection | [`wallarm_rule_rate_limit_enum`](../resources/rule_rate_limit_enum) | `mode`, `threshold`, `reaction`, `advanced_conditions`, `arbitrary_conditions` |
| File upload restriction policy | [`wallarm_rule_file_upload_size_limit`](../resources/rule_file_upload) | `mode`, `max_size`, `size_unit`, `file_types` |

## Parameter Mapping

### Scope

The UI "Scope" maps to the `action` block in Terraform — it defines which requests the control applies to. An empty `action` applies the control to all traffic.

```hcl
action {
  type  = "iequal"
  value = "example.com"
  point = {
    header = "HOST"
  }
}
```

### Mitigation Mode

The UI "Mitigation mode" maps to the `mode` attribute.

| UI mode | Terraform value | Available on |
|---------|----------------|--------------|
| Inherited | `"default"` | `rule_mode`, `graphql_detection`, `file_upload_size_limit` |
| Monitoring | `"monitoring"` | All |
| Blocking | `"block"` | All |
| Safe blocking | `"safe_blocking"` | `rule_mode` |
| Excluding | `"off"` | `rule_mode`, `graphql_detection`, `file_upload_size_limit` |

### Threshold

The UI "Threshold" section maps to the `threshold` block. Defines when the control triggers.

```hcl
threshold {
  count  = 30    # Number of requests
  period = 60    # Time window in seconds
}
```

Available on: `rule_brute`, `rule_bola`, `rule_enum`, `rule_forced_browsing`, `rule_rate_limit_enum`.

### Reaction

The UI "Reaction" section maps to the `reaction` block. Defines what happens when the threshold is exceeded. Values are durations in seconds (600–315569520).

```hcl
reaction {
  block_by_ip      = 3600   # Block source IP for 1 hour
  block_by_session = 3600   # Block session for 1 hour
}
```

| UI field | Terraform attribute | Notes |
|----------|-------------------|-------|
| Block by IP | `block_by_ip` | Duration in seconds |
| Block by session | `block_by_session` | Duration in seconds |
| Graylist by IP | `graylist_by_ip` | Only valid when `mode = "graylist"` |

Available on: `rule_brute`, `rule_bola`, `rule_enum`, `rule_forced_browsing`, `rule_rate_limit_enum`.

### Advanced Conditions

The UI "Advanced conditions" maps to the `advanced_conditions` block. Filters on session context parameters.

```hcl
advanced_conditions {
  field    = "ip"
  operator = "equal"
  value    = ["1.2.3.4"]
}
```

Available on: `rule_brute`, `rule_bola`, `rule_enum`, `rule_forced_browsing`, `rule_rate_limit_enum`.

### Arbitrary Conditions

Custom conditions on specific request points, mapped via `arbitrary_conditions`.

```hcl
arbitrary_conditions {
  point    = [["header", "X-Custom"]]
  operator = "equal"
  value    = ["blocked"]
}
```

Available on: `rule_brute`, `rule_bola`, `rule_enum`, `rule_forced_browsing`, `rule_rate_limit_enum`.

### Title and Comment

The `title` and `comment` are independent fields. In the Wallarm Console UI, "Title" corresponds to the `title` attribute and "Description" corresponds to `comment`. Both are optional. `comment` defaults to `"Managed by Terraform"` when not specified.

