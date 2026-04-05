# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform provider for Wallarm's unified API security platform, built with `terraform-plugin-sdk/v2` and the `wallarm-go` API client. Manages Wallarm rules, integrations, IP lists, applications, tenants, and nodes via the Wallarm API.

## Build & Development Commands

```bash
make build        # Build provider binary locally (sideloaded)
make install      # Install to $GOPATH/bin
make test         # Unit tests (-timeout=30s -parallel=4 -race)
make testacc      # Acceptance tests (requires TF_ACC=1, WALLARM_API_TOKEN, WALLARM_API_HOST)
make lint         # Run golangci-lint
make vet          # Run go vet
make fmt          # Format code
make fmtcheck     # Verify formatting
```

Run a single test:
```bash
go test -v -run TestAccRuleWmodeCreate_Basic ./wallarm/provider/ -timeout=120m
```

Acceptance tests require environment variables: `WALLARM_API_TOKEN`, `WALLARM_API_HOST`, and `TF_ACC=1`.

**Local wallarm-go development**: The `go.mod` has a `replace` directive pointing to `../wallarm-go`. This gets commented out by `go mod tidy` or pre-commit hooks — re-enable it before building if you see "undefined: wallarm.APIError" or similar errors.

## Architecture

### Entry Point
- `main.go` initializes the Terraform plugin server; supports `-debug` flag.

### Provider Configuration (`wallarm/provider/provider.go`, `config.go`)
- Auth via `WALLARM_API_TOKEN` env var (or legacy UUID/secret pair)
- API host defaults to `https://api.wallarm.com`
- `Config.Client()` creates a `wallarm.API` client, wrapped in `CachedClient`, stored in `ProviderMeta`
- `ProviderMeta` struct holds `Client`, `DefaultClientID`, and `RequireExplicitClientID`
- Use `apiClient(m)` to get the `wallarm.API` from meta, `retrieveClientID(d, m)` for client ID resolution

### Resource Pattern
All rule resources follow a consistent pattern:
1. Define resource-specific schema fields, then merge with `commonResourceRuleFields` via `lo.Assign()`
2. CRUD functions use `apiClient(m)` and `retrieveClientID(d, m)` helpers
3. Resource IDs use slash-separated format: `{clientID}/{actionID}/{ruleID}` (some resources add a 4th segment, e.g. `/{mode}`)
4. Read operations share `resourcerule.ResourceRuleWallarmRead()` from `wallarm/common/resourcerule/`
5. Import is supported via `resourceWallarm*Import` functions — all rule resources support import
6. All CRUD functions use `CreateContext`/`ReadContext`/`UpdateContext`/`DeleteContext` with `diag.Diagnostics`

### Mitigation Controls vs Rules

The provider manages two categories of rule-like resources that differ in scope and behavior:

**Mitigation Controls** — session-based, real-time threat mitigation. In the Wallarm Console UI these have a separate interface from regular rules. Their logic operates on sessions/counters rather than individual requests:

| Resource | Protection type |
|----------|----------------|
| `wallarm_rule_mode` | Real-time blocking mode |
| `wallarm_rule_graphql_detection` | GraphQL API protection |
| `wallarm_rule_enum` | Enumeration attack protection |
| `wallarm_rule_bola` | BOLA/IDOR protection |
| `wallarm_rule_forced_browsing` | Forced browsing protection |
| `wallarm_rule_brute` | Brute force protection |
| `wallarm_rule_rate_limit_enum` | DoS protection (rate limiting) |
| `wallarm_rule_file_upload_size_limit` | File upload restriction policy |

Mitigation controls with `reaction` blocks (`brute`, `bola`, `enum`, `rate_limit_enum`, `forced_browsing`) use threshold/reaction schemas from `default.go` and have API-enforced constraints on valid reaction keys per mode (see "Threshold/Reaction Rules — API Validation" section below).

Counter resources (`wallarm_rule_bruteforce_counter`, `wallarm_rule_dirbust_counter`, `wallarm_rule_bola_counter`) define the request parameters used for counting. They work with **triggers** (`wallarm_trigger`), not directly with mitigation controls — a trigger references a counter and defines threshold/reaction logic.

**Rules** — request-level, applied per-request during traffic analysis. These include virtual patches, detection tuning, data masking, parser control, and false positive suppression. They follow the standard `commonResourceRuleFields` pattern and share the `resourcerule.ResourceRuleWallarmRead()` Read function.

Both categories share the same underlying API model (Action + Hint), use the same provider patterns (schema merge via `lo.Assign()`, `CustomizeDiff` for action scope), and support import. The distinction is in their product semantics, not their provider implementation.

### Shared Schema & Defaults (`wallarm/provider/default.go`)
- `defaultPointSchema` — 2D list-of-lists point structure
- `commonResourceRuleFields` — shared fields (rule_id, client_id, comment, active, title, rule_type, etc.)
- `APIListLimit = 500` — default limit for API list requests (defined in `constants.go`)
- `mitigation` field: Optional+Computed, never sent to API (API rejects invalid values)
- `variativity_disabled` field: `Optional: true, Default: true`. Read returns actual API value. Schema default triggers Update on import (same pattern as `comment`).
- `comment` field: `Optional: true, Default: "Managed by Terraform"`. Read returns actual API value. Schema Default provides value when not set in config → triggers Update on import. **`title` and `comment` are independent fields** — `title` is the UI "Title", `comment` is the UI "Description". Do not conflate them.

### Rule Defaults (IMPORTANT — do not change)

All Terraform-managed rules SHOULD have:
1. `variativity_disabled = true`
2. `comment = "Managed by Terraform"` (when comment is empty)

Schema and Read behavior:
- `variativity_disabled`: `Optional: true, Default: true`. Read returns actual API value. When not in config, schema default `true` creates a diff if API has `false` → triggers Update on import.
- `comment`: `Optional: true, Default: "Managed by Terraform"`. Read returns actual API value. Schema Default provides value when not set in config.
- On Create: both fields get Defaults from schema. `variativity_disabled` is also hardcoded `true` in each resource's Create function.
- On Import: Read returns actual API values. Schema defaults create diffs → Update sets both `variativity_disabled=true` and `comment="Managed by Terraform"` in the same apply.

### Common Utilities (`wallarm/common/`)
- `common.go` — string conversion helpers
- `const.go` — constants for point keys and match types
- `resourcerule/resource_rule.go` — shared CRUD logic, `ExpandSetToActionDetailsList()` for action expansion
- `mapper/tftoapi/` — Terraform schema to API format conversion
- `mapper/apitotf/` — API response to Terraform schema conversion

### API Limits & Constants (`wallarm/provider/constants.go`)

All API pagination and batch size limits are centralized in one file:

| Constant | Value | Purpose |
|----------|-------|---------|
| `IPListPageSize` | 1000 | IP list groups per API call (pagination) |
| `IPListMaxSubnets` | 1000 | Max subnet values per IP list resource |
| `IPListCacheMaxRetries` | 3 | Cache refresh retries on empty result after Create |
| `IPListCacheRetryDelay` | 3s | Wait between cache refresh retries |
| `APIListLimit` | 500 | Default limit for rule/user/app list requests |
| `HintBulkFetchLimit` | 500 | Hints per page during cache lazy pagination |
| `HitFetchBatchSize` | 500 | Hits per API call in data source |

wallarm-go methods accept `limit` as a parameter — the provider passes the value from these constants. No hardcoded limits in wallarm-go.

### Key Reference
`.claude/REFERENCE.md` is the authoritative internal reference for action conditions, point keys, match types, and path-to-action expansion logic. Consult it when modifying rule resources.

## Hint Cache (`wallarm/provider/hint_cache.go`)

The `CachedClient` wraps `wallarm.API` and intercepts `HintRead` calls to avoid redundant API requests during `terraform plan`/`refresh`.

**How it works — lazy pagination:**
1. On first `HintRead` with a single-ID filter, `GetOrFetch()` fetches one page (500 hints) from the API and checks if the requested ID is in the response
2. If found → return immediately (1 API call for potentially all managed rules if they fit on page 1)
3. If not found → fetch next page, check again, repeat until found or all pages exhausted
4. Subsequent reads for IDs already cached → zero API calls (cache hit)
5. Non-cacheable queries (multi-ID, ActionID/Type filters) pass through to the real API
6. `LoadAll()` fetches all remaining pages — used by `data.wallarm_rules` which needs the complete set

**Mutation handling:**
- `HintCreate` → `Insert` (adds to cache, no invalidation)
- `HintUpdateV3` → `Insert` (updates in cache, no invalidation)
- `HintDelete` → passthrough to API (no cache invalidation — cache starts fresh each plan cycle)

**Performance:**
- 0 managed rules → 0 API calls (no Reads triggered)
- 5 managed rules on page 1 → 1 API call (1 page fetch serves all 5)
- 100 managed rules spread across pages → N page fetches (stops when last needed ID found)
- 3000 managed rules → ~15 page fetches (same as before, fetches everything)
- Create N new rules → 0 extra fetch calls (Insert keeps cache warm)

**Thread safety**: All cache access is protected by `sync.Mutex`.

## Credential Stuffing Cache (`wallarm/provider/credential_stuffing_cache.go`)

Caches `GET /v4/clients/{clientID}/credential_stuffing/configs` responses. The API returns all configs in a single call (no pagination). Stored in `ProviderMeta.CredentialStuffingCache`.

- `GetOrFetch(client, clientID, ruleID)` — first call fetches all configs, subsequent calls serve from cache
- `LoadAll(client, clientID)` — returns all configs (uses cache if loaded, otherwise fetches)
- `Invalidate()` — clears cache, called on Delete
- Used by `resource_rule_credential_stuffing_point` and `resource_rule_credential_stuffing_regex` (Read, Import)
- Used by `data_source_rules` for merging credential stuffing configs into the rules list
- Create invalidates the cache (new rule not in cache; next Read re-fetches)
- Update (HintUpdateV3 for comment/variativity) does NOT invalidate (different API endpoint)

## HCL Generator (`wallarm/provider/resource_hcl_generator.go`)

The `wallarm_rule_generator` resource generates HCL config files from cached rule data or existing API rules.

**Two source modes:**
- `source = "rules"` (default) — generates HCL from pre-built rules via `rules_json` (e.g., from the hits-to-rules workflow `_all_rules` local)
- `source = "api"` — fetches existing rules directly from the Wallarm API via `HintRead`, groups by action, generates HCL

**Key fields:**
- `output_dir` — directory for generated files
- `output_filename` — filename when `split = false` (defaults to `{prefix}_rules.tf`)
- `rules_json` — JSON-encoded list of pre-built rules (required when `source = "rules"`)
- `rule_types` — filter by type (default: `["disable_stamp", "disable_attack_type"]`)
- `split` — one file per rule (true) or all in one file (false, default)
- `resource_prefix` — prefix for resource names (default: `"fp"` for rules, `"rule"` for api)
- `source` — `"rules"` or `"api"`

Each rule in `rules_json` carries its own `action` block — rules from different action scopes are correctly generated with their respective action conditions. The `expandedRule` struct has an `Actions` field for per-rule action conditions.

**Templates** (`hcl_generator_templates.go`): use `hclwrite` + `cty` for proper HCL generation with correct escaping of special characters in point values (XML namespaces, URIs, etc.). This handles cases where `terraform plan -generate-config-out` fails.

## wallarm-go Client Library

The `wallarm-go` library (`../wallarm-go`) is the HTTP client for the Wallarm API.

### Error Handling
- `APIError` struct with `StatusCode` and `Body` fields
- Use `errors.As(err, &apiErr)` to check specific status codes
- `isNotFoundError(err)` helper in the provider for 404 checks

### Retry Logic (transport level)
All API calls automatically retry on transient errors:

| Status | Delay | Max retries | Total wait |
|--------|-------|-------------|------------|
| **423** (Rules locked) | 5s fixed | 12 | 60s |
| **5xx** (Server error) | 10s fixed | 12 | 120s |
| **429** (Rate limit) | Exponential backoff | 12 | ~30s |

### Gzip Compression
All API requests include `Accept-Encoding: gzip`. Responses with `Content-Encoding: gzip` are decompressed transparently in `wallarm.go`. ~19x reduction in response payload size (142KB → 7.5KB for a typical hints page).

### Pagination Fix (Critical)
All paginated API methods in wallarm-go MUST set `response.Body.Objects = nil` before each `json.Unmarshal`. Without this, Go's JSON decoder reuses the backing array from the previous page's slice, causing page N to silently overwrite page N-1's data in the result. This was the root cause of all IP list matching bugs.

### HTTP Logging (`wallarm/provider/logging_transport.go`)
- Enabled by `WALLARM_API_CLIENT_LOGGING=true` (or `api_client_logging` provider attribute)
- SDK-style formatted request/response logging with token masking
- Handles gzip decompression before logging for readable response bodies
- Captured by `TF_LOG=DEBUG`

## IP List Resources

IP list resources (`wallarm_allowlist`, `wallarm_denylist`, `wallarm_graylist`) are the most complex resources in the provider due to API behavior.

### API Endpoints
- **Create**: `POST /v1/blocklist/clients/{id}/access_rules` — returns `{"status": 200}` with NO ID
- **Read**: `GET /v1/blocklist/clients/{id}/groups` — supports `filter[rule_type][]`, `filter[list]`, and JSON `filter` with `query` param
- **Delete**: `DELETE /v1/blocklist/clients/{id}/groups` — deletes by rule_type + group IDs
- **Search**: `GET /v1/blocklist/clients/{id}/groups?filter={"rule_type":["subnet"],"query":"1.2.3.4"}` — find specific entry

### API Grouping Behavior (Critical)

| Rule type | API grouping | Example |
|-----------|-------------|---------|
| `location` (country) | All values in **1 group** with 1 ID | `values: ["US","UK","DE",...]` |
| `datacenter` | All values in **1 group** with 1 ID | `values: ["aws","gce","azure"]` |
| `proxy_type` | All values in **1 group** with 1 ID | `values: ["TOR","VPN","PUB"]` |
| `subnet` (IPs) | **1 group per IP**, each with own ID | `values: ["1.2.3.4/32"]` |

Bare IPs are normalized by the API: `1.2.3.4` → `1.2.3.4/32`

### Constraints
- **Max 1000 subnets** per IP list resource (`IPListMaxSubnets`)
- **One rule type per resource** — enforced by `ConflictsWith` (can't mix ip_range with country/datacenter/proxy_type)
- **One API page** for subnets — with `IPListPageSize=1000` and max 1000 subnets, all entries fit in a single bulk fetch

### Resource ID & Computed Attributes
- Resource ID: synthetic hash `{clientID}/{listType}/{ruleType}/{valuesHash}` (e.g., `8649/deny/subnet/3df19f29`)
- `address_id` — list of API group entries (rule_type, value, ip_id)
- `entry_count` — number of config values found in API (not group count — for grouped types, 3 proxy values = 3, not 1)
- `untracked_count` — number of config values NOT found in API
- `untracked_ips` — list of specific config values not found (for user to review/remove from config)

### IP List Cache (`wallarm/provider/ip_list_cache.go`)

Shared at `ProviderMeta` level. Maps normalized IP values → API group IDs via per-rule-type bulk fetches. Eliminates per-value scanning and prevents cross-resource contamination.

```
ProviderMeta.IPListCache
├── entries: map[listType]map[normalizedValue]IPCacheEntry
├── createMu: map[listType]*sync.Mutex       — serializes Creates per list type
├── EnsureLoaded(client, listType, clientID)  — lazy first load (all rule types)
├── RefreshRuleTypes(client, listType, clientID, ruleTypes) — fetch specific rule types only
├── RefreshUntilFound(client, ..., values, ruleTypes, retries, delay)
├── LookupMany(listType, values) → (found, missing)
└── Invalidate(listType)
```

**Per-rule-type fetching**: Create for a subnet resource only fetches subnet groups from the API — doesn't pull countries/datacenters/proxy. Existing cache entries for other rule types are preserved (merge, not replace).

**Create serialization**: `LockCreate(listType)` / `UnlockCreate(listType)` — per-list-type mutex prevents concurrent Creates from racing on cache refresh. Denylist resources queue up. Allowlist + denylist still run in parallel.

**Normalization**: subnets indexed by both raw (`"1.2.3.4/32"`) and bare IP (`"1.2.3.4"`). Others as-is.

**Log format**: `loaded 1001 groups (2003 map entries) for list=gray [subnet=1000, datacenter=1] in 808ms`

### CRUD Flow

**Create:**
1. Acquire per-list-type Create lock (serializes concurrent Creates)
2. Single POST with all values
3. `RefreshUntilFound` — fetch only this resource's rule types from API
   - Retry only if NONE of this resource's values are found (API empty after Create)
   - Once some values appear → stop retrying
   - `IPListSearch` per missing value (pagination boundary stragglers, typically 0-1)
4. Build `address_id` from found entries, set `entry_count`, `untracked_count`, `untracked_ips`
5. Release Create lock

**Read:**
1. `cache.EnsureLoaded()` — bulk fetch all rule types if not already loaded (shared across resources)
2. `cache.LookupMany()` — O(1) per config value, returns found entries + missing list
3. Build `address_id` from found entries
4. If nothing found and `address_id` was previously populated → remove from state (deleted externally)
5. If nothing found and `address_id` was empty → keep in state (not propagated yet)

**Update (subnet diff — ip_range changed, metadata unchanged):**
1. Compute diff: `oldIPs - newIPs` = removed, `newIPs - oldIPs` = added
2. For removed IPs: `cache.LookupMany` → get group IDs. Fallback to `IPListSearch` for any not in cache
3. Single DELETE with all collected group IDs
4. Single CREATE with all added IPs
5. Invalidate cache. Don't update `address_id` — next refresh handles it

**Update (grouped types or metadata changed):**
1. `deleteByAddrIDs` — delete using group IDs from state (uses old IDs, not new schema values)
2. Invalidate cache
3. `resourceWallarmIPListCreate` — recreate with new values

**Delete (terraform destroy):**
1. `deleteByAddrIDs` — single DELETE with all group IDs from `address_id` (populated by Read during refresh)
2. Cleanup: refresh cache → `LookupMany` config values → DELETE any found stragglers
3. Invalidate cache

### Cache Lifecycle

| Event | Cache Action |
|-------|-------------|
| First Read for a list type | `EnsureLoaded` → bulk fetch all rule types |
| After Create | `RefreshUntilFound` → fetch specific rule types only + `IPListSearch` fallback |
| After Delete | Cleanup refresh → `Invalidate` |
| After diff Update (add/remove) | `Invalidate` |

### Key Functions (`resource_ip_list.go`)
- `ipListConfigValues()` — extracts config values (ip_range/country/datacenter/proxy_type) from schema
- `cacheEntriesToAddrIDs()` — converts cache entries to `address_id` schema format, sorted by group ID
- `ipListSubnetDiffUpdate()` — targeted add/delete for subnet changes using cache lookups
- `deleteByAddrIDs()` — delete using group IDs from state (used by Update for grouped types)
- `IPListSearch()` (wallarm-go) — targeted search with `query` filter param, fallback for pagination boundary misses

### Import

IP list resources support `terraform import` with two ID formats:

```bash
# Grouped types (country/datacenter/proxy): import by group ID
terraform import wallarm_denylist.countries 8649/52000393

# Subnets: import all IPs with same expiration (must all share same app scope)
terraform import wallarm_denylist.ips 8649/subnet/1804809600

# Subnets: import by expiration AND application scope (when different apps exist)
terraform import wallarm_denylist.ips_app1 8649/subnet/1804809600/apps/1,3
terraform import wallarm_denylist.ips_all 8649/subnet/1804809600/apps/all
```

The importer groups subnets by `(expired_at, application_ids)`. If the simple format is used but entries have mixed app scopes, it errors with guidance to use the `/apps/{appIDs}` format. `{appIDs}` is sorted comma-separated IDs or `all`. Strips `/32` from bare IPs and computes the synthetic resource ID hash.

### Data Source: `wallarm_ip_lists`

Reads all existing IP list entries for a given list type. Used by the import module and for querying existing entries.

```hcl
data "wallarm_ip_lists" "deny" {
  list_type = "denylist"  # allowlist, denylist, or graylist
}
```

Output: `entries` list with id, rule_type, values, reason, expired_at, created_at, application_ids, status. Expired entries are filtered out.

### wallarm-go IP List Methods

All methods accept `limit` as parameter — provider passes `IPListPageSize` from `constants.go`. All paginated methods set `response.Body.Objects = nil` before each `json.Unmarshal` to prevent slice reuse bugs.

- `IPListRead(listType, clientID, limit)` — fetch all groups (all rule types), paginated
- `IPListReadByRuleType(listType, clientID, ruleTypes, limit)` — fetch groups filtered by specific rule types, paginated. Primary method used by the cache.
- `IPListSearch(listType, clientID, ruleType, query)` — search for specific value with JSON `filter` + `query` param (returns exact match, limit=1). Fallback for pagination boundary misses.
- `IPListCreate(clientID, params)` — create entries
- `IPListDelete(clientID, rules)` — delete by rule_type + group IDs

## Hits, Attacks & Automatic False Positive Rules

### Domain Model

A **hit** represents a single detected threat within an HTTP request. One proxied request can produce multiple hits — different attack vectors in different request parameters generate separate hit objects. Each hit contains the request payload (full or partial), detected attack type metadata, and the point (request parameter) where the threat was found.

Hits sharing the same HTTP request are linked by `request_id`. Hits from the same attack campaign share an `attack_id`. An attack is a logical grouping of related hits.

**IMPORTANT: Hits are ephemeral** — they have a retention period and can be dropped from the API at any time. Rules generated from hits must be cached/persisted after the first fetch. Re-fetching hits on every plan would cause rules to be destroyed when their source hits expire. Any HCL module that creates rules from hits MUST store the hit data in state or on disk after the initial fetch.

### False Positive Workflow

Some detections are false positives (FPs) — legitimate requests misidentified as attacks. The provider supports an automated pipeline to create rules that suppress these FPs:

1. **Fetch**: The `wallarm_hits` data source retrieves all hits for given `request_id`(s) from the API
2. **Group by Action**: Hits are grouped by their HTTP request conditions (Host header + URI path). This is called an **Action** — the first grouping key. Hits must match on Host and path, not just `request_id`, as a strict guard against excessive rule creation. The goal is minimalistic, precise rules.
3. **Group by Point**: Within each action, hits are further grouped by **point** — the specific request parameter where the attack was detected (header, body field, query param, cookie, etc.)
4. **Generate Rules**: Two rule types are used for FP suppression:
   - **`disable_stamp`** — allows specific attack signatures (stamps) at a given point
   - **`disable_attack_type`** — allows specific attack types at a given point
5. **One resource per rule**: Each stamp and each attack_type is a separate Terraform resource (`wallarm_rule_disable_stamp` or `wallarm_rule_disable_attack_type`), matching the API 1:1. The `for_each` key is `{action_hash}_{point_hash}_{stamp}` or `{action_hash}_{point_hash}_{attack_type}`. When generating HCL files with `split = true`, each rule gets its own `.tf` file.

### Grouping Logic

- **Action match** (Host + path) is mandatory — hits with different actions are never merged
- **Point match** determines which parameters the rule applies to
- Hits within a group produce either a `disable_stamp` or `disable_attack_type` rule depending on the detection type
- The `attack_id` field on each hit enables expanding from individual request FPs to all related hits in the same attack campaign

### Data Source: `wallarm_hits`

Fetches hits from the Wallarm API and aggregates them into rule-ready structures.

**Input**: `request_id` (single string) + `mode` variable (`"request"` or `"attack"`). Called per-request_id via `for_each` in HCL.

**Two modes, same output structure:**
- **Request mode**: Fetch hits directly by `request_id` — produces rules for the specific request
- **Attack mode**: Fetch hits by `request_id`, then collect `attack_id` from each hit and fetch ALL related hits belonging to those attacks (filtered by type). Merges related hits with the original request data and processes through the same pipeline

**Fetch pipeline (attack mode):**
1. User provides `request_id`(s) as input
2. Fetch direct hits from API by `request_id`
3. Collect `attack_id` from each returned hit
4. Fetch all hits belonging to those `attack_id`s (filtered by allowed attack types)
5. Fetch in batches of 500 hits per request
6. Validate each batch — all hits must share the same action conditions (Host + path). Discard hits with different action hash
7. Group hits by point (stamps and attack_types aggregated per point group)
8. Build `aggregated` JSON output for caching in `terraform_data`

**Hit filtering — allowed attack types:**
`xss`, `sqli`, `rce`, `ptrav`, `crlf`, `redir`, `nosqli`, `ldapi`, `scanner`, `mass_assignment`, `ssrf`, `ssi`, `mail_injection`, `ssti`, `xxe`, `invalid_xml`

**API call pattern (hits by attack_id):**
```bash
POST /v1/objects/hit
{
  "filter": {
    "state": null,
    "type": ["xss","sqli","rce","xxe","ptrav","crlf","redir","nosqli","ldapi",
             "scanner","mass_assignment","ssrf","ssi","mail_injection","ssti"],
    "time": [[<start_ts>, <end_ts>]],
    "!state": "falsepositive",
    "security_issue_id": null,
    "clientid": <client_id>,
    "!experimental": true,
    "!aasm_event": true,
    "!wallarm_scanner": true,
    "attackid": ["<attack_id_1>", "<attack_id_2>"]
  },
  "limit": 500,
  "offset": 0,
  "order_by": "time",
  "order_desc": true
}
```

**Key filter fields:**
- `attackid` — list of attack IDs to fetch related hits
- `type` — restricted to the allowed attack types list above
- `!state: "falsepositive"` — excludes already-marked FPs
- `!experimental`, `!aasm_event`, `!wallarm_scanner` — excludes noise
- Pagination: `limit: 500`, increment `offset` per batch

**Aggregation rules:**
- Group by action hash (Host + path) — mandatory, strict match
- Within action: group by point (request parameter)
- Each unique point produces multiple rules: one `disable_stamp` per stamp and one `disable_attack_type` per attack type. Stamp groups and attack_type groups are separate (stamps are not attack-type-scoped).
- Output structure is identical for both request and attack modes

## Tenant Resource (`wallarm_tenant`)

- Resource ID is just `client_id` (integer)
- `client_id` field is Optional+Computed — can be set for existing tenants (multi-tenant scenarios) or auto-populated on create
- Import: `terraform import wallarm_tenant.foo {client_id}`
- Delete safety: disables tenant first, only permanently deletes if `prevent_destroy=false` AND `WALLARM_ALLOW_CLIENT_DELETE=1`
- `prevent_destroy` is a provider-side attribute (default true), not stored in API

## Node Resource (`wallarm_node`)

- Resource ID: `{client_id}/{node_id}`
- Import: `terraform import wallarm_node.foo {client_id}/{node_id}`
- Read matches by `node_id` (unique integer), not hostname
- Default application (`app_id=-1`) is protected from deletion

## Trigger Resource (`wallarm_trigger`)

- Resource ID: `{client_id}/{template_id}/{trigger_id}`
- Import: `terraform import wallarm_trigger.foo {client_id}/{template_id}/{trigger_id}`
- Import limitation: Read only populates `trigger_id` and `client_id` — full field population from API response is a TODO
- `comment`, `lock_time`, `lock_time_format` are `Optional+Computed` (no `Default` — defaults are applied in Go code to avoid "non-computed attribute" warnings)

## API Domain Model: Action & Condition

**CRITICAL: This section is the ONLY authoritative source for action condition structure. When constructing, reviewing, or testing action conditions — read this section first. Do NOT guess or approximate from memory.**

Rules in the Wallarm API are stored as **Action + Rule (Hint)** pairs. Understanding this model is essential for working with the provider.

### Action

An Action (`actions` table) represents a **scope** — a set of conditions that define where a rule applies. Multiple rules can share the same Action (same scope).

**Key fields:**
- `id` — unique Action ID (`action_id` in Terraform)
- `clientid` — tenant/client ID
- `name` — optional name (validated: `[A-Za-z0-9_.-]+`)
- `conditions` — ordered list of Condition objects (the action conditions)
- `conditions_hash` — hash of conditions for fast lookup
- `endpoint_path`, `endpoint_domain`, `endpoint_instance` — cached scope fields

**Behavior:**
- `find_or_create` — if an Action with the same conditions already exists for the client, it's reused (not duplicated). This is why the provider checks `existsAction` before creating.
- `conditions_hash` enables fast equality matching — two Actions with identical conditions produce the same hash.
- `nested` — finds Actions whose conditions are a subset of this Action's conditions (for rule inheritance).
- After commit, triggers LOM compilation (`ScheduleLomCompilation`) and rule application (`ApplyAllForSingleAction`).

### Condition

A Condition (`action_conditions` table) represents a single matching rule within an Action scope. Maps to one `action {}` block in Terraform.

**Key fields:**
- `type` — match type: `equal`, `iequal`, `regex`, `absent` (validated, default: `equal`)
- `point` — serialized as JSON, represents the request part to match (e.g., `["header", "HOST"]`, `["path", 0]`). Deserialized as `Proton::Point`.
- `value` — the value to match against. For `iequal` type, automatically downcased before save.

**Behavior:**
- `iequal` values are **always lowercased** (`before_validation :iequal_values_downcase`). This is why the provider must lowercase `iequal` values (e.g., domain names via HOST header).
- `point` is stored as JSON and deserialized via `PointJsonDecoder` into `Proton::Point` objects.
- `to_h` output: `{ type: :equal, point: [...], value: "..." }` — this is the format the API returns and the provider processes.

### Relationship

```
Action (scope)  ←──has_many──→  Conditions (action conditions)
Action (scope)  ←──has_many──→  Rules/Hints (wallarm_mode, vpatch, etc.)
```

Multiple rules sharing the same scope (same conditions) point to the same Action. Creating a rule with the same conditions as an existing one reuses the Action. Deleting the last rule under an Action may delete the Action itself.

## Wallarm User Roles & Multi-Tenancy

The provider operates under an API token scoped to a specific user role. The role determines what the provider can manage.

### Roles

| Role | Scope | Key Permissions |
|------|-------|-----------------|
| **Administrator** | Single account | Full access: manage nodes, rules, integrations, users, API tokens, filtration mode |
| **Analyst** | Single account | View/manage attacks, incidents, vulnerabilities, API inventory. View-only for triggers, rules. Personal API tokens only |
| **Read Only** | Single account | View-only access to most entities. Export IP lists |
| **API Developer** | Single account | View/download API inventory and specs only |
| **Deploy** | Single account | Create filtering nodes only, no Console access |
| **Global Administrator** | Multi-tenant | Same as Administrator but across technical tenant + all linked tenant accounts |
| **Global Administrator Extended** | Multi-tenant | Same as Global Administrator + can manage `disable_stamp` rules (FP suppression by signature) |
| **Global Analyst** | Multi-tenant | Same as Analyst but across technical tenant + all linked tenant accounts |
| **Global Read Only** | Multi-tenant | Same as Read Only but across technical tenant + all linked tenant accounts |

### Provider Implications

- Only **Administrator** and **Global Administrator** can manage users, integrations, and rules
- Only **Global Administrator Extended** can manage `disable_stamp` rules — standard Administrator/Global Administrator tokens will get 403 on `HintCreate` for this rule type
- Only **Global** roles can operate across tenant accounts — standard roles are limited to the technical tenant account
- `client_id` in the provider config or on individual resources determines which tenant account to target
- If `client_id` is omitted, the provider auto-detects it from the API token via `UserDetails()`
- `require_explicit_client_id` provider attribute (default false) forces all resources to specify `client_id` — safety net for Global Admin multi-tenant scenarios
- User management operations (create/delete) require admin-level tokens
- Tenant creation/deletion requires Global Administrator or equivalent elevated permissions (`WALLARM_EXTRA_PERMISSIONS`)

### Acceptance Test Env Vars

| Variable | Purpose |
|----------|---------|
| `WALLARM_API_TOKEN` | API token for authentication |
| `WALLARM_API_HOST` | API endpoint (e.g. `https://api.wallarm.com`) |
| `WALLARM_API_CLIENT_ID` | Target client/tenant ID |
| `WALLARM_EXTRA_PERMISSIONS` | Set to enable tests requiring elevated permissions (tenant CRUD) |
| `WALLARM_GLOBAL_ADMIN` | Set to skip tests that expect 403 errors (token has global admin role) |
| `WALLARM_ALLOW_CLIENT_DELETE` | Set to `1` to allow permanent tenant deletion (safety guard) |
| `WALLARM_API_CLIENT_LOGGING` | Set to `true` to enable HTTP request/response logging |

## Testing Conventions

- Test files are colocated with implementation in `wallarm/provider/`
- Acceptance tests use `resource.Test()` with `TestCase` structs containing `Steps`
- Config generators return HCL strings via `fmt.Sprintf`
- `generateRandomResourceName()` creates unique resource names
- `testAccPreCheck()` verifies credentials are set
- Destroy checks verify cleanup via `testAccCheckWallarm*Destroy` functions
- All rule resource tests use `ImportStateVerifyIgnore: []string{"rule_type"}` because `rule_type` is set during import but not during normal Read
- Tenant tests ignore `prevent_destroy` (provider-side only, not in API)
- Node tests ignore `token` (may not be returned on read after import)
- Data sources requiring elevated permissions (e.g., `security_issues`) must use `retrieveClientID(d, m)` not `d.Get("client_id")`

## SDK v2 Migration (Completed)

The provider has been fully migrated from terraform-plugin-sdk v1 patterns to v2:

### What was done
- All resources migrated from `Create`/`Read`/`Update`/`Delete` to `CreateContext`/`ReadContext`/`UpdateContext`/`DeleteContext` + `diag.Diagnostics`
- All importers migrated from `State:` to `StateContext:`
- Global `var ClientID int` replaced with `ProviderMeta` struct (`Client`, `DefaultClientID`, `RequireExplicitClientID`)
- `m.(wallarm.API)` replaced with `apiClient(m)` helper everywhere
- `retrieveClientID(d)` changed to `retrieveClientID(d, m)` with `ProviderMeta` delegation
- All test files updated: `testAccProvider.Meta().(wallarm.API)` → `testAccProvider.Meta().(*ProviderMeta).Client`
- Import state test steps added to all resources that support import

### wallarm-go changes for migration
- `APIError` struct replacing string-based error formatting — enables `errors.As()` for type-safe error checks
- `ClientFields.Enabled` changed from `bool` to `*bool` (fixes `omitempty` dropping `false`)
- `ClientDelete` method added for tenant deletion
- Internal logger removed — replaced by provider-side `loggingTransport`
- `HintDeleteFilter.ID` changed from `int` to `[]int` — supports batch delete (API limit: 1000 per call)
- Gzip compression — `Accept-Encoding: gzip` on all requests + transparent decompression
- HTTP headers copied (not replaced) — preserves Go's default transport headers

## Comment Field — API Validation

The `comment` field on rules (HintUpdateV3) rejects empty strings:
- API returns HTTP 400: `{"comment":{"error":"is too short (minimum is 1 character)"}}`
- Empty string `""` is not valid — must be at least 1 character or omitted entirely
- Provider should never send `comment: ""` to the API. Use `null` to omit or a non-empty string.

## Threshold/Reaction Rules — API Validation

Rules with `reaction` block (`brute`, `bola`, `enum`, `rate_limit_enum`, `forced_browsing`) have API-enforced constraints on which reaction keys are valid per mode:

- **mode `"block"`**: `reaction` must contain at least one of `block_by_session` or `block_by_ip`. Cannot use `graylist_by_ip`.
- **mode `"graylist"`** (if supported): allows `graylist_by_ip`.

The API returns HTTP 400: `"keys should contain at least one of [:block_by_session, :block_by_ip] keys for the mode block, keys should contain only [:block_by_session, :block_by_ip] keys for the mode block"` when invalid combinations are used.

Reaction values (`block_by_session`, `block_by_ip`, `graylist_by_ip`) must be in range **600..315569520** (10 minutes to ~10 years in seconds). The API returns HTTP 400: `"key values should be in 600..315569520"` for out-of-range values.

TODO: add provider-side validation for both constraints.

## GraphQL Detection Rule — API Limits

The `wallarm_rule_graphql_detection` resource has API-enforced value ranges (no provider-side validation yet):

| Field | Range | Note |
|-------|-------|------|
| `max_value_size_kb` | 1..100 | |
| `max_doc_size_kb` | (API-enforced) | |
| `max_alias_size_kb` | (API-enforced) | |
| `max_doc_per_batch` | (API-enforced) | |
| `max_depth` | (API-enforced) | |

The API returns HTTP 400 with `{"max_value_size_kb":{"error":"should be in 1..100"}}` for out-of-range values. TODO: add `ValidateFunc` to enforce these at plan time.

## Regex Syntax (Pire Engine)

Rules that use regex fields (`wallarm_rule_regex`, `wallarm_rule_credential_stuffing_regex`, action conditions with `type = "regex"`) are executed by the Wallarm node using the **Pire** regex engine ([yandex/pire](https://github.com/yandex/pire)). Pire is optimized for high-throughput matching (~400 MB/s) but has **limited syntax** compared to PCRE/RE2.

**Supported**: `|` (alternation), `*`, `+`, `{n}`, `{n,m}` (repetition), `.`, `[a-z]`, `\w`, `\W`, `\d`, `\D`, `\s`, `\S` (character classes), `^`, `$` (anchors), `&` (intersection), `~` (complement).

**NOT supported**: lookahead/lookbehind (`(?=...)`, `(?!...)`), backreferences (`\1`), capture groups, conditional patterns, non-greedy quantifiers (except in SlowScanner: `*?`, `+?`, `??`).

**Common pitfalls**:
- Standard email/URL regexes often use syntax unsupported by Pire. The API returns HTTP 400 with `{"regex":{"error":"invalid"}}` for unsupported syntax. Always test regex values against Pire syntax before using them.
- **Terraform HCL escaping**: Backslashes in regex must be doubled in HCL strings. `\w` → `\\w`, `\d` → `\\d`, `\s` → `\\s`. HCL does not recognize `\w` as a valid escape sequence and will error with `"The symbol \"w\" is not a valid escape sequence selector"`. Example: `regex = "\\w+@\\w+"` not `regex = "\w+@\w+"`.

## Known Issues / SDK Gotchas

### nil vs empty string in action TypeSet (terraform-plugin-sdk/v2)

The `action` field is a `TypeSet` of objects with `type`, `value`, and `point` attributes. Terraform's TypeSet uses a hash function (`hashResponseActionDetails`) to track elements. When `value` has `Computed: true`, Terraform treats `nil` and `""` differently — the API may return `null` for absent conditions, but the SDK stores `""` in state. This can cause spurious `(known after apply)` diffs on subsequent plans.

**Workaround**: The `actionValueString()` helper normalizes nil→"" when converting ActionDetails to schema maps. Any code producing action conditions for the TypeSet (CustomizeDiff, forward mapping) must ensure string values are never nil.

### Header name case sensitivity in action conditions

The Wallarm API stores header names in uppercase (`X-API-KEY`, `CONTENT-TYPE`). The provider's `ExpandSetToActionDetailsList()` and `hashResponseActionDetails()` uppercase header values. Any code producing action conditions (forward mapping in `ExpandPathToActions`, `actionDetailToSchemaMap` in CustomizeDiff) must also uppercase header names to avoid plan diffs.

### Detection point structure (`point` field)

**CRITICAL: This section is the ONLY authoritative source for point structure (paired vs simple elements, chaining rules). When constructing, reviewing, or testing point values — read this section first. Do NOT guess or approximate from memory.**

The `point` field is a list of lists of strings representing a path through the request parser chain. The `WrapPointElements()` function in `wallarm/common/resourcerule/resource_rule.go` is the authoritative reference for paired vs simple classification. Full point chaining data is in `.claude/point_map_exact.json` (fetched by `.claude/fetch_point_refs.py`). Proton type definitions (IDs, simple/keys/array/parser flags, attack type IDs) are in `.claude/types.rb`.

#### Base points (level 1)

Available in the Wallarm Console UI as top-level request parts:

| Base point(s) | Allowed children |
|---------------|-----------------|
| `action_ext`, `action_name`, `get_name`, `header_name`, `path`, `path_all`, `uri` | `base64`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `get`, `get_all` | `array`, `array_all`, `base64`, `gql`, `gzip`, `hash`, `hash_all`, `hash_name`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `header`, `header_all` | `array`, `array_all`, `base64`, `cookie`, `cookie_all`, `cookie_name`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `post` | `base64`, `form_urlencoded`, `form_urlencoded_all`, `form_urlencoded_name`, `gql`, `grpc`, `grpc_all`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `multipart`, `multipart_all`, `multipart_name`, `percent`, `xml` |

#### Paired elements (2-part: `["element", "value"]`)

| Element | Value type | Example |
|---------|-----------|---------|
| `header`, `cookie`, `get`, `hash`, `form_urlencoded`, `multipart`, `content_disp`, `response_header` | String (key/field name) | `["header", "HOST"]` |
| `jwt`, `json`, `json_obj`, `xml_tag`, `xml_attr`, `protobuf` | String (key/field name) | `["jwt", "payload"]` |
| `gql_query`, `gql_mutation`, `gql_subscription`, `gql_fragment`, `gql_dir`, `gql_spread`, `gql_type`, `gql_var` | String (operation/field name) | `["gql_query", "getUser"]` |
| `viewstate_dict`, `viewstate_sparse_array` | String (key name) | `["viewstate_dict", "key"]` |
| `path`, `array`, `json_array`, `grpc` | Integer (index) | `["path", 0]`, `["grpc", 1]` |
| `xml_pi`, `xml_dtd_entity`, `xml_tag_array`, `xml_comment` | Integer (index) | `["xml_pi", 0]` |
| `viewstate_array`, `viewstate_pair`, `viewstate_triplet` | Integer (index) | `["viewstate_array", 0]` |

#### Simple elements (1-part: `["element"]`)

`post`, `json_doc`, `xml`, `uri`, `action_name`, `action_ext`, `route`, `remote_addr`, `response_body`, `file`, `base64`, `gzip`, `htmljs`, `percent`, `pollution`, `gql`, `gql_alias`, `gql_arg`, `gql_inline`, `viewstate`, `viewstate_dict_key`, `viewstate_dict_value`, `protobuf_int32`, `protobuf_int64`, `protobuf_varint`, `xml_dtd`

#### Context-specific children (chains where unique elements appear)

Some elements only appear as children in specific contexts:

| Element | Context required | Example chain |
|---------|-----------------|---------------|
| `cookie`, `cookie_all`, `cookie_name` | Under `header` or `header_all` | `[["header", "COOKIE"], ["cookie", "session"]]` |
| `form_urlencoded`, `multipart`, `grpc`, `gql` | Under `post` | `[["post"], ["form_urlencoded", "field"]]` |
| `gql` in `json_doc` | Under `post > json_doc` | `[["post"], ["json_doc"], ["gql"]]` (not available under `action_ext > json_doc`) |
| `gql` in `percent` | Under `post > form_urlencoded > percent` or `get > percent` | `[["get", "q"], ["percent"], ["gql"]]` |
| `protobuf`, `protobuf_all`, `protobuf_name` | Under `grpc` (which is under `post`) | `[["post"], ["grpc", 1], ["protobuf", "field"]]` |
| `viewstate` and sub-elements | Under `base64` after a parser context | `[["post"], ["form_urlencoded", "f"], ["base64"], ["viewstate"]]` |
| `file`, `header` (nested) | Under `multipart` | `[["post"], ["multipart", "upload"], ["file"]]` |
| `content_disp` | Under `multipart > header` | `[["post"], ["multipart", "f"], ["header", "Content-Disposition"], ["content_disp", "filename"]]` |
| Post-context parsers in `gzip` | Under `post > gzip` | `post > gzip > json_doc` adds `form_urlencoded`, `grpc`, `multipart`, `gql` as children |

**Examples:**
```hcl
# Correct: form_urlencoded is 2-part, takes field name
point = [["post"], ["form_urlencoded", "username"]]

# Correct: hash is 2-part, takes key name
point = [["post"], ["json_doc"], ["hash", "password"]]

# Correct: get is 2-part, takes query param name
point = [["get", "search"]]

# WRONG: form_urlencoded without field name
point = [["post"], ["form_urlencoded"]]
```

## Action Resource & Hashing (`wallarm/common/resourcerule/`)

### Ruby-Compatible Hashing (`hash.go`)

`ConditionsHash(conditions []ActionDetails) string` — deterministic SHA256 matching Ruby's `Action.calculate_conditions_hash`. Used for action matching, directory naming, and deduplication. DB-verified against 4 real examples.

`PointHash(point []interface{}) string` — SHA256 matching Ruby's `HasPoint.calculate_point_hash`. Used per hit for grouping by detection point.

Both use `rawPack()` — a port of Ruby's `JSON.raw_pack` deterministic JSON serializer. The key detail: conditions are serialized as sorted `[["point","..."],["type","..."],["value",...]]` arrays with double-encoding of the point field.

### Action Directory Naming (`action_dir.go`)

`ActionDirName(conditions []ActionDetails) string` — 64-char max filesystem-safe directory names:
- Format: `{instance}_{domain}_{path}_{hash8}`
- Empty conditions → `_default` (no hash)
- Path-only → `_` prefix: `_api_v1_users_e3a1ef0f`
- Path wildcards: `*` → `.`, `**` → `..`, filename dot → `_` when wildcards present
- Truncation at `_` boundary if prefix exceeds 55 chars

Sort hierarchy: `_default` < `_path_only` < `13_instance` < `domain_based`

### Action Validation (`action_scope.go`)

`validateActionBlocks()` in `ActionScopeCustomizeDiff` validates explicit action blocks:
- Valid point keys only (header, method, path, action_name, action_ext, query, proto, scheme, uri, instance)
- Single key per point map
- `uri` conflicts with `path`/`action_name`/`action_ext`/`query` (mutually exclusive — no more URI fallback for deep paths)
- Instance requires `type = ""`
- Point-value points (action_name, method, etc.) require `value = ""`
- Header/query require non-empty value

### `wallarm_action` Resource (`resource_action.go`)

Read-only resource available for manual action tracking. Registered in the provider but **not used by the rules_engine module** — the module uses a pure local `action_map` instead (zero API calls on steady-state plans). The resource remains available for future use cases (e.g., manual imports, standalone action tracking).

### `data.wallarm_actions` Data Source (`data_source_actions.go`)

Discovery — fetches all non-empty actions with pagination. Returns list with `conditions_hash`, `dir_name`, `action_id`, `conditions`, `endpoint_*`, `updated_at`. Available for advanced configurations that need to organize rules by action scope.

### wallarm-go Action API Methods

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `ActionList(params)` | `POST /v1/objects/action` | List actions by filter (clientid, hint_type, empty) |
| `ActionReadByID(actionID)` | `GET /v3/action/{id}` | Single action by ID |
| `ActionReadByHitID(hitID)` | `POST /v1/objects/action/by_hit` | Action conditions for a hit (may not exist yet) |

Response struct `ActionEntry` has typed `Conditions []ActionDetails` + `EndpointPath/Domain/Instance *string`. `ActionDelete` was removed — API manages action lifecycle.

### Rule Delete Simplification

All 19 rule resources now use direct `HintDelete` only. The old pattern of `ActionList` → check hints count → `ActionDelete` was removed (the hints count fields were never in the API response, making the branch dead code).

### Hit Data Source Updates (`data_source_hits.go`)

- `action_hash` — now uses `ConditionsHash` (Ruby-compatible) instead of custom `computeActionHash`
- `point_hash` — new per-hit field using `PointHash` (Ruby-compatible). Replaces HCL-side `sha256(jsonencode(h.point_wrapped))`
- Action validation: after building conditions from hit data, calls `ActionReadByHitID` and compares hashes. Mismatch → error (expansion logic bug).

### Instance in Action Conditions (API behavior)

The Wallarm API has **two modes** for instance (pool ID) in action conditions — it can either include or exclude instance from rule action conditions. This is a per-client setting, not a per-request toggle.

**Impact on `data.wallarm_hits`:**
- `buildActionFromHit` includes instance when `include_instance=true` (default)
- `ActionReadByHitID` may or may not include instance depending on the client's API-side mode
- The hash validation compares provider-built conditions against `ActionReadByHitID` response — if the client's API mode excludes instance but `include_instance=true`, the hashes will differ
- When debugging a mismatch, the error now prints both condition sets inline — check whether instance is the only difference before investigating further

**`path = "[multiple]"` special case:**
- Some hits have `path: "[multiple]"` meaning the attack spans multiple URL paths
- In this case, `buildActionFromHit` skips all path/action_name/action_ext conditions, producing a HOST-header-only scope (`/**/*.*` wildcard)
- In attack mode, related hit filtering also skips path comparison when `refPath == "[multiple]"` — matches on domain + poolid only

## Rules Engine Module

This module lives in a separate repository (`terraform-wallarm-api`). Full documentation is in `.claude/rules_engine_module.md`.

## Future TODOs

### Performance
- **Integrations cache**: Currently each integration resource makes its own API read call, but the API returns the full list of all integrations in every response. With 11 integrations enabled, `terraform plan` makes 12 API calls (11 integration reads + 1 `/v1/user`) when 1 would suffice. Implement a simple shared cache at `ProviderMeta` level (similar to `HintCache`/`IPListCache`) that fetches once and serves all integration Read functions from cache.

### Code Quality
- **Integration resource factory extraction**: Many integration resources (email, Slack, Splunk, etc.) share identical CRUD patterns — extract into a shared factory. Significant refactor, separate PR.
- **Trigger complexity reduction**: Trigger resource has complex filter/action expansion that could be simplified.
- **Fully migrate `ResourceRuleWallarmCreate` callers**: 7 rule files still pass Read as a callback to the shared Create function. Should migrate to direct calls.

### Testing
- Fill testing TODOs: bola exact, brute exact, enum exact modes need test coverage.
- Some resources lack tests entirely — add basic CRUD + import tests.
- **Acceptance tests for `data.wallarm_actions` data source**.
- **Acceptance test for `ActionScopeCustomizeDiff`** — verify scope fields produce correct action blocks on real rule resources.
- **Acceptance test for URI validation** — verify `uri` conflicts with `path`/`action_name`/`action_ext`/`query`.

### Hints Cache / Import
- **Auto-refresh trigger**: detect when YAML file count changes → auto-refresh index. Currently manual (`import_rules=true`).
- **`is_managed` in index**: currently computed in parent outputs (avoids circular dependency). Explore storing in index for direct use.
- **`variativity_disabled` on import**: resolved — uses `Optional+Default:true` schema, same pattern as `comment`.

### Trigger
- **Trigger Read should populate all fields**: Currently only sets `trigger_id` and `client_id`. Should parse full API response (actions, thresholds, filters, name, comment, enabled) so `ImportStateVerify` works.

### IP Lists
- **Counts API validation**: The API has a `/access_rules/counts` endpoint that could validate entry counts without full fetch — useful for post-Create/Delete validation (see `memory/reference_ip_list_counts_api.md`).
- **Remove unused `get_vulns.go`** from wallarm-go (`POST /v1/objects/vuln`) — legacy endpoint not used by the provider.

### HCL Generator
- **`source = "convert"` mode**: Add a conversion mode to `wallarm_rule_generator` that reads existing `.tf` files with `action {}` blocks and rewrites them using `action_*` scope fields. Uses `ReverseMapActions()` from `resourcerule/action_reverse_map.go`. New `input_path` field for source file/directory, output to `output_dir`.

### Provider Framework
- **terraform-plugin-framework migration**: The current provider uses `terraform-plugin-sdk/v2`. Future major version could migrate to `terraform-plugin-framework` for better type safety, plan modifiers, and validator support. This is a large effort.
- **Integration tests for retry logic**: The wallarm-go retry logic (423/5xx/429) and hint cache invalidation should have dedicated unit tests.

## CI/CD

- **Unit tests**: run on push/PR across ubuntu + macos, Go 1.24
- **Acceptance tests**: run on push to master/develop, PRs, and scheduled (Friday 12:00 UTC)
- **Release**: triggered by `v*` tags, uses GoReleaser with GPG-signed checksums, builds for linux/darwin/windows/freebsd across multiple architectures

## HintsDB Action Model Reference

**CRITICAL: This section and "API Domain Model: Action & Condition" above are the ONLY authoritative sources for action/condition/point structures. Always verify against these sections.**

This section documents how Actions work on the backend (hintsdb service) to inform correct Terraform provider implementation.

### Action = Request Signature

An **Action** is a set of conditions that identifies a specific request scope (e.g. "POST requests to api.example.com/users"). Rules/hints are attached to actions. An **Endpoint** is an Action with `endpoint: true` — same table, different default scope.

### Database Schema (actions table)

Key fields:
- `id`, `clientid`, `name` (unique per client, format: `[A-Za-z0-9_.-]+`)
- `conditions_hash` — SHA256 of serialized conditions; unique constraint on `(clientid, conditions_hash)`
- `conditions_count` — denormalized count (0–60)
- `endpoint` (bool), `endpoint_url`, `endpoint_domain`, `endpoint_path`, `endpoint_instance`, `method`
- `actual` (bool), `internal` (bool), `hidden` (bool), `orphan` (bool)
- `endpoint_risk_score` — decimal(3,1), range 1–10
- `hits_count`, `request_stats` (JSONB)
- `discovered_at`, `changed_at`, `created_at`, `updated_at`

### Conditions (action_conditions table)

Each action has 0–60 conditions. A condition has:
- `type` — `equal`, `iequal`, `regex`, `absent`
- `point` — Proton::Point array identifying the request part (e.g. `[:header, 'HOST']`, `[:path, 0]`, `[:method]`, `[:instance]`, `[:action_name]`, `[:action_ext]`, `[:uri]`)
- `value` — binary match value (absent for `absent` type; lowercase for `iequal`)

**conditions_hash** is computed from all conditions sorted and SHA256'd. This is the primary lookup key — two actions with identical conditions on the same client are the same action.

### Condition Points → Terraform Scope Fields Mapping

| Proton Point | Terraform field | Notes |
|---|---|---|
| `[:header, 'HOST']` | `action.point.domain` / `action_domain` | Domain/host match |
| `[:path, N]` | `action.point.path` / `action_path` | Path segment at index N |
| `[:action_name]` | `action.point.action_name` / `action_action_name` | Last path segment (filename) |
| `[:action_ext]` | `action.point.action_ext` / `action_action_ext` | File extension |
| `[:method]` | `action.point.method` / `action_method` | HTTP method |
| `[:instance]` | `action.point.instance` / `action_instance` | Application/pool ID |
| `[:uri]` | `action.point.uri` / `action_uri` | Full URI (conflicts with path/action_name/action_ext) |
| `[:query, KEY]` | `action.point.query` / `action_query` | Query parameter |

### Action Lifecycle

1. **FindOrCreate** — looks up by `conditions_hash + clientid`; creates transactionally if not found; handles race conditions with retry
2. **Rules attached** — rules (hints) are created pointing to action via `actionid` FK
3. **LOM compilation** — triggered after rule changes; `Action.with_payload` loads actions, converts via `to_lom_action`, compiles into binary LOM
4. **Delete cascade** — when all rules removed from an action, the API auto-cleans empty actions (no conditions → action persists; with conditions + no rules → eligible for cleanup)

### Nested Actions

Actions whose conditions are a subset of another action's conditions. Used for hierarchical rule application — a rule on `/api/*` applies to `/api/users` too. The `with_nested` delete parameter removes parent actions.

### Rule (Hint) on an Action

The `hints` table stores rules. Key fields: `actionid` (FK), `type` (e.g. `wallarm_mode`, `disable_stamp`, `bruteforce_counter`), `system` (bool), `data` (msgpack blob containing full rule payload including `point`, `regex_id`, `clientid`).

The Terraform provider's `HintCreate`/`HintDelete` API calls map to creating/removing rows in this table, which triggers LOM recompilation.
