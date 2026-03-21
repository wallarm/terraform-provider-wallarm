# wallarm_rules

Parent module for unified Wallarm WAF rule management. All rules — custom, hit-based FP suppression, and imported — share the same YAML config format and are processed by a single rules engine.

## Overview

Three rule sources, one engine:

1. **Custom rules** — User writes YAML configs in `custom_rules/` directory. Path/domain/method/headers/query auto-expand into Wallarm action conditions.

2. **Hit-based FP rules** — Given request IDs, the module fetches attack hits from the Wallarm API, generates YAML configs in `fp_rules/<request_id>/`, and creates `disable_stamp` / `disable_attack_type` rules. User can edit generated YAML to add wildcards or modify stamps.

3. **Imported rules** — Fetches existing rules from the Wallarm API, reverse-maps action conditions back to path/domain/etc., groups expandable types (stamps, attack_types, file_types, parsers), and generates YAML configs in `import_rules/`.

All three sources feed into the **rules_engine** module which reads YAML, expands paths, creates resources, and generates reference HCL.

## Architecture

```
wallarm_rules/
│
├─ module "hits" (for_each = var.requests)
│  └─ hits_fetcher/         Fetch + aggregate hits, persist in terraform_data
│
├─ module "hits_generator" (for_each = var.requests)
│  └─ hits_generator/       Convert hit data → universal rule objects
│
├─ module "import_generator"
│  └─ import_generator/     Fetch rules from API → universal rule objects
│                            Groups stamps/attack_types/file_types/parsers by action_id
│
└─ module "rules" (single instance)
   └─ rules_engine/         Unified engine for all rule sources
      ├─ Read YAML from custom_rules/, fp_rules/, import_rules/
      ├─ Accept generated_rules from hits_generator + import_generator
      ├─ Expand path → action conditions (wildcards: *, **)
      ├─ Expand multi-value rules (stamps, attack_types, file_types, parsers)
      ├─ Create all 25 resource types
      └─ Generate reference HCL in _reference/ subdirs
```

## Directory Structure

```
rules_config/
├── custom_rules/          ← user writes YAML configs here
│   ├── block_admin.yaml
│   ├── mask_passwords.yaml
│   └── _reference/        ← auto-generated standard HCL (read-only)
│       └── wallarm_rule_mode_block_admin.tf
├── fp_rules/              ← generated from hits
│   ├── <request_id>/
│   │   ├── <hash>_disable_stamp.yaml
│   │   └── _reference/
│   │       └── wallarm_rule_disable_stamp_<hash>_disable_stamp_1001.tf
│   └── ...
└── import_rules/          ← generated from API import
    ├── imported_disable_stamp_12345.yaml
    └── _reference/
        └── wallarm_rule_disable_stamp_imported_disable_stamp_12345_1001.tf
```

## Universal YAML Config Format

All rule configs share the same shape:

```yaml
name: block_admin
resource_type: wallarm_rule_mode
comment: "Block admin panel"

# Scope (auto-expanded into action conditions)
path: "/api/v1/admin/*"      # Wildcards: * (any segment), ** (any depth)
domain: "example.com"
instance: "1"
method: "POST"
scheme: "https"
query:
  - key: "token"
    value: "secret"
headers:
  - name: "X-Custom"
    value: "val"

# Rule-specific
mode: "block"

# Detection point (for rules that support it)
point:
  - ["post"]
  - ["json_doc"]
```

### Multi-value grouping (one config = multiple resources)

```yaml
# 3 disable_stamp resources from one config
name: disable_sqli_login
resource_type: wallarm_rule_disable_stamp
stamps: [1001, 1002, 1003]
point: [["post"], ["form_urlencoded"]]
path: "/auth/login"
domain: "example.com"

# 2 disable_attack_type resources
name: disable_xss_sqli
resource_type: wallarm_rule_disable_attack_type
attack_types: [sqli, xss]

# 3 uploads resources
name: allow_uploads
resource_type: wallarm_rule_uploads
file_types: [docs, images, music]

# 3 parser_state resources (state always "disabled")
name: disable_parsers
resource_type: wallarm_rule_parser_state
parsers: [json_doc, xml, base64]
```

## Usage

```hcl
module "wallarm_rules" {
  source = "./modules/wallarm_rules"

  client_id = 12345

  # Directories
  config_dir        = "${path.root}/rules_config/custom_rules"
  fp_config_dir     = "${path.root}/rules_config/fp_rules"
  import_config_dir = "${path.root}/rules_config/import_rules"

  # Hit-based FP rules
  hits_mode = "attack"  # "request" (default) or "attack"
  requests  = {
    "abc123" = ["disable_stamp", "disable_attack_type"]
  }

  # Import existing rules from API
  is_importing      = false  # Set to true for first import
  import_rule_types = []     # Empty = all types
}
```

### Workflow: Custom Rules

1. Create a YAML file in `custom_rules/` directory
2. `terraform apply` — engine reads YAML, expands path, creates resources
3. Edit YAML to change the rule — next apply updates the resource
4. Delete YAML to destroy the rule

### Workflow: FP Rules from Hits

1. Add request_id to `var.requests`
2. `terraform apply` — fetches hits, generates YAML in `fp_rules/<request_id>/`, creates rules
3. Edit generated YAML (add wildcards, remove stamps) — next apply updates rules
4. Remove request_id from `var.requests` — rules and configs destroyed

### Workflow: Import Existing Rules

1. `terraform apply -var='is_importing=true'` — fetches all rules from API, generates YAML in `import_rules/`
2. Set `is_importing = false` — subsequent applies read from YAML, no API fetch
3. Edit YAML configs as needed
4. Delete YAML files for rules you don't want to manage

### Hits Mode

- **`"request"`** (default) — Fetches hits only for the exact `request_id`
- **`"attack"`** — Expands to all related hits sharing the same `attack_id`. Related hits are filtered by allowed attack types and must match the same action (Host + path)

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | — | yes | Wallarm client ID |
| `config_dir` | `string` | — | yes | Directory for custom rule YAML configs |
| `fp_config_dir` | `string` | — | yes | Directory for FP rule YAML configs |
| `import_config_dir` | `string` | — | yes | Directory for imported rule YAML configs |
| `requests` | `map(list(string))` | `{}` | no | Map of `request_id => [rule_types]` for FP rules |
| `hits_mode` | `string` | `"request"` | no | `"request"` or `"attack"` |
| `is_importing` | `bool` | `false` | no | Set true to fetch rules from API |
| `import_rule_types` | `list(string)` | `[]` | no | Filter import by API rule type(s) |

## Outputs

| Name | Description |
|------|-------------|
| `rule_ids` | Flat map of all created rule IDs (all sources) |
| `config_files` | Paths to generated YAML config files (from hits/import) |
| `reference_files` | Paths to generated reference HCL files |

## Submodules

| Name | Description |
|------|-------------|
| [rules_engine](modules/rules_engine/) | Core engine: YAML discovery, path expansion, multi-value expansion, resource creation, reference HCL |
| [hits_fetcher](modules/hits_fetcher/) | Fetch hits from API, aggregate by detection point, persist in state |
| [hits_generator](modules/hits_generator/) | Convert hit data → universal rule objects |
| [import_generator](modules/import_generator/) | Fetch rules from API, reverse-map actions, group expandable types |

### Deprecated (replaced by rules_engine)

| Name | Status |
|------|--------|
| [custom_rules](modules/custom_rules/) | Replaced by rules_engine. Kept for reference. |
| [fp_rules](modules/fp_rules/) | Replaced by rules_engine. Kept for reference. |

## Path-to-Action Expansion

The `path` field is parsed into Wallarm action conditions:

| Path | Conditions generated |
|------|---------------------|
| `/api/v1/users` | path[0]="api", path[1]="v1", action_name="users", action_ext=absent, path[2]=absent |
| `/api/v1/data.json` | path[0]="api", path[1]="v1", action_name="data", action_ext="json", path[2]=absent |
| `/` | action_name="", action_ext=absent, path[0]=absent |
| `/api/*/users` | path[0]="api", (path[1] skipped), action_name="users", path[2]=absent |
| `/api/**/admin` | path[0]="api", action_name="admin" (no limiter — any depth) |
| (empty) | No path conditions (global rule) |

Wildcard rules:
- `*` in a directory segment — matches any value at that position
- `*` as action_name — skips action_name condition
- `*.ext` — skips action_ext condition
- `**` as last directory — allows any depth (no path limiter)

## Reverse Mapping (Import)

The `import_generator` module uses Go code in the provider (`ReverseMapActions()`) to convert API action conditions back to path/domain/method/etc. This is the inverse of the path expansion above. The reverse mapping handles all condition types including wildcards (gaps in path indices → `*`, no limiter → `**`).

## Supported Resource Types (25)

`wallarm_rule_binary_data`, `wallarm_rule_masking`, `wallarm_rule_disable_attack_type`, `wallarm_rule_disable_stamp`, `wallarm_rule_vpatch`, `wallarm_rule_uploads`, `wallarm_rule_ignore_regex`, `wallarm_rule_parser_state`, `wallarm_rule_regex`, `wallarm_rule_file_upload_size_limit`, `wallarm_rule_rate_limit`, `wallarm_rule_credential_stuffing_point`, `wallarm_rule_credential_stuffing_regex`, `wallarm_rule_mode`, `wallarm_rule_set_response_header`, `wallarm_rule_overlimit_res_settings`, `wallarm_rule_graphql_detection`, `wallarm_rule_brute`, `wallarm_rule_bruteforce_counter`, `wallarm_rule_dirbust_counter`, `wallarm_rule_bola`, `wallarm_rule_bola_counter`, `wallarm_rule_enum`, `wallarm_rule_rate_limit_enum`, `wallarm_rule_forced_browsing`
