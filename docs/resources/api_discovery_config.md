---
layout: "wallarm"
page_title: "Wallarm: wallarm_api_discovery_config"
subcategory: "API Discovery"
description: |-
  Manages the per-tenant API Discovery configuration.
---

# wallarm_api_discovery_config

Manages the [API Discovery][1] configuration for a Wallarm tenant — the settings page in the console under **Settings → API Discovery**: the master toggle, which protocols to analyse, request-content filtering, endpoint-stability thresholds, parameter-type and PII detection thresholds, and the per-application include/exclude list.

API Discovery configuration is a **singleton per `client_id`** — the record always exists on the server. Declaring two `wallarm_api_discovery_config` resources for the same client will conflict (each apply overwrites the previous).

## Example Usage

```hcl
resource "wallarm_api_discovery_config" "this" {
  enabled                  = true
  apply_extended_filter    = true
  type_detection_threshold = 0.5
  pii_detection_threshold  = 0.1
  disabled_apps            = []

  protocols {
    rest    = true
    graphql = true
    soap    = true
    grpc    = true
    mcp     = true
  }

  endpoint_stability {
    min_count = 2
    min_time  = 300
  }
}
```

Multi-tenant:

```hcl
resource "wallarm_api_discovery_config" "tenant_a" {
  client_id = 22510
  enabled   = true
}
```

## Argument Reference

### General

* `client_id` - (Optional, ForceNew) Client ID. Defaults to the provider's client ID.
* `enabled` - (Optional, Default: `true`) Master toggle for API Discovery.

### Protocols

A `protocols` block configures which protocols API Discovery analyses. The block is `Optional` — omitting it enables all five.

* `rest` - (Optional, Default: `true`)
* `graphql` - (Optional, Default: `true`)
* `soap` - (Optional, Default: `true`)
* `grpc` - (Optional, Default: `true`)
* `mcp` - (Optional, Default: `true`)

### Filtering

* `apply_extended_filter` - (Optional, Default: `true`) Filter endpoints by response content type.
* `disabled_apps` - (Optional) List of pool/application IDs to exclude from API Discovery. Default: empty list.

### Detection thresholds

Fractions in the `[0.0, 1.0]` range (the console displays them as percentages).

* `type_detection_threshold` - (Optional, Default: `0.5`) Fraction of requests used to determine parameter types.
* `pii_detection_threshold` - (Optional, Default: `0.1`) Fraction of requests used to detect sensitive data.

### Endpoint stability

An `endpoint_stability` block defines when a newly observed endpoint is promoted to the discovered inventory.

* `min_count` - (Optional, Default: `2`, Range: `1–100`) Minimum number of requests.
* `min_time` - (Optional, Default: `300`, Range: `1–900`) Minimum time window in seconds.

## Attributes Reference

In addition to the arguments above, the following read-only attributes are populated from the API on Read:

* `call_points_storage_limit` - Storage limit for call points.
* `group_soap` - Whether SOAP endpoints are grouped under a parent definition.
* `allowed_content_types_patterns` - Content-type patterns considered for discovery.
* `sensitive_samples` - Sample-masking configuration (block): `enabled`, `min_masked`, `max_masked`, `mask_symbols`.
* `server_variability` - Server-variability heuristics (block): `enabled`, `by_date_enabled`, `by_local_code_enabled`, `by_email_enabled`, `by_alphanumeric_id_enabled`; nested `by_custom_paths` sub-block with `enabled` and `paths`.
* `extensions_whitelist` - File-extension allowlist (block): `enabled`, `extensions`.

These mirror the API's full response shape but cannot be set via Terraform — they're surfaced for drift visibility.

## Import

```bash
terraform import wallarm_api_discovery_config.this 22510/apid_config
```

The import ID format is `{client_id}/apid_config`. The suffix is constant; the `client_id` prefix identifies the tenant.

Deleting the resource from Terraform state is a noop — the singleton record persists on the server. To restore defaults, set the schema's Optional+Default values explicitly in HCL and `terraform apply`.

[1]: https://docs.wallarm.com/api-discovery/overview/
