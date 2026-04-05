---
layout: "wallarm"
page_title: "Creating Rules from Hits"
description: |-
  How to create false positive suppression rules from Wallarm hit data using Terraform.
---

# Creating Rules from Hits

The hits-to-rules workflow creates false positive suppression rules directly from Wallarm hit data. Rules persist in Terraform state even after hits expire from the API.

## Overview

Wallarm hits are **ephemeral** -- they have a retention period and can be dropped from the API at any time. The `data.wallarm_hits` data source should only be called once per request ID to perform the initial fetch. After that, the rules data must be cached in Terraform state so that subsequent plans do not re-fetch from the API. If hits have expired, re-fetching would return empty results and Terraform would destroy the rules.

This module handles this automatically: `data.wallarm_hits` is gated to only fetch **new** (uncached) request IDs, and `terraform_data.cache` persists the data in state permanently.

Two rule types are supported:

- **`wallarm_rule_disable_stamp`** -- allows specific attack signatures (stamps) at a given request point
- **`wallarm_rule_disable_attack_type`** -- allows specific attack types at a given request point

~> **Note:** `xxe` and `invalid_xml` attack types do not produce stamps. Hits of these types can only be suppressed via `disable_attack_type` rules. When filtering with `rule_types = ["disable_stamp"]`, these attack types will not generate any rules -- use `disable_attack_type` or both rule types (default).

## Quick Start

### Step 1: Add request IDs and apply

Copy the example from `examples/hits-to-rules/`. Find the request IDs of the false positive hits in the Wallarm Console (Attacks section). Add them to `terraform.tfvars`:

```hcl
request_ids = {
  "4666dee205d69757fd45cf94d2f0d7eb" = "{}"
}
```

Then apply:

```bash
terraform apply
```

This creates the `wallarm_hits_index`, fetches the hits, caches the data, and creates the suppression rules.

### Step 2: Add more request IDs

Add new entries to `terraform.tfvars` and apply again. Only new (uncached) request IDs trigger API calls. Existing rules are preserved from state.

```hcl
request_ids = {
  "4666dee205d69757fd45cf94d2f0d7eb" = "{}"
  "b7ae0af153248f1b1d4f3bea3cd5ea90" = "{}"
}
```

## Per-Request Configuration

Each request ID maps to a JSON config string. Use `"{}"` for defaults.

```hcl
request_ids = {
  # Default: request mode, all rule types, all attack types
  "abc123" = "{}"

  # Attack mode: expand to all related hits by attack_id
  "def456" = "{\"mode\":\"attack\"}"

  # Filter: only generate disable_stamp rules
  "ghi789" = "{\"rule_types\":[\"disable_stamp\"]}"

  # Filter: only create rules for sqli hits
  "jkl012" = "{\"attack_types\":[\"sqli\"]}"

  # Combined: attack mode, only xss and rce hits
  "mno345" = "{\"mode\":\"attack\", \"attack_types\":[\"xss\",\"rce\"]}"
}
```

### Config options

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `mode` | `request`, `attack` | `request` | `request` fetches direct hits only. `attack` expands to all related hits sharing the same attack campaign. |
| `rule_types` | `["disable_stamp"]`, `["disable_attack_type"]` | all types | Filter which rule types to generate. |
| `attack_types` | `["sqli"]`, `["xss","rce"]`, etc. | all standard types | Filter which attack types produce rules. In attack mode, also controls which types to fetch from the API. |

## How It Works

The module uses three components:

1. **`wallarm_hits_index`** -- tracks which request IDs have been fetched. Exposes `ready` (false on first create, true after) and `cached_request_ids` (set of known IDs) for gating.
2. **`data.wallarm_hits`** -- fetches hit data from the API. Gated by `wallarm_hits_index` to only query new request IDs.
3. **`terraform_data.cache`** -- stores the `aggregated` output from `data.wallarm_hits` per request_id with `ignore_changes` on input. Data persists even after hits expire.

HCL locals then build a deduplicated map keyed by `action_hash` -- multiple request IDs sharing the same action (same host and path) are merged, with stamps unioned and new point groups added. Actions are stored separately to avoid duplication. Rules are expanded from this deduplicated map.

On subsequent plans, the data source is skipped for cached request IDs, and groups are read from `terraform_data.cache` state. This means:

- No API calls for previously fetched hits
- Rules survive even after hits expire from the API
- Adding new request IDs only fetches the new ones

### Single apply

On the very first apply with request IDs, `wallarm_hits_index` is created with `ready=false` (known at plan time via `CustomizeDiff`). This causes all request IDs to be fetched, cached, and rules created in a single apply. On subsequent applies, `ready=true` and only new request IDs are fetched.

## Resource Naming

Rule resources use `for_each` keys derived from hash prefixes:

```
wallarm_rule_disable_stamp.this["{action_hash}_{point_hash}_{stamp}"]
wallarm_rule_disable_attack_type.this["{action_hash}_{point_hash}_{attack_type}"]
```

Where:
- **`action_hash`** (16 hex chars) -- identifies the action scope (Host + path + pool). Derived from a SHA256 of the sorted action conditions (Ruby-compatible `ConditionsHash`).
- **`point_hash`** (16 hex chars) -- identifies the detection point (e.g., query parameter, header). Derived from a SHA256 of the point structure.
- **`stamp`** or **`attack_type`** -- the specific signature or attack type being suppressed.

Example: `wallarm_rule_disable_stamp.this["ed1d2ad7a1b2c3d4_48c0e969f1e2d3c4_6961"]`

Stamp groups (keyed by `action_hash_point_hash`) and attack_type groups (keyed by `action_hash_point_hash_attack_type`) are separate because stamps are not attack-type-scoped in the Wallarm API.

## Generating HCL Config Files

Optionally generate standalone `.tf` files for reference or migration:

```bash
terraform apply -var='generate_configs=true'
```

This uses `wallarm_rule_generator` with `source = "rules"` to write HCL files from the cached rules data.

### Migrating to standalone resources with moved blocks

Generated files include `moved` blocks that map from the `for_each`-based resources to standalone named resources. This lets you migrate without destroying and recreating rules.

**Example generated output:**

```hcl
resource "wallarm_rule_disable_stamp" "fp_ed1d2ad7a1b2c3d4_48c0e969f1e2d3c4_6961" {
  client_id            = 8649
  comment              = "Managed by Terraform"
  variativity_disabled = true
  stamp                = 6961
  # ...
}

moved {
  from = wallarm_rule_disable_stamp.this["ed1d2ad7a1b2c3d4_48c0e969f1e2d3c4_6961"]
  to   = wallarm_rule_disable_stamp.fp_ed1d2ad7a1b2c3d4_48c0e969f1e2d3c4_6961
}
```

**Migration steps:**

1. Generate configs:
   ```bash
   terraform apply -var='generate_configs=true'
   ```

2. Copy generated files into your working directory:
   ```bash
   cp ./generated_rules/*.tf .
   ```

3. Remove the `for_each`-based resource blocks (`wallarm_rule_disable_stamp.this` and `wallarm_rule_disable_attack_type.this`) from `main.tf`.

4. Verify with plan -- should show `moved` operations, no destroy/create:
   ```bash
   terraform plan
   ```

5. Apply:
   ```bash
   terraform apply
   ```

6. Remove the `moved` blocks from the generated files after one successful apply -- they are only needed for the migration.

## Deduplication

Multiple request IDs may produce identical rules -- for example, when the same request was repeated multiple times generating hits with different request IDs but the same structure. Deduplication happens in HCL locals: groups from different `terraform_data.cache` entries sharing the same `action_hash` and detection point are merged. Stamps are unioned via `distinct(flatten(...))`. This means only one Terraform resource is created per unique rule regardless of how many request IDs produced it.

Different request IDs with different actions (different hosts or paths) produce separate groups and separate rules -- they are never merged.

This prevents drift loops where the Wallarm API would deduplicate identical rules, causing Terraform to detect changes on every plan.

## Removing Request IDs

Remove a request ID from `terraform.tfvars` and apply. Terraform will:

1. Remove the ID from the `wallarm_hits_index`
2. Destroy the corresponding `terraform_data.cache` entry
3. The deduplicated locals recompute -- if no other request ID references the same action, the rules are destroyed

## Variables Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `api_token` | `string` | -- | Wallarm API token (sensitive) |
| `api_host` | `string` | `https://us1.api.wallarm.com` | Wallarm API endpoint |
| `client_id` | `number` | `null` | Client ID (uses provider default if null) |
| `request_ids` | `map(string)` | `{}` | Map of request_id to config JSON |
| `default_mode` | `string` | `request` | Default fetch mode |
| `include_instance` | `bool` | `true` | Include instance (pool ID) in action conditions. Set to `false` if your account excludes instance from actions. |
| `generate_configs` | `bool` | `false` | Generate HCL config files |
| `output_dir` | `string` | `./generated_rules` | Output directory for generated configs |

## Outputs

| Output | Description |
|--------|-------------|
| `rules_created` | Count of rules by type and total |
| `rule_ids` | Map of rule key to Wallarm rule ID |
