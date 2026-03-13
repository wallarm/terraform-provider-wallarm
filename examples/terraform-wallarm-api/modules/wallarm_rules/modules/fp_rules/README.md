# fp_rules

Creates false-positive suppression rules (`disable_stamp` and `disable_attack_type`) from aggregated hit data.

## Overview

This module takes aggregated hit data (grouped by detection point) and creates Wallarm rules to suppress false-positive detections. It generates one YAML config file per `point_hash x rule_type` combination and expands multi-value fields (stamps, attack types) into individual resources.

## How It Works

```
Input: points map (point_hash => { point_wrapped, stamps, attack_types, attack_ids })
  │
  ├─ Cartesian product: point_hash × rule_type
  │   └─ One YAML config file per combination
  │
  ├─ disable_stamp: one resource per (point_hash × stamp)
  │   └─ wallarm_rule_disable_stamp
  │
  └─ disable_attack_type: one resource per (point_hash × attack_type)
      └─ wallarm_rule_disable_attack_type
```

### Config File Handling

Each config file stores editable fields (stamps, attack_types, point, action) alongside metadata (request_id, domain, path). On subsequent applies:

- **Editable fields** are read back from the YAML file (preserved across applies)
- **Metadata** is always written fresh (never read back)
- **Exception**: `attack_types` in `disable_stamp` configs are always taken from hit data, not the YAML

## Usage

Called by the `wallarm_rules` parent module:

```hcl
module "fp_rules" {
  for_each   = local.rules_by_request
  source     = "./modules/fp_rules"

  client_id  = var.client_id
  request_id = each.value.request_id
  rule_types = each.value.rule_types
  action     = each.value.action
  domain     = each.value.domain
  path       = each.value.path
  poolid     = each.value.poolid
  points     = each.value.points
  config_dir = var.fp_config_dir
}
```

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | — | yes | Wallarm client ID |
| `request_id` | `string` | — | yes | Original request ID from hits |
| `rule_types` | `list(string)` | — | yes | Rule types to create: `"disable_stamp"`, `"disable_attack_type"` |
| `action` | `any` | — | yes | Action conditions from the hit data source |
| `domain` | `string` | — | yes | Request domain from hit |
| `path` | `string` | — | yes | Request path from hit |
| `poolid` | `number` | — | yes | Application pool ID from hit |
| `points` | `any` | — | yes | Map of `point_hash => { point_wrapped, stamps, attack_types, attack_ids }` |
| `config_dir` | `string` | — | yes | Directory where YAML config files are written |

## Outputs

| Name | Description |
|------|-------------|
| `rule_ids` | Map of rule keys to their created rule IDs (all types) |
| `config_files` | Paths to generated config files |

## Config File Organization

Config files are organized in per-request_id subdirectories:

```
<config_dir>/
├── <request_id_1>/
│   ├── a1b2c3d4_disable_stamp.yaml
│   └── a1b2c3d4_disable_attack_type.yaml
├── <request_id_2>/
│   └── e5f6g7h8_disable_stamp.yaml
└── ...
```

File naming pattern: `<point_hash_prefix>_<rule_type>.yaml`

The point hash prefix is the first 8 characters of the SHA256 hash of the detection point.
