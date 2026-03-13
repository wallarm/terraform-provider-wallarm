# custom_rules

Creates Wallarm WAF rules from variable definitions with automatic path-to-action expansion. Supports 25 resource types.

## Overview

This module takes a list of rule definitions from variables, expands their `path` field into Wallarm action conditions, generates editable config files (YAML or HCL), and creates the appropriate `wallarm_rule_*` resources.

### Key Features

- **Path-to-action expansion** — `/api/v1/users.json` auto-expands into HOST, path segment, action_name, and action_ext conditions
- **Wildcard support** — `*` (any single segment) and `**` (any depth)
- **25 resource types** — From simple masking to complex brute-force protection
- **Variables-first pattern** — Variable values always override config file values
- **Config format choice** — YAML (default) or HCL via `config_format`

## Supported Resource Types

| Resource Type | Description |
|---------------|-------------|
| `wallarm_rule_binary_data` | Mark request part as binary data |
| `wallarm_rule_masking` | Mask sensitive data |
| `wallarm_rule_disable_attack_type` | Disable detection of specific attack types |
| `wallarm_rule_disable_stamp` | Disable specific detection stamps |
| `wallarm_rule_vpatch` | Virtual patch (block specific attack types) |
| `wallarm_rule_uploads` | Configure file upload handling |
| `wallarm_rule_ignore_regex` | Disable a regex-based detection rule |
| `wallarm_rule_parser_state` | Enable/disable specific parsers |
| `wallarm_rule_regex` | Custom regex-based detection |
| `wallarm_rule_file_upload_size_limit` | File upload size limits |
| `wallarm_rule_rate_limit` | Rate limiting |
| `wallarm_rule_credential_stuffing_point` | Credential stuffing detection points |
| `wallarm_rule_credential_stuffing_regex` | Credential stuffing regex detection |
| `wallarm_rule_mode` | Filtration mode (monitoring/block/off) |
| `wallarm_rule_set_response_header` | Add/modify response headers |
| `wallarm_rule_overlimit_res_settings` | Overlimit resource processing |
| `wallarm_rule_graphql_detection` | GraphQL-specific detection and limits |
| `wallarm_rule_bruteforce_counter` | Brute-force counter endpoint |
| `wallarm_rule_dirbust_counter` | Directory busting counter endpoint |
| `wallarm_rule_bola_counter` | BOLA counter endpoint |
| `wallarm_rule_brute` | Brute-force protection with threshold/reaction |
| `wallarm_rule_bola` | BOLA protection with threshold/reaction |
| `wallarm_rule_enum` | Account enumeration protection |
| `wallarm_rule_rate_limit_enum` | Rate-limit-based enumeration protection |
| `wallarm_rule_forced_browsing` | Forced browsing protection |

## Path-to-Action Expansion

The `path` field is parsed into Wallarm action conditions that specify where the rule applies.

### How It Works

Given `path = "/api/v1/users.json"`, `domain = "example.com"`, `instance = "101"`:

| Condition | type | value | point |
|-----------|------|-------|-------|
| Instance | `null` | `""` | `{ instance = "101" }` |
| Domain | `iequal` | `example.com` | `{ header = "HOST" }` |
| action_name | `equal` | `""` | `{ action_name = "users" }` |
| action_ext | `equal` | `""` | `{ action_ext = "json" }` |
| Path segment 0 | `equal` | `api` | `{ path = "0" }` |
| Path segment 1 | `equal` | `v1` | `{ path = "1" }` |
| Limiter | `absent` | `""` | `{ path = "2" }` |

### Wildcard Support

| Wildcard | Meaning | Example |
|----------|---------|---------|
| `*` | Any single value in a path segment | `/api/*/users` |
| `**` | Any depth (indefinite path) | `/api/**/users` |

**Rules for `**`:**
- Must be the last directory segment (before the action component)
- Cannot be the final path component (e.g. `/api/**` is invalid)
- Suppresses the path limiter (allows any depth)

**Rules for `*`:**
- In path segments: skips the `path[N]` condition (matches any value)
- As action_name: skips the `action_name` condition
- As extension (`*.json`): skips the `action_ext` condition
- As domain: skips the HOST header condition

### Special Cases

| Path | Behavior |
|------|----------|
| `/` or `""` | Root path: `action_name=""`, `action_ext=absent`, `path[0]=absent` |
| Deep path (>10 segments) | Falls back to single `uri` condition |
| No dot in last segment | `action_ext = absent` |
| Dot in last segment | Splits into `action_name` + `action_ext` |

## Usage Examples

See [EXAMPLES.tfvars](EXAMPLES.tfvars) for commented examples of all 25 resource types.

### Basic Examples

```hcl
custom_rules = [
  # Block mode for admin area
  {
    name          = "block_admin"
    resource_type = "wallarm_rule_mode"
    mode          = "block"
    path          = "/admin"
    domain        = "example.com"
  },

  # Mask sensitive data
  {
    name          = "mask_passwords"
    resource_type = "wallarm_rule_masking"
    point         = [["post"], ["json_doc"], ["hash", "password"]]
    path          = "/api/auth/login"
    domain        = "example.com"
  },

  # Rate limiting
  {
    name          = "rate_limit_api"
    resource_type = "wallarm_rule_rate_limit"
    rate          = 100
    burst         = 5
    rsp_status    = 503
    time_unit     = "rps"
    path          = "/api/v1"
    domain        = "example.com"
  },

  # Wildcard: any version
  {
    name          = "any_api_version"
    resource_type = "wallarm_rule_mode"
    mode          = "monitoring"
    path          = "/api/*/users"
    domain        = "example.com"
  },

  # Globstar: any depth
  {
    name          = "deep_api_match"
    resource_type = "wallarm_rule_mode"
    mode          = "block"
    path          = "/api/**/admin"
    domain        = "example.com"
  },
]
```

### Header and Query Conditions

```hcl
{
  name          = "json_api_rule"
  resource_type = "wallarm_rule_mode"
  mode          = "block"
  path          = "/api/endpoint"
  domain        = "example.com"
  method        = "POST"
  headers = [
    { name = "Content-Type", value = "application/json", type = "iequal" },
    { name = "X-Custom",     value = "test",             type = "equal" },
  ]
  query = [
    { key = "version", value = "2", type = "equal" },
  ]
}
```

### Cross-Referencing Rules

`wallarm_rule_ignore_regex` can reference a `wallarm_rule_regex` by name:

```hcl
custom_rules = [
  {
    name          = "my_regex"
    resource_type = "wallarm_rule_regex"
    attack_type   = "sqli"
    regex         = ".*pattern.*"
    path          = "/api/search"
    domain        = "example.com"
  },
  {
    name          = "ignore_my_regex"
    resource_type = "wallarm_rule_ignore_regex"
    regex_rule    = "my_regex"  # references the rule above by name
    path          = "/api/webhook"
    domain        = "example.com"
  },
]
```

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | — | yes | Wallarm client ID |
| `rules` | `list(object)` | `[]` | no | Rule definitions (see full type below) |
| `config_dir` | `string` | — | yes | Directory where config files are written |
| `config_format` | `string` | `"yaml"` | no | Config file format: `"yaml"` or `"hcl"` |

### Rule Object Type

```hcl
{
  # Required
  name          = string         # Unique rule name (used as config filename and resource key)
  resource_type = string         # One of the 25 supported wallarm_rule_* types

  # Scope (path auto-expanded into action conditions)
  comment  = optional(string, "Managed by Terraform")
  path     = optional(string, "")
  domain   = optional(string, "")
  instance = optional(string, "")
  method   = optional(string, "")
  scheme   = optional(string, "")
  proto    = optional(string, "")

  # Detection point
  point = optional(list(list(string)))

  # Header conditions
  headers = optional(list(object({ name = string, value = string, type = optional(string, "equal") })), [])

  # Query parameter conditions
  query = optional(list(object({ key = string, value = string, type = optional(string, "equal") })), [])

  # Multi-value expansion
  attack_types = optional(list(string), [])    # disable_attack_type, vpatch
  stamps       = optional(list(number), [])    # disable_stamp

  # Rule-specific fields (use as needed for each resource_type)
  attack_type    = optional(string, "")
  mode           = optional(string, "")
  regex          = optional(string, "")
  regex_id       = optional(number, 0)
  regex_rule     = optional(string, "")
  experimental   = optional(bool, false)
  parser         = optional(string, "")
  state          = optional(string, "")
  file_type      = optional(string, "")
  delay          = optional(number, 0)
  burst          = optional(number, 0)
  rate           = optional(number, 0)
  rsp_status     = optional(number, 0)
  time_unit      = optional(string, "")
  size           = optional(number, 0)
  size_unit      = optional(string, "")
  header_name    = optional(string, "")
  header_mode    = optional(string, "")
  header_values  = optional(list(string), [])
  overlimit_time = optional(number, 0)
  introspection  = optional(bool, false)
  debug_enabled  = optional(bool, false)
  max_depth         = optional(number, 0)
  max_value_size_kb = optional(number, 0)
  max_doc_size_kb   = optional(number, 0)
  max_alias_size_kb = optional(number, 0)
  max_doc_per_batch = optional(number, 0)
  login_point     = optional(list(list(string)), [])
  login_regex     = optional(string, "")
  case_sensitive  = optional(bool, false)
  cred_stuff_type = optional(string, "default")
  threshold       = optional(object({ period = number, count = number }))
  reaction        = optional(object({ block_by_session = number, block_by_ip = number, graylist_by_ip = number }))
  enumerated_parameters = optional(object({ mode = string, points = list, name_regexps = list, ... }))
}
```

## Outputs

| Name | Description |
|------|-------------|
| `rule_ids` | Map of custom rule keys to their created rule IDs (all resource types) |
| `config_files` | Paths to generated config files |

## Config File Pattern

Config files are generated per rule:
- **YAML**: `<config_dir>/<resource_type>_<name>.yaml`
- **HCL**: `<config_dir>/<resource_type>_<name>.tf`

The variables-first merge pattern:
1. YAML/HCL file provides base values
2. Variable values override file values
3. `action` is always computed from path expansion (never read from file)
