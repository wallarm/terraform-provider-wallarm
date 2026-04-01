---
layout: "wallarm"
page_title: "Creating Rules from Hits"
subcategory: "Guide"
description: |-
  How to create false positive suppression rules from Wallarm hit data using Terraform.
---

# Creating Rules from Hits

The hits-to-rules workflow creates false positive suppression rules directly from Wallarm hit data. Rules persist in Terraform state even after hits expire from the API.

## Overview

Wallarm hits are **ephemeral** -- they have a retention period and can be dropped from the API at any time. This module fetches hit data once, caches it in Terraform state, and creates rules that survive independently of the source hits.

Two rule types are supported:

- **`wallarm_rule_disable_stamp`** -- allows specific attack signatures (stamps) at a given request point
- **`wallarm_rule_disable_attack_type`** -- allows specific attack types at a given request point

## Quick Start

### Step 1: Initialize state

Copy the example from `examples/hits-to-rules/` and run the first apply with no request IDs:

```bash
terraform apply
```

This creates the `wallarm_hits_index` resource in state. The `request_ids` variable defaults to `{}`, so no hits are fetched.

### Step 2: Add request IDs and apply

Find the request IDs of the false positive hits in the Wallarm Console (Attacks section). Add them to `terraform.tfvars`:

```hcl
request_ids = {
  "4666dee205d69757fd45cf94d2f0d7eb" = "{}"
}
```

Then apply:

```bash
terraform apply
```

This fetches the hits, caches the rules in state, and creates the suppression rules.

### Step 3: Add more request IDs

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
  # Default: request mode, all rule types
  "abc123" = "{}"

  # Attack mode: expand to all related hits by attack_id
  "def456" = "{\"mode\":\"attack\"}"

  # Filter: only generate disable_stamp rules
  "ghi789" = "{\"rule_types\":[\"disable_stamp\"]}"
}
```

### Config options

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `mode` | `request`, `attack` | `request` | `request` fetches direct hits only. `attack` expands to all related hits sharing the same attack campaign. |
| `rule_types` | `["disable_stamp"]`, `["disable_attack_type"]` | all types | Filter which rule types to generate. |

## How It Works

The module uses three resources working together:

1. **`wallarm_hits_index`** -- tracks which request IDs have been fetched (persistent index in state)
2. **`data.wallarm_hits`** -- fetches hit data from the API, gated to only query new (uncached) request IDs
3. **`terraform_data.rules_cache`** -- persists rules per request ID using `ignore_changes` on input

On subsequent plans, the data source is skipped for cached request IDs, and rules are read from the `terraform_data` state. This means:

- No API calls for previously fetched hits
- Rules survive even after hits expire from the API
- Adding new request IDs only fetches the new ones

### Why two applies?

On the very first apply, `wallarm_hits_index` doesn't exist in state yet. Its `cached_request_ids` output is `(known after apply)`, which Terraform cannot use in `for_each`. Running the first apply with empty `request_ids` creates the resource in state, making `cached_request_ids` known on the next plan.

## Generating HCL Config Files

Optionally generate standalone `.tf` files for reference or migration:

```bash
terraform apply -var='generate_configs=true'
```

This uses `wallarm_rule_generator` to write HCL files to `./generated_rules/` (configurable via `output_dir`).

### Migrating to standalone resources with moved blocks

Generated files include `moved` blocks that map from the `for_each`-based resources to standalone named resources. This lets you migrate without destroying and recreating rules.

**Example generated output:**

```hcl
resource "wallarm_rule_disable_stamp" "fp_4666dee2_48c0e969_7994" {
  client_id            = 8649
  comment              = "Managed by Terraform"
  variativity_disabled = true
  stamp                = 7994
  # ...
}

moved {
  from = wallarm_rule_disable_stamp.this["4666dee2_48c0e969_7994"]
  to   = wallarm_rule_disable_stamp.fp_4666dee2_48c0e969_7994
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

## Removing Request IDs

Remove a request ID from `terraform.tfvars` and apply. Terraform will:

1. Destroy the corresponding `terraform_data.rules_cache` entry
2. Destroy the associated rule resources (`wallarm_rule_disable_stamp`, `wallarm_rule_disable_attack_type`)
3. Remove the ID from the `wallarm_hits_index`

## Variables Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `api_token` | `string` | -- | Wallarm API token (sensitive) |
| `api_host` | `string` | `https://us1.api.wallarm.com` | Wallarm API endpoint |
| `client_id` | `number` | `null` | Client ID (uses provider default if null) |
| `request_ids` | `map(string)` | `{}` | Map of request_id to config JSON |
| `default_mode` | `string` | `request` | Default fetch mode |
| `generate_configs` | `bool` | `false` | Generate HCL config files |
| `output_dir` | `string` | `./generated_rules` | Output directory for generated configs |

## Outputs

| Output | Description |
|--------|-------------|
| `rules_created` | Count of rules by type and total |
| `rule_ids` | Map of rule key to Wallarm rule ID |
