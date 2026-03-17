# hits_fetcher

Fetches attack hits from the Wallarm API for a single request ID, aggregates them by detection point, and persists the result in Terraform state.

## Overview

This module provides a **fetch-once, persist-forever** data pipeline with two fetch modes:

- **Request mode** (default) — Fetches hits directly by `request_id`
- **Attack mode** — Fetches hits by `request_id`, then expands to all related hits via `attack_id`. Related hits are filtered by allowed attack types and must share the same action (Host + path). This enables full false-positive coverage across an entire attack campaign, not just a single request.

The fetch lifecycle:

- **First apply** (`fetch_hits = true`) — Calls `data.wallarm_hits`, aggregates raw hits by detection point (`point_hash`), and stores the result in `terraform_data.hits_state`
- **Subsequent applies** (`fetch_hits = false`) — Reads persisted data directly from Terraform state. No API calls.

Both paths produce **plan-time known** values, so downstream modules can safely use outputs in `for_each` without "known after apply" issues.

## How It Works

```
fetch_hits = true?
  ├─ YES → data.wallarm_hits API call (mode = "request" or "attack")
  │        → local.aggregated (group by point_hash)
  │        → terraform_data.hits_state stores aggregated data (ignore_changes)
  │        → outputs from local.aggregated
  │
  └─ NO  → outputs from terraform_data.hits_state.output (persisted state)
```

### Fetch Modes

**Request mode** (`mode = "request"`):
1. Fetch hits by `request_id` → return directly

**Attack mode** (`mode = "attack"`):
1. Fetch hits by `request_id`
2. Collect `attack_id` from each hit
3. Fetch all related hits belonging to those attack IDs (batches of 500)
4. Filter related hits by allowed attack types (xss, sqli, rce, xxe, ptrav, etc.)
5. Discard related hits with different action (domain + path) — strict guard against excessive rule creation
6. Merge with direct hits, deduplicate

Both modes produce the same output structure.

### Aggregation

Raw hits sharing the same `point_wrapped` (identified by `sha256(jsonencode(point_wrapped))`) are grouped together:

```
points = {
  "<point_hash>" = {
    point_wrapped = [...]         # Detection point (nested list)
    stamps        = [1, 2, 3]    # Sorted, deduplicated detection stamps
    attack_types  = ["sqli"]     # Distinct attack types across hits
    attack_ids    = [...]        # Distinct attack IDs
    hit_ids       = [...]        # Distinct hit IDs
  }
}
```

### State Persistence

```hcl
resource "terraform_data" "hits_state" {
  input = local.aggregated
  lifecycle { ignore_changes = [input] }
}
```

The `ignore_changes = [input]` lifecycle rule makes this write-once: the first apply stores the aggregated data, and all subsequent applies preserve it unchanged — even if `local.aggregated` evaluates to empty defaults (because the data source has `count = 0`).

## Usage

Called by the `wallarm_rules` parent module:

```hcl
module "hits" {
  for_each = var.requests
  source   = "./modules/hits_fetcher"

  client_id  = var.client_id
  request_id = each.key
  mode       = var.hits_mode  # "request" or "attack"

  # Auto-detect: fetch from API only when no fp_rules configs exist yet
  fetch_hits = length(try(fileset("${var.fp_config_dir}/${each.key}", "*.yaml"), toset([]))) == 0
}
```

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | — | yes | Wallarm client ID |
| `request_id` | `string` | — | yes | Request ID to fetch hits for |
| `time` | `list(number)` | `[]` | no | Optional `[from, to]` unix timestamps. Empty = provider defaults (6 months) |
| `fetch_hits` | `bool` | `false` | no | When true, fetch from API. When false, read from Terraform state |
| `mode` | `string` | `"request"` | no | Fetch mode: `"request"` (direct hits only) or `"attack"` (expand to related hits by attack_id) |
| `attack_types` | `list(string)` | `[]` | no | Override allowed attack types for filtering in attack mode. Empty = provider defaults |

## Outputs

| Name | Description |
|------|-------------|
| `action` | Rule action conditions derived from the hit (host, path, instance) |
| `action_hash` | SHA256 hash of sorted action conditions |
| `domain` | Request domain from the first hit |
| `path` | Request path from the first hit |
| `poolid` | Application pool ID from the first hit |
| `points` | Map of `point_hash => { point_wrapped, stamps, attack_types, attack_ids, hit_ids }` |
| `has_hits` | Whether any hits are available |
