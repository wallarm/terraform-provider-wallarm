# Design: Deduplicated Hits-to-Rules Cache in `wallarm_hits_index`

## Problem

Multiple `request_id`s can produce identical aggregated hit data (same action + same point groups + same stamps/attack_types). The current implementation stores a separate `terraform_data.rules_cache` entry per request_id, causing 100% data duplication when hits are identical. With 100 identical request_ids, there are 100 identical cache entries in Terraform state but only 1 set of rules.

Additionally, the current design places critical data processing logic (grouping, deduplication, merging) in HCL locals, which can be modified unintentionally. This logic should reside in the provider.

## Design Principles

1. **No redundant caching** -- identical rule data is stored once regardless of how many request_ids produced it.
2. **Provider does the heavy lifting** -- data processing, grouping, deduplication, and merge logic live in Go.
3. **HCL is a thin client** -- passes data to/from the provider, expands rules using standard Terraform operations.
4. **Single resource for metadata and data** -- `wallarm_hits_index` is the single source of truth.
5. **Automatic cleanup** -- removing a request_id cleans up all related data when no other request_id references it.

## Architecture

### Single Resource: `wallarm_hits_index`

`wallarm_hits_index` stores both the index (request_id tracking) and the deduplicated rule cache. No `terraform_data` needed.

#### Schema

**Inputs:**

| Attribute | Type | Description |
|-----------|------|-------------|
| `client_id` | `number`, Optional | Client ID (uses provider default if null) |
| `request_ids` | `set(string)`, Required | All active request_ids to track |
| `new_entries` | `map(string)`, Optional | Map of `request_id -> aggregated_json` for newly fetched hits. Empty on steady-state plans. |

**Computed Outputs:**

| Attribute | Type | Description |
|-----------|------|-------------|
| `cached_request_ids` | `string` | Comma-separated list of processed request_ids (used to gate `data.wallarm_hits`) |
| `cache` | `string` | JSON map of `action_hash -> aggregated_json`. Deduplicated -- one entry per unique action. |
| `request_to_action` | `string` | JSON map of `request_id -> action_hash`. Cross-reference for traceability and cleanup. |

### Data Model

**Cross-reference (internal to provider state):**

```
request_to_action: {
  "req_abc123" -> "48c0e969"
  "req_def456" -> "48c0e969"    // same action as abc123
  "req_xyz789" -> "a1b2c3d4"    // different action
}
```

**Cache (deduplicated by action_hash):**

```
cache: {
  "48c0e969" -> {
    "action_hash": "48c0e969",
    "action": [{"type": "iequal", "value": "example.com", "point": {"header": "HOST"}}],
    "groups": [
      {"key": "f1e2d3c4_xss", "point": [["header", "User-Agent"]], "stamps": [7994, 8001], "attack_type": "xss"},
      {"key": "a5b6c7d8_sqli", "point": [["get", "search"]], "stamps": [1234], "attack_type": "sqli"}
    ]
  },
  "a1b2c3d4" -> { ... different action ... }
}
```

Each group is keyed by `point_hash_prefix + "_" + attack_type` and has a single `attack_type` field. This scopes stamps to the attack type that detected them, improving traceability.

100 request_ids with the same action = 1 cache entry + 100 entries in request_to_action (~20 bytes each).

### Provider CRUD Logic

**Create:**
1. Process `new_entries`: for each `request_id -> aggregated_json`, compute content hash (SHA256 of aggregated JSON, truncated to 16 hex chars) to derive `action_hash`
2. Build `cache` map: `action_hash -> aggregated_json`
3. Build `request_to_action` map: `request_id -> action_hash`
4. Set `cached_request_ids` from `request_ids` input

**Update:**
1. Read existing `cache` and `request_to_action` from state
2. Process `new_entries`:
   - For each new entry, compute `action_hash`
   - If `action_hash` already in `cache`: **merge groups** -- add new groups, merge stamps into existing groups with same key (point_hash + attack_type)
   - If `action_hash` is new: add to `cache`
   - Map `request_id -> action_hash` in `request_to_action`
3. Sync with `request_ids` input:
   - Remove entries from `request_to_action` for request_ids not in the input set
   - Remove entries from `cache` where no request_id points to that action_hash
4. Set all computed outputs

**Read:**
1. Preserve existing `cache` and `request_to_action` from state
2. Re-sync `cached_request_ids` to match `request_ids` input (existing behavior)

**Delete:**
1. Clear state

### Group Merge Logic (Provider-Side)

When two request_ids produce the same `action_hash` but potentially different groups:

```
Existing cache["abc123"]: groups = [
  {key: "f1e2d3c4_xss", stamps: [7994], attack_type: "xss"}
]

New entry for same action_hash: groups = [
  {key: "f1e2d3c4_xss", stamps: [7994, 8001], attack_type: "xss"},  // same point+type, new stamp
  {key: "a5b6c7d8_sqli", stamps: [1234], attack_type: "sqli"}       // new point+type
]

After merge: cache["abc123"]: groups = [
  {key: "f1e2d3c4_xss", stamps: [7994, 8001], attack_type: "xss"},  // stamps merged
  {key: "a5b6c7d8_sqli", stamps: [1234], attack_type: "sqli"}       // new group added
]
```

Merge rules:
- Match groups by `key` (point_hash prefix + attack_type)
- Stamps: union of both lists, sorted, deduplicated
- Point and attack_type: take from either (identical for same key)
- New groups (key not in existing): add

### Content Hashing

The `action_hash` used as cache key is derived from the aggregated JSON content. The `buildAggregatedJSON` function already produces deterministic output (sorted group keys, sorted stamps/attack_types). The provider computes SHA256 of this string and truncates to 16 hex chars.

However, note that the `aggregated` output from `data.wallarm_hits` already includes `action_hash` (first 8 chars of the Ruby-compatible `ConditionsHash`). This is the **action identity** -- same host+path+pool = same hash. The content hash of the full JSON would be different (includes groups data). 

**Decision:** Use the `action_hash` from the aggregated JSON (the action identity), NOT a hash of the full content. This is because:
- Two request_ids with the same action but different groups should merge into the same cache entry
- The action_hash identifies the rule scope; groups within are merged
- This matches the cross-reference semantics: `request_id -> action_hash` means "this request produced rules for this action scope"

The provider extracts `action_hash` from the aggregated JSON by parsing it.

## HCL Flow

### Example Module (`examples/hits-to-rules/main.tf`)

```hcl
# ---- Gating ----

resource "wallarm_hits_index" "this" {
  client_id   = var.client_id
  request_ids = keys(var.request_ids)

  new_entries = {
    for id in local._new_request_ids :
    id => data.wallarm_hits.new[id].aggregated
  }
}

locals {
  _cached = toset(compact(split(",", wallarm_hits_index.this.cached_request_ids)))

  _new_request_ids = toset([
    for id in keys(var.request_ids) : id
    if !contains(local._cached, id)
  ])

  _request_configs = {
    for id, cfg_json in var.request_ids :
    id => jsondecode(cfg_json)
  }
}

# ---- Data Source (only for new request_ids) ----

data "wallarm_hits" "new" {
  for_each         = local._new_request_ids
  client_id        = var.client_id
  request_id       = each.key
  mode             = try(local._request_configs[each.key].mode, var.default_mode)
  attack_types     = try(local._request_configs[each.key].attack_types, [])
  rule_types       = try(local._request_configs[each.key].rule_types, [])
  include_instance = var.include_instance
}

# ---- Expand cached data into rules ----

locals {
  _cache = try(jsondecode(wallarm_hits_index.this.cache), {})

  # Flatten all groups from all cached actions into a single map.
  # Group key is already globally unique: point_hash_prefix + attack_type.
  # Action hash prefix is added for namespace isolation across actions.
  _groups = merge([
    for ah, agg_json in local._cache : {
      for g in jsondecode(agg_json).groups :
      "${ah}_${g.key}" => {
        action      = jsondecode(agg_json).action
        point       = g.point
        stamps      = g.stamps
        attack_type = g.attack_type
      }
    }
  ]...)

  # Expand stamps: one rule per group per stamp.
  stamp_rules = merge([
    for gk, g in local._groups : {
      for s in g.stamps :
      "${gk}_${s}" => {
        stamp  = s
        point  = g.point
        action = g.action
      }
    }
  ]...)

  # Expand attack types: one rule per group (attack_type is singular).
  attack_type_rules = {
    for gk, g in local._groups :
    gk => {
      attack_type = g.attack_type
      point       = g.point
      action      = g.action
    }
  }
}

# ---- Rule Resources ----

resource "wallarm_rule_disable_stamp" "this" {
  for_each             = local.stamp_rules
  client_id            = var.client_id
  comment              = "Managed by Terraform"
  variativity_disabled = true
  stamp                = each.value.stamp
  point                = each.value.point

  dynamic "action" {
    for_each = each.value.action
    content {
      type  = action.value.type
      value = action.value.value
      point = action.value.point
    }
  }
}

resource "wallarm_rule_disable_attack_type" "this" {
  for_each             = local.attack_type_rules
  client_id            = var.client_id
  comment              = "Managed by Terraform"
  variativity_disabled = true
  attack_type          = each.value.attack_type
  point                = each.value.point

  dynamic "action" {
    for_each = each.value.action
    content {
      type  = action.value.type
      value = action.value.value
      point = action.value.point
    }
  }
}
```

### Data Flow (Steady State -- Adding a New Request ID)

```
1. User adds request_id to var.request_ids
2. HCL: _new_request_ids = {new_id}  (not in cached_request_ids)
3. data.wallarm_hits.new["new_id"] fetches from API
   -> returns aggregated JSON with action_hash + groups
4. wallarm_hits_index Update receives new_entries = {"new_id": aggregated_json}
   -> extracts action_hash from JSON
   -> if action_hash in cache: MERGE groups
   -> if action_hash new: ADD to cache
   -> maps new_id -> action_hash in request_to_action
   -> updates cached_request_ids
5. HCL: _cache now has updated/new entry
   -> _groups expanded
   -> stamp_rules / attack_type_rules computed
   -> only genuinely new rules trigger resource creation
   -> existing rules see no change (same stamp + point + action)
```

### Data Flow (Cleanup -- Removing a Request ID)

```
1. User removes request_id from var.request_ids
2. wallarm_hits_index Update:
   -> removes request_id from request_to_action
   -> checks: does any other request_id point to this action_hash?
   -> if yes: cache entry preserved (other request_ids still use it)
   -> if no: cache entry REMOVED
3. HCL: _cache shrinks (if action_hash removed)
   -> _groups shrinks
   -> stamp_rules / attack_type_rules shrinks
   -> removed rules destroyed via for_each
```

### Data Flow (First Run -- Double Apply)

```
Apply 1: (empty state)
  -> wallarm_hits_index created with request_ids, new_entries = {}
  -> cached_request_ids set (all IDs marked as cached)
  -> cache = {}, request_to_action = {}
  -> no data.wallarm_hits triggered (new_entries empty, _new_request_ids unknown)

Apply 2: (hits_index in state)
  -> cached_request_ids known from state
  -> _new_request_ids = all request_ids (none fetched yet)
  -> data.wallarm_hits fetches for all
  -> new_entries populated -> provider merges into cache
  -> rules created
```

Wait -- this doesn't work. On Apply 1, `cached_request_ids` gets set to ALL request_ids (via `syncCachedRequestIDs`). Then on Apply 2, `_new_request_ids` is empty because all IDs are already "cached" -- but no data was actually fetched.

**Fix:** `cached_request_ids` should only include request_ids that have data in the cache (i.e., appear in `request_to_action`), not all request_ids from the input set.

**Revised `syncCachedRequestIDs` logic:**

```
cached_request_ids = request_ids that exist in request_to_action
```

This means:
- Apply 1: `request_to_action` is empty -> `cached_request_ids` is empty -> `_new_request_ids` = all IDs. But `data.wallarm_hits` can't run because `cached_request_ids` is `(known after apply)`.
- Apply 2: `cached_request_ids` is empty (known, from state) -> `_new_request_ids` = all IDs -> data source fetches -> `new_entries` populated -> provider stores in cache -> `cached_request_ids` updated to include fetched IDs.

This preserves the double-apply on first run (inherent to the `for_each` / computed value pattern) but correctly gates subsequent applies.

## Edge Cases

### Same action, new groups from different request_id

Request_id A produced 2 groups (xss at User-Agent, sqli at search param). Request_id B (same action) produces 3 groups (xss at User-Agent with a new stamp, sqli at search param, rce at search param).

Provider merges: existing 2 groups have stamps merged, new group (rce at search param) added. Result: 3 groups. Only genuinely new rule resources are created — existing stamp/attack_type rules see no change.

### Request_id removed, shared action_hash

Request_id A and B share action_hash. A is removed. Provider removes A from `request_to_action` but keeps cache entry (B still references it). No rules destroyed.

### Request_id removed, sole reference to action_hash

Request_id A is the only reference to action_hash X. A is removed. Provider removes A from `request_to_action`, sees no references to X, removes X from cache. Rules for X are destroyed.

### Re-adding a previously removed request_id

Request_id A was removed (and its action_hash cleaned up). User re-adds A. It's not in `cached_request_ids` (was removed), so `data.wallarm_hits` fetches again. Data re-enters cache. Rules recreated.

## Files to Modify

| File | Change |
|------|--------|
| `wallarm/provider/resource_hits_index.go` | Add `new_entries`, `cache`, `request_to_action` schema fields. Add merge logic, cleanup logic, revised `syncCachedRequestIDs`. |
| `wallarm/provider/data_source_hits.go` | Update `aggregatedGroup` struct: `attack_type` (singular) replaces `attack_types` list. Update `groupHitsForRules` to group by `point_hash + attack_type`. Update `buildAggregatedJSON` accordingly. |
| `wallarm/provider/provider.go` | No change (resource already registered) |
| `examples/hits-to-rules/main.tf` | Remove `terraform_data.rules_cache`. Read from `wallarm_hits_index.this.cache`. Simplify locals — `attack_type_rules` becomes a direct map (no inner loop). |
| `docs/guides/hits_to_rules.md` | Update How It Works, remove terraform_data references |
| `docs/resources/hits_index.md` | Document new attributes |
| `wallarm/provider/resource_hits_index_test.go` | Add tests for merge, cleanup, cross-reference |
| `wallarm/provider/data_source_hits_test.go` | Update tests for new group structure |

## State Size

| Scenario | Before | After |
|----------|--------|-------|
| 100 identical request_ids (~500B each) | 100 terraform_data entries = 50KB | 1 cache entry + 100 x 20B refs = ~2.5KB |
| 10 unique + 90 duplicates | 100 terraform_data entries = 50KB | 10 cache entries + 100 x 20B refs = ~7KB |
| 1 request_id | 1 terraform_data entry = 500B | 1 cache entry + 1 ref = ~520B |

## Testing

1. **Unit tests** for merge logic: same action + overlapping groups, same action + new groups, different actions
2. **Unit tests** for cleanup: remove request_id with shared action_hash, remove sole reference
3. **Unit tests** for `cached_request_ids` sync: only includes request_ids in `request_to_action`
4. **Acceptance test**: add 2 request_ids with same action -> verify 1 cache entry, correct rules
5. **Acceptance test**: remove 1 request_id -> verify cache preserved (shared action)
6. **Acceptance test**: remove last request_id for an action -> verify cache cleaned up, rules destroyed
