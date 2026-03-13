# wallarm_rules

Parent module for all Wallarm WAF rule creation. Orchestrates hit fetching, aggregation, false-positive rule creation, and custom rule creation.

## Overview

This module combines two rule creation pipelines:

1. **Hit-based rules (false-positive suppression)** — Given a set of request IDs, the module fetches attack hits from the Wallarm API (once per request_id), persists aggregated data in Terraform state, and creates `disable_stamp` / `disable_attack_type` rules.

2. **Custom rules (variable-based)** — Rules defined in `terraform.tfvars` with a `path` field that auto-expands into Wallarm action conditions. Supports 25 resource types.

Both pipelines generate editable config files (YAML or HCL) and follow the variables-first pattern where variable values always override config file values.

## Architecture

```
wallarm_rules/
│
├─ module "hits" (for_each = var.requests)
│  └─ hits_fetcher/       Fetch + aggregate hits, persist in terraform_data
│
├─ locals
│  └─ rules_by_request    Map from hits_fetcher outputs (action, domain, path, poolid, points)
│
├─ module "fp_rules" (for_each = rules_by_request)
│  └─ fp_rules/           Create disable_stamp + disable_attack_type rules
│                          Config files: fp-rules-configs/<request_id>/*.yaml
│
└─ module "custom_rules"
   └─ custom_rules/       25 resource types with path-to-action expansion
                           Config files: rules_config/*.<yaml|tf>
```

## Usage

```hcl
module "wallarm_rules" {
  source = "./modules/wallarm_rules"

  client_id     = 12345
  hits_mode     = "attack"  # "request" (default) or "attack"
  requests      = {
    "abc123" = ["disable_stamp", "disable_attack_type"]
  }
  custom_rules  = [
    {
      name          = "block_admin"
      resource_type = "wallarm_rule_mode"
      mode          = "block"
      path          = "/admin"
      domain        = "example.com"
    }
  ]
  config_dir    = "${path.root}/rules_config"
  fp_config_dir = "${path.root}/fp-rules-configs"
  config_format = "yaml"
}
```

### Hits Mode

- **`"request"`** (default) — Fetches hits only for the exact `request_id`. Produces rules for that specific request.
- **`"attack"`** — Fetches hits for the `request_id`, then expands to all related hits sharing the same `attack_id`. Related hits are filtered by allowed attack types and must match the same action (Host + path). This produces broader FP suppression covering the entire attack campaign.

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | — | yes | Wallarm client ID |
| `requests` | `map(list(string))` | `{}` | no | Map of `request_id => [rule_types]` |
| `hits_mode` | `string` | `"request"` | no | Fetch mode: `"request"` (direct hits) or `"attack"` (expand to related hits by attack_id) |
| `custom_rules` | `any` | `[]` | no | Custom rule definitions (passed to custom_rules child module) |
| `config_dir` | `string` | — | yes | Directory where custom rule config files are written |
| `fp_config_dir` | `string` | — | yes | Directory where false-positive rule config files are written |
| `config_format` | `string` | `"yaml"` | no | Config file format: `"yaml"` or `"hcl"` |

## Outputs

| Name | Description |
|------|-------------|
| `rule_ids_by_request` | Rule IDs grouped by request_id (from fp_rules) |
| `custom_rule_ids` | Map of custom rule names to their created rule IDs |
| `all_rule_ids` | Flat map of all created rule IDs across fp_rules and custom_rules |

## Submodules

| Name | Description |
|------|-------------|
| [hits_fetcher](modules/hits_fetcher/) | Fetch hits from Wallarm API, aggregate by detection point, persist in Terraform state |
| [fp_rules](modules/fp_rules/) | Create false-positive suppression rules from aggregated hit data |
| [custom_rules](modules/custom_rules/) | Create rules from variable definitions with path expansion |

## How It Works

### Hit-Based Pipeline

1. **Gate** — `fileset()` checks if fp_rules config files exist for this request_id. If empty, `fetch_hits = true`
2. **Fetch + Aggregate** — `hits_fetcher` calls `data.wallarm_hits` per request ID, aggregates hits by detection point (`point_hash`), and stores the result in `terraform_data.hits_state`
3. **Persist** — `terraform_data.hits_state` with `ignore_changes = [input]` ensures data is written once and never overwritten
4. **Create** — `fp_rules` generates YAML config files in `fp-rules-configs/<request_id>/` and creates `wallarm_rule_disable_stamp` / `wallarm_rule_disable_attack_type` resources

On subsequent applies, the fileset gate detects existing config files and skips the API call entirely. All data comes from Terraform state.

### Custom Rules Pipeline

1. **Expand** — `path` field is parsed into action conditions (HOST header, path segments, action_name/ext, limiter)
2. **Config** — YAML or HCL config file generated per rule
3. **Merge** — Config file values merged with variable values (variables win)
4. **Create** — Appropriate `wallarm_rule_*` resource created

### Managing Request IDs

| Scenario | What happens |
|----------|------|
| Add new request_id | Added to `requests` map — auto-fetched (no configs yet), rules created |
| Subsequent applies | Config files exist — no API call, data from state |
| Remove request_id | Remove from `requests` map — terraform_data, rules, and config files destroyed |
