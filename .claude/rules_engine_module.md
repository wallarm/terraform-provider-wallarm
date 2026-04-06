# Rules Engine Module

**Note**: This module lives in a separate repository (`terraform-wallarm-api`), not in this provider repo. Documented here for context since the provider's resources and caches are designed to support it.

Located at `terraform-wallarm-api/modules/wallarm_rules/modules/rules_engine/`.

## Directory-Based Organization

Rules are organized in action directories. Each directory has a `.action.yaml` defining the action scope, plus rule YAML files:

```
configs/
├── _default/
│   ├── .action.yaml          ← action config (conditions: [])
│   └── scanner_mode.yaml
├── example.com_api_a3f2e1b7/
│   ├── .action.yaml          ← action config (conditions from API)
│   ├── hits_e5f6g7h8_disable_stamp.yaml
│   ├── imported_a1b2c3d4_vpatch.yaml
│   └── set_mode.yaml
```

## Config Discovery (`config.tf`)

- `_raw_yaml` — cached `yamldecode`, each file read once
- Separates `.action.yaml` (→ `action_configs`) from rule YAMLs (→ `all_yaml_files`)
- `action_map` — `conditions_hash → action_id` for fast lookup
- Duplicate name detection across directories

## Action Map (`actions.tf`)

Pure local — no resources, zero API calls on steady-state plans. Maps `conditions_hash → { action_id, dir_name, conditions }` from two sources merged at plan time:
- **Source A**: `.action.yaml` files on disk (from `fileset`)
- **Source B**: `var.generated_rules` (from hits, plan-time known — enables single-apply)
- Default action (`_default`, empty conditions) always included

This replaces the earlier `wallarm_action` resource approach which caused per-action API calls on every plan.

## Hint Cache Optimization

- `HintCreate` → `Insert` (adds to cache, no re-fetch)
- `HintUpdateV3` → `Insert` (updates cache, no re-fetch)
- Lazy pagination: only fetches pages needed to find managed rule IDs
- `LoadAll` only called by `data.wallarm_rules` (import/refresh)

## API Call Budget

| Operation | API calls | Notes |
|-----------|-----------|-------|
| Plan (0 managed rules) | 0 | No Reads, no cache activity |
| Plan (5 managed rules, page 1) | 1 page fetch | Lazy pagination finds all on first page |
| Plan (100 managed rules) | 1-N page fetches | Stops when all IDs found |
| Create N rules (first apply) | N creates + 1 page fetch + 1 hits fetch | Single apply |
| Destroy N rules | N deletes | HintDelete only, cache invalidated |
| Import (refresh) | ~15 page fetches (LoadAll) | Full fetch for data.wallarm_rules |

## Rule Sources and Naming

Three sources, one `configs/` directory tree (single `var.configs_dir`):
- **Hits**: `hits_{point_hash_prefix}_{rule_type}.yaml` — auto-generated from `data.wallarm_hits`, `origin: hit` in metadata
- **Imports**: `imported_{terraform_resource}_{rule_id}.yaml` — from hints_cache, `origin: import` in metadata
- **Manual**: user-written, no prefix, no origin field

All three sources converge to the same thing: **YAML file on disk is the source of truth** after initial creation. No runtime dependency on data sources, terraform_data, or generated_rules after the file exists.

## Hints Cache (`modules/hints_cache/`)

Persistent index of all rules from the Wallarm API. Stores references only — no rule-specific data fields. Full rule data is fetched ephemerally during refresh for import.

**Index entry** (persistent in `terraform_data.hints_index`):
```
rule_id, action_id, import_id, terraform_resource, conditions_hash, action_dir_name
```

**Suffix lookup** (persistent in `terraform_data.hints_suffix`):
Maps rule_id → expansion value (stamp, attack_type, file_type, parser) for building correct `for_each` keys.

**`is_managed`**: computed in parent module outputs by comparing index entries against `keys(module.rules.rule_ids)`. Not stored in the index (would create circular dependency).

**Behavior:**

| Scenario | API calls | Result |
|----------|-----------|--------|
| `terraform plan` (steady state) | 0 | Reads index from state |
| `terraform apply -var='import_rules=true'` (step 1) | 1 rules fetch | Index refreshed, import blocks + YAMLs written |
| `terraform apply -var='import_rules=true'` (step 2) | 0 (hint cache) | Rules imported + updated (variativity + comment) |
| `rm wallarm_rule_imports.tf && terraform apply` | 0 | Cleanup null_resources from state, files persist |

## Import Flow

```bash
# Step 1: Fetch rules from API, write import blocks + YAML configs
terraform apply -var='import_rules=true'

# Step 2: Import rules + update defaults (variativity_disabled=true, comment)
terraform apply -var='import_rules=true'

# Step 3: Cleanup state references (files persist on disk)
rm wallarm_rule_imports.tf
terraform apply

# Steady state — YAML files drive everything
terraform plan  # clean
```

**What happens during import:**
1. `hints_cache` refreshes → fetches all rules via `data.wallarm_rules` (uses `rules_export`)
2. Builds index (references only) + full rule data (ephemeral)
3. `import_rules` output provides `generated_rules` to rules_engine with `variativity_disabled=true` and `comment` defaulted
4. Import blocks written to `wallarm_rule_imports.tf`
5. YAML configs written via `null_resource` (persist after state cleanup)
6. `.action.yaml` written per action directory
7. On step 2: Terraform matches import blocks → resource blocks → import + update

**YAML persistence:** Uses `null_resource` + `local-exec` instead of `local_file`. Files persist on disk even after `null_resource` is destroyed from state (when `generated_rules` becomes empty on steady state). `local_file` would delete the file on destroy.

## Hits Fetch Gating

`data.wallarm_hits` runs only when needed:
- Auto-gate: `fileset(configs_dir, "**/hits_*.yaml")` — if no hits YAML files exist, fetch from API
- After first apply: YAML files exist → data source skipped → zero API calls
- Override: `var.fetch_hits = true` to force re-fetch (e.g., when adding new request_ids)
- `include_instance` (bool, default true) — controls whether instance (pool ID) is included in action conditions

## Delete Cascade

Removing rule YAML files → next apply:
1. Rule resources destroyed (`HintDelete` API calls)
2. API auto-cleans empty actions
3. `.action.yaml` remains (informational) — can be deleted manually

## TODOs

- **Auto-refresh trigger**: detect when YAML file count changes → auto-refresh index. Currently manual (`import_rules=true`).
- **`is_managed` in index**: currently computed in parent outputs (avoids circular dependency). Explore storing in index for direct use.
