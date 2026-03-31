---
layout: "wallarm"
page_title: "Wallarm: wallarm_rules_settings"
subcategory: "Common"
description: |-
  Provides the resource to manage Wallarm rules engine settings.
---

# wallarm_rules_settings

Provides the resource for managing [rules engine settings][2] — custom ruleset compilation, snapshot configuration, and API risk score weights.

Each client should have only one `wallarm_rules_settings` resource. Multiple resources for the same client will conflict.

## Example Usage

```hcl
resource "wallarm_rules_settings" "settings" {
  min_lom_format            = 50
  max_lom_format            = 54
  max_lom_size              = 10240
  lom_disabled              = false
  lom_compilation_delay     = 0
  rules_snapshot_enabled    = true
  rules_snapshot_max_count  = 5
  rules_manipulation_locked = false
  heavy_lom                 = false
  parameters_count_weight   = 6
  path_variativity_weight   = 6
  pii_weight                = 8
  request_content_weight    = 6
  open_vulns_weight         = 9
  serialized_data_weight    = 6
  risk_score_algo           = "maximum"
}
```

```hcl
# Multi-tenant: configure per client
resource "wallarm_rules_settings" "tenant_settings" {
  client_id      = 8649
  min_lom_format = 50
  max_lom_format = 70
}
```

## Argument Reference

### General

* `client_id` - (Optional, ForceNew) Client ID. Defaults to the provider's client ID.

### Custom Ruleset Compilation

* `min_lom_format` - (Optional) Minimum custom ruleset format version. Set to `0` to use server default.
* `max_lom_format` - (Optional) Maximum custom ruleset format version. Set to `0` to use server default.
* `max_lom_size` - (Optional) Maximum custom ruleset size in bytes. Minimum: `1025`.
* `lom_disabled` - (Optional) Forbid custom ruleset compilation (prevents rule updates on nodes).
* `lom_compilation_delay` - (Optional) Delay before custom ruleset compilation (seconds).
* `heavy_lom` - (Optional) Compile custom ruleset in a special queue for large rulesets.

### Snapshots

* `rules_snapshot_enabled` - (Optional) Enable rule snapshots during custom ruleset compilation.
* `rules_snapshot_max_count` - (Optional) Maximum number of rule snapshots stored. Range: 0–99.

### Rules Manipulation

* `rules_manipulation_locked` - (Optional) Lock rules to prevent changes.

### Risk Score Weights

Weights for API [risk score][1] calculation. Range: 0–10 for all weight fields.

* `parameters_count_weight` - (Optional) Weight of query and body parameters count.
* `path_variativity_weight` - (Optional) Weight of path variability (BOLA/IDOR risk).
* `pii_weight` - (Optional) Weight of parameters with sensitive data.
* `request_content_weight` - (Optional) Weight of file upload capability.
* `open_vulns_weight` - (Optional) Weight of active vulnerabilities.
* `serialized_data_weight` - (Optional) Weight of XML/JSON object acceptance.
* `risk_score_algo` - (Optional) Risk score calculation method. Possible values: `maximum`, `average`.

## Attributes Reference

All arguments are also exported as attributes.

## Import

```bash
terraform import wallarm_rules_settings.settings 8649/rules_settings
```

The import ID format is `{clientID}/rules_settings`.

[1]: https://docs.wallarm.com/api-discovery/overview/#endpoint-risk-score
[2]: https://docs.wallarm.com/user-guides/rules/rules/
