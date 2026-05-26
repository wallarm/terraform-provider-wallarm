# Action Conditions & Path Expansion — Internal Reference

This file documents how Wallarm rule `action` conditions work, how paths are
expanded into conditions, and how the Terraform module implements this.
Use as authoritative reference when modifying custom_rules or fp_rules modules.

---

## 1. Action Condition Structure

Every `action` block in a `wallarm_rule_*` resource has three fields:

```hcl
action {
  type  = "<match_type>"   # "equal", "iequal", "regex", "absent", or null
  value = "<match_value>"  # The value to match against, or ""
  point = { <key> = "<val>" }  # Exactly one key identifying what to match
}
```

### Point Keys (what is being matched)

| Point Key      | Meaning                          | Example point map              |
|----------------|----------------------------------|--------------------------------|
| `instance`     | Application pool ID              | `{ instance = "101" }`         |
| `header`       | HTTP request header (UPPERCASED) | `{ header = "HOST" }`          |
| `path`         | URL path segment by index (0-based) | `{ path = "0" }`           |
| `action_name`  | Filename/endpoint (last path segment without ext) | `{ action_name = "users" }` |
| `action_ext`   | File extension of last segment   | `{ action_ext = "php" }`       |
| `uri`          | Full URI (fallback for too-deep paths) | `{ uri = "/a/b/c/d/..." }` |
| `method`       | HTTP method                      | `{ method = "POST" }`          |
| `scheme`       | URL scheme                       | `{ scheme = "https" }`         |
| `proto`        | HTTP protocol version            | `{ proto = "1.1" }`            |
| `query`        | Query parameter by name          | `{ query = "search" }`         |

### Match Types

| Type      | Meaning                                    |
|-----------|--------------------------------------------|
| `equal`   | Exact match (case-sensitive)               |
| `iequal`  | Case-insensitive match (used for HOST)     |
| `regex`   | Regular expression match                   |
| `absent`  | The field must not exist                   |
| `""` (empty/null) | No type check (used for method, scheme, proto) |

### Special Conventions

- **`instance`**: type="equal" value="" — the instance ID goes in the point map value
- **`header`**: header names are ALWAYS uppercased in the point map (`upper(v)` in TF)
- **`method`/`scheme`/`proto`**: type="" — the value goes in the point map value
- **`action_name`/`action_ext`**: type="equal" value="" — the actual name/ext goes in the point map value
- **`path`**: type="equal" value=segment — the segment value goes in `value`, the index goes in point map

---

## 2. Path-to-Action Expansion Algorithm

The module takes a user-friendly `path` string and expands it into a list of
action conditions. This mirrors the Go functions in the provider:
`buildActionFromHit()` + `locationToConditions()` + `actionNameExtConditions()`

### Input Fields (from variable)

```hcl
{
  path     = "/api/v1/users.json"  # URL path
  domain   = "example.com"         # HOST header match
  instance = "101"                 # Pool ID
  method   = "POST"                # HTTP method
  scheme   = "https"               # URL scheme
  proto    = "1.1"                 # HTTP version
  headers  = [...]                 # Additional header conditions
  query    = [...]                 # Query parameter conditions
}
```

### Expansion Steps

Given `path = "/api/v1/users.json"`:

1. **Strip leading `/`** and **split by `/`**:
   `raw_parts = ["api", "v1", "users.json"]`

2. **Separate directory segments from last segment**:
   - `dir_segments = ["api", "v1"]` (all but last)
   - `last_segment = "users.json"` (the action component)

3. **Split last segment into action_name / action_ext**:
   - If contains `.`: `action_name = "users"`, `action_ext = "json"`
   - If no `.`: `action_name = "users"`, `action_ext = absent`

4. **Build conditions list** (in this order):
   ```
   instance  → { type: "equal",  value: "",             point: { instance: "101" } }
   domain    → { type: "iequal", value: "example.com",  point: { header: "HOST" } }
   headers   → { type: <type>,   value: <value>,        point: { header: "<NAME>" } }  // per header
   action_name → { type: "equal", value: "",            point: { action_name: "users" } }
   action_ext  → { type: "equal", value: "",            point: { action_ext: "json" } }
   path[0]   → { type: "equal",  value: "api",          point: { path: "0" } }
   path[1]   → { type: "equal",  value: "v1",           point: { path: "1" } }
   limiter   → { type: "absent", value: "",             point: { path: "2" } }
   method    → { type: "equal",  value: "",              point: { method: "POST" } }
   scheme    → { type: "equal",  value: "",              point: { scheme: "https" } }
   proto     → { type: "equal",  value: "",              point: { proto: "1.1" } }
   query     → { type: <type>,   value: <value>,         point: { query: "<key>" } }  // per entry
   ```

### The Limiter

The **limiter** is an `absent` condition on `path[N]` where N = number of
directory segments. It fixes the path length — ensures the rule only matches
paths with exactly this many segments, not deeper paths.

Example: `/api/v1/users` → limiter at `path[2]` means `/api/v1/users/extra`
does NOT match.

The limiter is **suppressed** when `**` (globstar) is present.

---

## 3. Special Path Cases

### Root Path (`/` or `""`)

```
action_name → { type: "equal",  value: "", point: { action_name: "" } }
action_ext  → { type: "absent", value: "", point: { action_ext: "" } }
path limiter→ { type: "absent", value: "", point: { path: "0" } }
```

### Too-Deep Path (> max_path_depth segments, default 10)

Falls back to a single URI condition:
```
uri → { type: "equal", value: "/the/full/path/...", point: { uri: "/the/full/path/..." } }
```
The `**` wildcard exempts a path from the too-deep check.

### No-Dot Last Segment (e.g. `/api/users`)

When the last segment has no `.`, the extension is `absent`:
```
action_name → { type: "equal",  value: "", point: { action_name: "users" } }
action_ext  → { type: "absent", value: "", point: { action_ext: "" } }
```

### Dot in Last Segment (e.g. `/api/users.json`)

Extension is extracted and matched:
```
action_name → { type: "equal", value: "", point: { action_name: "users" } }
action_ext  → { type: "equal", value: "", point: { action_ext: "json" } }
```

---

## 4. Wildcard Support

### Single-Star `*` — Match Any Value

Can appear in any position. The condition for that position is **skipped**
(not emitted), which means "match anything".

| Position         | Effect                                |
|------------------|---------------------------------------|
| Path segment     | `path[N]` condition skipped           |
| Last segment     | `action_name` condition skipped       |
| Extension        | `action_ext` condition skipped        |
| Domain           | HOST header condition skipped         |

Examples:
- `/api/*/users` → path[0]="api", path[1] skipped (any), action_name="users"
- `/api/*.json` → path[0]="api", action_name skipped, action_ext="json"
- `domain = "*"` → no HOST header condition emitted

### Double-Star `**` — Any Depth (Globstar)

Allows matching paths of any depth beyond the specified prefix.

**Rules:**
- `**` MUST be the last **directory** segment (second-to-last element in raw_parts)
- `**` CANNOT be the final path component (there must be an action component after it)
- `**` is stripped from indexed_segments and suppresses the limiter

**Valid:** `/api/**/users` — matches `/api/users`, `/api/v1/users`, `/api/v1/v2/users`, etc.
**Invalid:** `/api/**` — no action component after `**`
**Invalid:** `/api/**/v1/**/users` — `**` can only appear once as last dir segment

Example: `path = "/api/**/users"`
```
raw_parts = ["api", "**", "users"]
dir_segments = ["api", "**"]
last_segment = "users"
has_globstar = true (** is last dir segment)
indexed_segments = ["api"]  (** stripped)
has_limiter = false (suppressed by **)
```

Generated conditions:
```
action_name → { type: "equal",  value: "", point: { action_name: "users" } }
action_ext  → { type: "absent", value: "", point: { action_ext: "" } }
path[0]     → { type: "equal",  value: "api", point: { path: "0" } }
// NO limiter — any depth allowed after "api"
```

### Validation

The module validates `**` patterns at plan time via `terraform_data.path_validation`:
- Final segment must not be `**`
- `**` must not appear in indexed_segments (only allowed as last dir, which gets stripped)

---

## 5. Header Conditions

Headers are added via the `headers` variable field:

```hcl
headers = [
  { name = "Content-Type", value = "application/json", type = "iequal" },
  { name = "X-Custom",     value = "test",             type = "equal" },
]
```

Each entry becomes:
```
{ type: "<type>", value: "<value>", point: { header: "<NAME_UPPERCASED>" } }
```

**Important:** The `domain` field is syntactic sugar for a HOST header condition
with type `iequal`. When `domain` is set, it generates:
```
{ type: "iequal", value: "<domain>", point: { header: "HOST" } }
```

---

## 6. Query Parameter Conditions

Query parameters are matched via the `query` variable field:

```hcl
query = [
  { key = "page",   value = "1",    type = "equal" },
  { key = "search", value = ".*",   type = "regex" },
]
```

Each entry becomes:
```
{ type: "<type>", value: "<value>", point: { query: "<key>" } }
```

---

## 7. Expansion Order in Generated Action List

The conditions are concatenated in this exact order:

1. **Instance** (if set)
2. **Domain** → HOST header (if set, skipped when `"*"`)
3. **Custom headers** (each entry)
4. **Too-deep fallback** → single URI condition (if path too deep)
5. **Root path** → action_name="" + action_ext absent + path[0] absent (if `/`)
6. **action_name** (if not wildcard `*`)
7. **action_ext** → absent (no dot), equal (specific), or skipped (wildcard `*`)
8. **Path segments** → path[N] for each dir segment (skip `*` wildcards)
9. **Limiter** → path[N] absent (suppressed when `**`)
10. **Method** (if set)
11. **Scheme** (if set)
12. **Proto** (if set)
13. **Query parameters** (each entry)

---

## 8. Provider-Side Go Functions

Source: `terraform-provider-wallarm/wallarm/provider/data_source_hits.go`

### `buildActionFromHit(domain, urlPath string, poolID int)`

Top-level function. Combines:
1. Instance condition (if poolID > 0)
2. HOST header condition (if domain != "")
3. `locationToConditions(urlPath)` results

### `locationToConditions(location string)`

Parses URL path into action conditions:
1. If path depth > maxPathDepth → single `{ uri: location }` fallback
2. Split by `/`, strip leading empty
3. Root path → `action_name=""` + `path[0] absent`
4. Normal: `actionNameExtConditions(last)` + `path[i]=segment` for dirs + terminating `path[N] absent`

### `actionNameExtConditions(segment string)`

Splits last path segment:
- Has dot: `action_name=name` + `action_ext=ext`
- No dot: `action_name=segment` + `action_ext absent`

---

## 9. Concrete Expansion Examples

### `/api/v1/users` on `example.com`
```
{ type: "iequal", value: "example.com", point: { header: "HOST" } }
{ type: "equal",  value: "",            point: { action_name: "users" } }
{ type: "absent", value: "",            point: { action_ext: "" } }
{ type: "equal",  value: "api",         point: { path: "0" } }
{ type: "equal",  value: "v1",          point: { path: "1" } }
{ type: "absent", value: "",            point: { path: "2" } }
```

### `/api/*/users` on `example.com` (wildcard segment)
```
{ type: "iequal", value: "example.com", point: { header: "HOST" } }
{ type: "equal",  value: "",            point: { action_name: "users" } }
{ type: "absent", value: "",            point: { action_ext: "" } }
{ type: "equal",  value: "api",         point: { path: "0" } }
// path[1] SKIPPED — * matches any
{ type: "absent", value: "",            point: { path: "2" } }
```

### `/api/**/users` on `example.com` (globstar)
```
{ type: "iequal", value: "example.com", point: { header: "HOST" } }
{ type: "equal",  value: "",            point: { action_name: "users" } }
{ type: "absent", value: "",            point: { action_ext: "" } }
{ type: "equal",  value: "api",         point: { path: "0" } }
// NO limiter — ** allows any depth
```

### `/api/data.json` on `*` (any domain)
```
// NO HOST condition — domain="*" skipped
{ type: "equal",  value: "",            point: { action_name: "data" } }
{ type: "equal",  value: "",            point: { action_ext: "json" } }
{ type: "equal",  value: "api",         point: { path: "0" } }
{ type: "absent", value: "",            point: { path: "1" } }
```

### `/` (root path) with instance 101
```
{ type: "equal",  value: "",            point: { instance: "101" } }
{ type: "equal",  value: "",            point: { action_name: "" } }
{ type: "absent", value: "",            point: { action_ext: "" } }
{ type: "absent", value: "",            point: { path: "0" } }
```

---

## 10. Variables-First Config Pattern

Rule configs use a three-way merge where variables are authoritative:

```hcl
merge(
  yaml_base,            # 1st — YAML file provides defaults only
  variable_values,      # 2nd — Variable values OVERRIDE yaml
  { action = computed } # 3rd — Action is always computed from path expansion
)
```

This means:
- Editing a YAML config file provides defaults for fields not set in variables
- Changing a variable value always takes effect (no stale YAML override)
- The `action` field is never read from YAML — always recomputed

---

## 11. Resource Types and Their Special Fields

| Resource Type | Key Fields |
|---------------|-----------|
| `wallarm_rule_binary_data` | point |
| `wallarm_rule_masking` | point |
| `wallarm_rule_disable_attack_type` | attack_types (expanded: one resource per type) |
| `wallarm_rule_disable_stamp` | stamps (expanded: one resource per stamp) |
| `wallarm_rule_vpatch` | attack_types (expanded: one resource per type) |
| `wallarm_rule_uploads` | file_type |
| `wallarm_rule_ignore_regex` | regex_id OR regex_rule (cross-reference) |
| `wallarm_rule_parser_state` | parser, state |
| `wallarm_rule_regex` | attack_type, regex, experimental |
| `wallarm_rule_file_upload_size_limit` | mode, size, size_unit |
| `wallarm_rule_rate_limit` | delay, burst, rate, rsp_status, time_unit |
| `wallarm_rule_credential_stuffing_point` | point, login_point, cred_stuff_type |
| `wallarm_rule_credential_stuffing_regex` | regex, login_regex, case_sensitive, cred_stuff_type |
| `wallarm_rule_mode` | mode (monitoring/safe_blocking/block/off/default) |
| `wallarm_rule_set_response_header` | header_name, header_mode, header_values |
| `wallarm_rule_overlimit_res_settings` | overlimit_time, mode |
| `wallarm_rule_graphql_detection` | mode, max_depth, max_value_size_kb, max_doc_size_kb, max_alias_size_kb, max_doc_per_batch, introspection, debug_enabled |
| `wallarm_rule_bruteforce_counter` | (counter only, no special fields) |
| `wallarm_rule_dirbust_counter` | (counter only, no special fields) |
| `wallarm_rule_bola_counter` | (counter only, no special fields) |
| `wallarm_rule_brute` | mode, threshold, reaction, enumerated_parameters |
| `wallarm_rule_bola` | mode, threshold, reaction, enumerated_parameters |
| `wallarm_rule_enum` | mode, threshold, reaction, enumerated_parameters |
| `wallarm_rule_rate_limit_enum` | mode, threshold, reaction |
| `wallarm_rule_forced_browsing` | mode, threshold, reaction |

### Multi-Value Expansion

- **`attack_types`** list → one `wallarm_rule_disable_attack_type` or `wallarm_rule_vpatch` per entry
  - Key format: `"${name}_${attack_type}"`
- **`stamps`** list → one `wallarm_rule_disable_stamp` per entry
  - Key format: `"${name}_${stamp}"`

### Cross-References

`wallarm_rule_ignore_regex` can reference a `wallarm_rule_regex` by name:
```hcl
{ name = "my_regex", resource_type = "wallarm_rule_regex", ... }
{ name = "ignore_it", resource_type = "wallarm_rule_ignore_regex", regex_rule = "my_regex", ... }
```
The module resolves `regex_rule` → `wallarm_rule_regex.this["my_regex"].regex_id`.

---

## 12. Action Block in Resources — Common Pattern

All 25 resource types use the same dynamic action block:

```hcl
dynamic "action" {
  for_each = try(local.rule_configs[each.key].action, [])
  content {
    type  = action.value.type == "" ? null : action.value.type
    value = try(action.value.value, "")
    point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
  }
}
```

Key details:
- `type = ""` → sent as `null` to the provider (instance, method, scheme, proto)
- Header names are always uppercased: `k == "header" ? upper(v) : v`
- Point is a single-key map — never multiple keys in one point

---

## 13. Key Implementation Files

| File | Purpose |
|------|---------|
| `modules/wallarm_rules/modules/custom_rules/main.tf` | Path expansion logic, action building, all 25 resource blocks |
| `modules/wallarm_rules/modules/custom_rules/variables.tf` | Full variable type definition with all fields |
| `modules/wallarm_rules/modules/custom_rules/EXAMPLES.tfvars` | Commented examples for all 25 resource types |
| `modules/wallarm_rules/modules/fp_rules/` | False-positive rules from hits (simpler, only disable_stamp + disable_attack_type) |
| `terraform-provider-wallarm/.../data_source_hits.go` | Go source: buildActionFromHit, locationToConditions, actionNameExtConditions |

---

## 14. Server-side Data Model (Action / Condition / Hint)

This section documents the API's database structure. The provider does not exercise these tables directly — it talks to HTTP endpoints — but knowing the shape clarifies why certain provider behaviors exist (`existingHintForAction`, `ConditionsHash`, the auto-cleanup of empty Actions).

### Action table (`actions`)

| Field | Notes |
|-------|-------|
| `id` | Unique Action ID (`action_id` in Terraform) |
| `clientid` | Tenant/client ID |
| `name` | Optional, validated `[A-Za-z0-9_.-]+`, unique per client |
| `conditions` | Ordered list of Condition objects |
| `conditions_hash` | SHA256 of serialized conditions; UNIQUE on `(clientid, conditions_hash)` |
| `conditions_count` | Denormalized count, 0–60 |
| `endpoint_path` / `endpoint_domain` / `endpoint_instance` | Cached scope fields |
| `endpoint` (bool) / `endpoint_url` / `method` | |
| `actual` / `internal` / `hidden` / `orphan` | Booleans |
| `endpoint_risk_score` | decimal(3,1), range 1–10 |
| `hits_count` / `request_stats` | Stats; `request_stats` is JSONB |
| `discovered_at` / `changed_at` / `created_at` / `updated_at` | Timestamps |

**Behavior:**

- **`find_or_create`** — same conditions reuse the existing Action (not duplicated). The provider's `existingHintForAction` relies on this contract.
- **`conditions_hash`** is the primary equality key. Two Actions with identical conditions produce the same hash. `ConditionsHash` in `wallarm/provider/hash.go` reproduces the Ruby `Action.calculate_conditions_hash` deterministically.
- **`nested`** — finds Actions whose conditions are a subset of this Action's, used for rule inheritance: a rule on `/api/*` applies to `/api/users`. The `with_nested` delete parameter cascades to parent Actions.
- After commit, the API triggers LOM compilation (`ScheduleLomCompilation`) and rule application (`ApplyAllForSingleAction`).

### Condition table (`action_conditions`)

Each row maps to one `action {}` block in Terraform.

| Field | Notes |
|-------|-------|
| `type` | `equal` / `iequal` / `regex` / `absent` (default `equal`) |
| `point` | JSON-serialized Proton Point array |
| `value` | Match value (absent for `type=absent`) |

**Behavior:**

- `iequal` values are **always lowercased** server-side (`before_validation :iequal_values_downcase`). The provider must lowercase these client-side to keep state consistent.
- `point` is stored as JSON and deserialized via `PointJsonDecoder` into `Proton::Point` objects.
- `to_h` output: `{ type: :equal, point: [...], value: "..." }` — this is the format the API returns and the provider processes.

### Hint table (`hints`)

| Field | Notes |
|-------|-------|
| `actionid` | FK → `actions.id` |
| `type` | Rule type string (`wallarm_mode`, `disable_stamp`, `bruteforce_counter`, ...) |
| `system` (bool) | System-managed flag |
| `data` | msgpack blob containing the full rule payload (`point`, `regex_id`, `clientid`, ...) |

The provider's `HintCreate`/`HintDelete` API calls map to inserts/deletes in this table; both trigger LOM recompilation.

### Action lifecycle

1. **FindOrCreate** — looks up by `conditions_hash + clientid`; creates transactionally if not found; handles race conditions with retry.
2. **Rules attached** — Hints are created pointing to the Action via `actionid` FK.
3. **LOM compilation** — `Action.with_payload` loads Actions, converts via `to_lom_action`, compiles into binary LOM.
4. **Delete cascade** — when all rules are removed from an Action, the API auto-cleans empty Actions: an Action with conditions and no rules becomes eligible for cleanup; an Action with no conditions persists. The provider only ever calls `HintDelete`, never `ActionDelete`, and relies on this auto-cleanup.
