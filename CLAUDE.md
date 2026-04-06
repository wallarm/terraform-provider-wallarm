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

## API Domain Model: Action, Condition & Rule

**CRITICAL: This section is the ONLY authoritative source for action condition structure. When constructing, reviewing, or testing action conditions — read this section first. Do NOT guess or approximate from memory.**

**Detailed reference:** `.claude/action_reference.md` — path expansion algorithm, wildcards, headers, query params, special cases.
**Real examples:** `.claude/actions_examples.json` — 343 real action condition examples from the API for validation.

Rules in the Wallarm API are stored as **Action + Rule (Hint)** pairs.

### Action

An Action (`actions` table) represents a **scope** — a set of conditions that define where a rule applies. Multiple rules can share the same Action (same scope). An **Endpoint** is an Action with `endpoint: true` — same table, different default scope.

**Database fields:**
- `id` — unique Action ID (`action_id` in Terraform)
- `clientid` — tenant/client ID
- `name` — optional name (unique per client, validated: `[A-Za-z0-9_.-]+`)
- `conditions` — ordered list of Condition objects (the action conditions)
- `conditions_hash` — SHA256 of serialized conditions; unique constraint on `(clientid, conditions_hash)`
- `conditions_count` — denormalized count (0–60)
- `endpoint_path`, `endpoint_domain`, `endpoint_instance` — cached scope fields
- `endpoint` (bool), `endpoint_url`, `method`
- `actual` (bool), `internal` (bool), `hidden` (bool), `orphan` (bool)
- `endpoint_risk_score` — decimal(3,1), range 1–10
- `hits_count`, `request_stats` (JSONB)
- `discovered_at`, `changed_at`, `created_at`, `updated_at`

**Behavior:**
- `find_or_create` — if an Action with the same conditions already exists for the client, it's reused (not duplicated). This is why the provider checks `existsAction` before creating.
- `conditions_hash` enables fast equality matching — two Actions with identical conditions produce the same hash. This is the primary lookup key.
- `nested` — finds Actions whose conditions are a subset of this Action's conditions (for rule inheritance). A rule on `/api/*` applies to `/api/users` too. The `with_nested` delete parameter removes parent actions.
- After commit, triggers LOM compilation (`ScheduleLomCompilation`) and rule application (`ApplyAllForSingleAction`).

### Condition

A Condition (`action_conditions` table) represents a single matching rule within an Action scope. Maps to one `action {}` block in Terraform. Each action has 0–60 conditions.

**Key fields:**
- `type` — match type: `equal`, `iequal`, `regex`, `absent` (validated, default: `equal`)
- `point` — serialized as JSON, represents the request part to match (e.g., `["header", "HOST"]`, `["path", 0]`). Proton Point array: `[:header, 'HOST']`, `[:path, 0]`, `[:method]`, `[:instance]`, `[:action_name]`, `[:action_ext]`, `[:uri]`.
- `value` — the value to match against. For `iequal` type, automatically downcased before save. Absent for `absent` type.

**Behavior:**
- `iequal` values are **always lowercased** (`before_validation :iequal_values_downcase`). This is why the provider must lowercase `iequal` values (e.g., domain names via HOST header).
- `point` is stored as JSON and deserialized via `PointJsonDecoder` into `Proton::Point` objects.
- `to_h` output: `{ type: :equal, point: [...], value: "..." }` — this is the format the API returns and the provider processes.

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

### Relationship

```
Action (scope)  ←──has_many──→  Conditions (action conditions)
Action (scope)  ←──has_many──→  Rules/Hints (wallarm_mode, vpatch, etc.)
```

Multiple rules sharing the same scope (same conditions) point to the same Action. Creating a rule with the same conditions as an existing one reuses the Action. Deleting the last rule under an Action may delete the Action itself.

### Rule (Hint) on an Action

The `hints` table stores rules. Key fields: `actionid` (FK), `type` (e.g. `wallarm_mode`, `disable_stamp`, `bruteforce_counter`), `system` (bool), `data` (msgpack blob containing full rule payload including `point`, `regex_id`, `clientid`).

The Terraform provider's `HintCreate`/`HintDelete` API calls map to creating/removing rows in this table, which triggers LOM recompilation.

### Action Lifecycle

1. **FindOrCreate** — looks up by `conditions_hash + clientid`; creates transactionally if not found; handles race conditions with retry
2. **Rules attached** — rules (hints) are created pointing to action via `actionid` FK
3. **LOM compilation** — triggered after rule changes; `Action.with_payload` loads actions, converts via `to_lom_action`, compiles into binary LOM
4. **Delete cascade** — when all rules removed from an action, the API auto-cleans empty actions (no conditions → action persists; with conditions + no rules → eligible for cleanup)

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

### Action-Related Provider Code

**Hashing (`hash.go`):**
- `ConditionsHash(conditions []ActionDetails) string` — deterministic SHA256 matching Ruby's `Action.calculate_conditions_hash`. DB-verified against 4 real examples.
- `PointHash(point []interface{}) string` — SHA256 matching Ruby's `HasPoint.calculate_point_hash`.
- Both use `rawPack()` — a port of Ruby's `JSON.raw_pack` deterministic JSON serializer.

**Directory naming (`action_dir.go`):**
- `ActionDirName(conditions []ActionDetails) string` — 64-char max filesystem-safe names: `{instance}_{domain}_{path}_{hash8}`

**Validation (`action_scope.go`):**
- `validateActionBlocks()` in `ActionScopeCustomizeDiff` validates explicit action blocks
- Valid point keys, single key per map, `uri` conflicts with `path`/`action_name`/`action_ext`/`query`
- `PointValuePoints` (exported) — map of point-value types where value goes in point map

**wallarm-go Action API Methods:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `ActionList(params)` | `POST /v1/objects/action` | List actions by filter |
| `ActionReadByID(actionID)` | `GET /v3/action/{id}` | Single action by ID |
| `ActionReadByHitID(hitID)` | `POST /v1/objects/action/by_hit` | Action conditions for a hit |

**Resources:**
- `wallarm_action` — read-only resource for manual action tracking
- `data.wallarm_actions` — discovery of all non-empty actions with pagination
- All 19 rule resources use direct `HintDelete` only (no `ActionDelete`)

## Detection Point Structure (`point` field)

**CRITICAL: This section is the ONLY authoritative source for point structure (paired vs simple elements, chaining rules). When constructing, reviewing, or testing point values — read this section first. Do NOT guess or approximate from memory.**

**Full chaining data:** `.claude/point_map_exact.json` (fetched by `.claude/fetch_point_refs.py`).
**Type definitions:** `.claude/types.rb` — Proton type IDs, simple/keys/array/parser flags, attack type IDs.

The `point` field is a list of lists of strings representing a path through the request parser chain. The `WrapPointElements()` function in `wallarm/common/resourcerule/resource_rule.go` is the authoritative reference for paired vs simple classification.

### Base points (level 1)

| Base point(s) | Allowed children |
|---------------|-----------------|
| `action_ext`, `action_name`, `get_name`, `header_name`, `path`, `path_all`, `uri` | `base64`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `get`, `get_all` | `array`, `array_all`, `base64`, `gql`, `gzip`, `hash`, `hash_all`, `hash_name`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `header`, `header_all` | `array`, `array_all`, `base64`, `cookie`, `cookie_all`, `cookie_name`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `post` | `base64`, `form_urlencoded`, `form_urlencoded_all`, `form_urlencoded_name`, `gql`, `grpc`, `grpc_all`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `multipart`, `multipart_all`, `multipart_name`, `percent`, `xml` |

### Paired elements (2-part: `["element", "value"]`)

| Element | Value type | Example |
|---------|-----------|---------|
| `header`, `cookie`, `get`, `hash`, `form_urlencoded`, `multipart`, `content_disp`, `response_header` | String (key/field name) | `["header", "HOST"]` |
| `jwt`, `json`, `json_obj`, `xml_tag`, `xml_attr`, `protobuf` | String (key/field name) | `["jwt", "payload"]` |
| `gql_query`, `gql_mutation`, `gql_subscription`, `gql_fragment`, `gql_dir`, `gql_spread`, `gql_type`, `gql_var` | String (operation/field name) | `["gql_query", "getUser"]` |
| `viewstate_dict`, `viewstate_sparse_array` | String (key name) | `["viewstate_dict", "key"]` |
| `path`, `array`, `json_array`, `grpc` | Integer (index) | `["path", 0]`, `["grpc", 1]` |
| `xml_pi`, `xml_dtd_entity`, `xml_tag_array`, `xml_comment` | Integer (index) | `["xml_pi", 0]` |
| `viewstate_array`, `viewstate_pair`, `viewstate_triplet` | Integer (index) | `["viewstate_array", 0]` |

### Simple elements (1-part: `["element"]`)

`post`, `json_doc`, `xml`, `uri`, `action_name`, `action_ext`, `route`, `remote_addr`, `response_body`, `file`, `base64`, `gzip`, `htmljs`, `percent`, `pollution`, `gql`, `gql_alias`, `gql_arg`, `gql_inline`, `viewstate`, `viewstate_dict_key`, `viewstate_dict_value`, `protobuf_int32`, `protobuf_int64`, `protobuf_varint`, `xml_dtd`

### Context-specific children

| Element | Context required | Example chain |
|---------|-----------------|---------------|
| `cookie`, `cookie_all`, `cookie_name` | Under `header` or `header_all` | `[["header", "COOKIE"], ["cookie", "session"]]` |
| `form_urlencoded`, `multipart`, `grpc`, `gql` | Under `post` | `[["post"], ["form_urlencoded", "field"]]` |
| `gql` in `json_doc` | Under `post > json_doc` | `[["post"], ["json_doc"], ["gql"]]` |
| `gql` in `percent` | Under `post > form_urlencoded > percent` or `get > percent` | `[["get", "q"], ["percent"], ["gql"]]` |
| `protobuf`, `protobuf_all`, `protobuf_name` | Under `grpc` (which is under `post`) | `[["post"], ["grpc", 1], ["protobuf", "field"]]` |
| `viewstate` and sub-elements | Under `base64` after a parser context | `[["post"], ["form_urlencoded", "f"], ["base64"], ["viewstate"]]` |
| `file`, `header` (nested) | Under `multipart` | `[["post"], ["multipart", "upload"], ["file"]]` |
| `content_disp` | Under `multipart > header` | `[["post"], ["multipart", "f"], ["header", "Content-Disposition"], ["content_disp", "filename"]]` |
| Post-context parsers in `gzip` | Under `post > gzip` | `post > gzip > json_doc` adds `form_urlencoded`, `grpc`, `multipart`, `gql` as children |

**Examples:**
```hcl
point = [["post"], ["form_urlencoded", "username"]]  # Correct: 2-part
point = [["post"], ["json_doc"], ["hash", "password"]] # Correct: 2-part
point = [["get", "search"]]                            # Correct: 2-part
point = [["post"], ["form_urlencoded"]]                # WRONG: missing field name
```

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

**Mitigation Controls** — session-based, real-time threat mitigation:

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

Counter resources (`wallarm_rule_bruteforce_counter`, `wallarm_rule_dirbust_counter`, `wallarm_rule_bola_counter`) work with **triggers** (`wallarm_trigger`), not directly with mitigation controls.

**Rules** — request-level, applied per-request during traffic analysis. Both categories share the same underlying API model (Action + Hint).

### Shared Schema & Defaults (`wallarm/provider/default.go`)
- `defaultPointSchema` — 2D list-of-lists point structure
- `commonResourceRuleFields` — shared fields (rule_id, client_id, comment, active, title, rule_type, etc.)
- `mitigation` field: Optional+Computed, never sent to API
- `variativity_disabled` field: `Optional: true, Default: true`
- `comment` field: `Optional: true, Default: "Managed by Terraform"`. **`title` and `comment` are independent fields.**

### Common Utilities (`wallarm/common/`)
- `common.go` — string conversion helpers
- `const.go` — constants for point keys and match types
- `resourcerule/resource_rule.go` — shared CRUD logic, `ExpandSetToActionDetailsList()` for action expansion
- `mapper/tftoapi/` — Terraform schema to API format conversion
- `mapper/apitotf/` — API response to Terraform schema conversion

### API Limits & Constants (`wallarm/provider/constants.go`)

| Constant | Value | Purpose |
|----------|-------|---------|
| `IPListPageSize` | 1000 | IP list groups per API call |
| `IPListMaxSubnets` | 1000 | Max subnet values per IP list resource |
| `IPListCacheMaxRetries` | 3 | Cache refresh retries |
| `IPListCacheRetryDelay` | 3s | Wait between retries |
| `APIListLimit` | 500 | Default limit for rule/user/app list requests |
| `HintBulkFetchLimit` | 500 | Hints per page during cache lazy pagination |
| `HitFetchBatchSize` | 500 | Hits per API call in data source |

## Hits, Attacks & Automatic False Positive Rules

### Domain Model

A **hit** represents a single detected threat within an HTTP request. Hits sharing the same HTTP request are linked by `request_id`. Hits from the same attack campaign share an `attack_id`.

**IMPORTANT: Hits are ephemeral** — they have a retention period and can be dropped from the API at any time.

### False Positive Workflow

1. **Fetch**: `wallarm_hits` data source retrieves hits for given `request_id`(s)
2. **Group by Action**: Hits grouped by Host header + URI path (the Action scope)
3. **Group by Point**: Within each action, grouped by detection point
4. **Generate Rules**: Two rule types for FP suppression:
   - **`disable_stamp`** — allows specific attack signatures (stamps) at a given point
   - **`disable_attack_type`** — allows specific attack types at a given point
5. **One resource per rule**: Each stamp and each attack_type is a separate Terraform resource, matching the API 1:1. The `for_each` key is `{action_hash}_{point_hash}_{attack_type}_{stamp}` for stamp rules or `{action_hash}_{point_hash}_{attack_type}` for attack_type rules. Hash prefixes are 16 hex chars.

**Stampless attack types:** `xxe` and `invalid_xml` do not produce stamps. Hits of these types can only be suppressed via `disable_attack_type` rules.

### Data Source: `wallarm_hits`

**Input**: `request_id` (single string) + `mode` variable (`"request"` or `"attack"`). Called per-request_id via `for_each` in HCL.

**Hit filtering — allowed attack types:**
`xss`, `sqli`, `rce`, `ptrav`, `crlf`, `redir`, `nosqli`, `ldapi`, `scanner`, `mass_assignment`, `ssrf`, `ssi`, `mail_injection`, `ssti`, `xxe`, `invalid_xml`

**Key computed outputs:**
- `aggregated` — compact JSON with `action_hash` (16 chars), `action` conditions, and `groups` (each keyed by `point_hash_16 + "_" + attack_type`, containing stamps for that type and the attack_type)
- `action_hash` — Ruby-compatible `ConditionsHash`
- Action validation via `ActionReadByHitID` hash comparison

### Hits-to-Rules Flow

Three components: `wallarm_hits_index` (gating), `data.wallarm_hits` (fetching), `terraform_data.cache` (persistence). Deduplication by action_hash in HCL locals. See `docs/guides/hits_to_rules.md`.

## HCL Generator (`wallarm/provider/resource_hcl_generator.go`)

The `wallarm_rule_generator` resource generates HCL config files from cached rule data or existing API rules.

**Two source modes:**
- `source = "rules"` (default) — generates HCL from pre-built rules via `rules_json`
- `source = "api"` — fetches existing rules from the Wallarm API via `HintRead`

Each rule in `rules_json` carries its own `action` block — rules from different action scopes are correctly generated with their respective action conditions. The `expandedRule` struct has an `Actions` field for per-rule action conditions.

**Templates** (`hcl_generator_templates.go`): use `hclwrite` + `cty` for proper HCL generation with correct escaping.

## Hint Cache (`wallarm/provider/hint_cache.go`)

The `CachedClient` wraps `wallarm.API` and intercepts `HintRead` calls. Lazy-paginated: fetches one page at a time, stops when the requested ID is found. Thread-safe via `sync.Mutex`.

## Credential Stuffing Cache (`wallarm/provider/credential_stuffing_cache.go`)

Caches credential stuffing configs. Single API call returns all configs. Stored in `ProviderMeta.CredentialStuffingCache`.

## IP List Resources

IP list resources (`wallarm_allowlist`, `wallarm_denylist`, `wallarm_graylist`) are the most complex resources in the provider due to API behavior. One rule type per resource, max 1000 subnets. Cache at `ProviderMeta` level with per-rule-type fetching and Create serialization.

## wallarm-go Client Library

The `wallarm-go` library (`../wallarm-go`) is the HTTP client for the Wallarm API.

- `APIError` struct with `StatusCode` and `Body` fields
- Automatic retry: 423 (5s × 12), 5xx (10s × 12), 429 (exponential backoff × 12)
- Gzip compression on all requests (~19x reduction)
- All paginated methods set `response.Body.Objects = nil` before each `json.Unmarshal` (prevents slice reuse bugs)
- HTTP logging via `WALLARM_API_CLIENT_LOGGING=true`

## Other Resources

### Tenant (`wallarm_tenant`)
- Delete safety: disables first, only permanently deletes if `prevent_destroy=false` AND `WALLARM_ALLOW_CLIENT_DELETE=1`

### Node (`wallarm_node`)
- Default application (`app_id=-1`) is protected from deletion

### Trigger (`wallarm_trigger`)
- Import limitation: Read only populates `trigger_id` and `client_id`

## Wallarm User Roles & Multi-Tenancy

| Role | Scope | Key Permissions |
|------|-------|-----------------|
| **Administrator** | Single account | Full access: manage nodes, rules, integrations, users, API tokens, filtration mode |
| **Analyst** | Single account | View/manage attacks, incidents, vulnerabilities, API inventory |
| **Read Only** | Single account | View-only access to most entities |
| **API Developer** | Single account | View/download API inventory and specs only |
| **Deploy** | Single account | Create filtering nodes only |
| **Global Administrator** | Multi-tenant | Same as Administrator across all tenant accounts |
| **Global Administrator Extended** | Multi-tenant | Same as Global Administrator + can manage `disable_stamp` rules (FP suppression by signature) |
| **Global Analyst** | Multi-tenant | Same as Analyst across all tenant accounts |
| **Global Read Only** | Multi-tenant | Same as Read Only across all tenant accounts |

### Provider Implications
- Only **Administrator** and **Global Administrator** can manage users, integrations, and rules
- Only **Global Administrator Extended** can manage `disable_stamp` rules — standard Administrator/Global Administrator tokens will get 403 on `HintCreate` for this rule type
- Only **Global** roles can operate across tenant accounts
- `client_id` determines which tenant account to target
- `require_explicit_client_id` forces all resources to specify `client_id`

### Acceptance Test Env Vars

| Variable | Purpose |
|----------|---------|
| `WALLARM_API_TOKEN` | API token for authentication |
| `WALLARM_API_HOST` | API endpoint |
| `WALLARM_API_CLIENT_ID` | Target client/tenant ID |
| `WALLARM_EXTRA_PERMISSIONS` | Enable tests requiring elevated permissions |
| `WALLARM_GLOBAL_ADMIN` | Skip tests expecting 403 errors |
| `WALLARM_ALLOW_CLIENT_DELETE` | Allow permanent tenant deletion |
| `WALLARM_API_CLIENT_LOGGING` | Enable HTTP request/response logging |

## Testing Conventions

- Test files colocated with implementation in `wallarm/provider/`
- `generateRandomResourceName()` creates unique resource names
- All rule resource tests use `ImportStateVerifyIgnore: []string{"rule_type"}`
- Data sources requiring elevated permissions must use `retrieveClientID(d, m)` not `d.Get("client_id")`

## Known Issues / SDK Gotchas

### nil vs empty string in action TypeSet
`actionValueString()` normalizes nil→"" for TypeSet hash consistency. Any code producing action conditions must ensure string values are never nil.

### Header name case sensitivity
The API stores header names in uppercase. `ExpandSetToActionDetailsList()` and `hashResponseActionDetails()` uppercase header values.

## API Validation Notes

### Comment Field
API rejects empty strings on `comment`. Provider should never send `comment: ""`.

### Threshold/Reaction Rules
Reaction values must be in range **600..315569520**. Mode `"block"` requires `block_by_session` or `block_by_ip`.

### GraphQL Detection Rule
`max_value_size_kb` range: 1..100. Other fields API-enforced.

## Regex Syntax (Pire Engine)

Rules using regex fields are executed by the **Pire** engine. Limited syntax: no lookahead/lookbehind, no backreferences, no capture groups. Terraform HCL: double backslashes (`\\w`, `\\d`).

## Rules Engine Module

This module lives in a separate repository (`terraform-wallarm-api`). Full documentation in `.claude/rules_engine_module.md`.

## CI/CD

- **Unit tests**: push/PR across ubuntu + macos, Go 1.24
- **Acceptance tests**: push to master/develop, PRs, and scheduled (Friday 12:00 UTC)
- **Release**: `v*` tags, GoReleaser with GPG-signed checksums

## Future TODOs

### Performance
- **Integrations cache**: shared cache at `ProviderMeta` level for integration Read functions

### Code Quality
- **Integration resource factory extraction**: shared CRUD patterns
- **Trigger complexity reduction**
- **Fully migrate `ResourceRuleWallarmCreate` callers**: 7 rule files still use callback pattern

### Testing
- Fill testing TODOs: bola exact, brute exact, enum exact modes
- Acceptance tests for `data.wallarm_actions`, `ActionScopeCustomizeDiff`, URI validation

### Trigger
- **Trigger Read should populate all fields** for `ImportStateVerify`

### IP Lists
- **Counts API validation** via `/access_rules/counts` endpoint
- **Remove unused `get_vulns.go`** from wallarm-go

### HCL Generator
- **`source = "convert"` mode**: rewrite `action {}` blocks using `action_*` scope fields via `ReverseMapActions()`

### Provider Framework
- **terraform-plugin-framework migration** for better type safety
- **Integration tests for retry logic**
