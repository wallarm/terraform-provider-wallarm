# Rules

Covers all `wallarm_rule_*` resources, the action+condition+hint domain model, point structure, the rule registration seams, shared utilities, caches, and HCL generation.

## Resource catalog

Full list of supported `wallarm_rule_*` resources, grouped by purpose. `wallarm_rule_generator` is excluded — it is not an API rule, but an HCL emitter used by the import workflow.

**Filtration mode**

| Resource | Description |
|----------|-------------|
| `wallarm_rule_mode` | Per-scope filtration mode (`block` / `monitoring` / `off` / `default`). |
| `wallarm_rule_api_abuse_mode` | Toggles API Abuse Prevention per scope; primary use is allowlisting trusted crawlers. |

**Mitigation Controls** ([product overview](https://docs.wallarm.com/about-wallarm/mitigation-controls-overview/))

| Resource | Description |
|----------|-------------|
| `wallarm_rule_graphql_detection` | GraphQL API protection (depth/size/alias/batch/introspection limits). |
| `wallarm_rule_enum` | Enumeration attack protection (parameter-based). |
| `wallarm_rule_bola` | BOLA / IDOR protection. |
| `wallarm_rule_forced_browsing` | Forced-browsing protection. |
| `wallarm_rule_brute` | Brute-force protection. |
| `wallarm_rule_rate_limit_enum` | DoS / rate-limiting protection (enumerated). |
| `wallarm_rule_file_upload_size_limit` | File-upload size restriction policy. |

**Counters** (paired with `wallarm_trigger` via `hint_tag` filters)

| Resource | Description |
|----------|-------------|
| `wallarm_rule_bruteforce_counter` | Brute-force hit counter. |
| `wallarm_rule_dirbust_counter` | Directory-busting hit counter. |
| `wallarm_rule_bola_counter` | BOLA hit counter. |

**Pattern matching**

| Resource | Description |
|----------|-------------|
| `wallarm_rule_regex` | User-defined attack signature (Pire regex); also covers `experimental_regex`. |
| `wallarm_rule_ignore_regex` | Suppress matches of an existing user regex at a point. |

**Rate limiting**

| Resource | Description |
|----------|-------------|
| `wallarm_rule_rate_limit` | Per-scope request rate limit (rate / burst / delay / response status). |

**False-positive suppression** (typically generated from `data.wallarm_hits`)

| Resource | Description |
|----------|-------------|
| `wallarm_rule_disable_attack_type` | Allow a specific attack type at a point. |
| `wallarm_rule_disable_stamp` | Allow a specific attack signature (stamp) at a point. Requires Administrator (extended). |

**Credential stuffing**

| Resource | Description |
|----------|-------------|
| `wallarm_rule_credential_stuffing_regex` | Credential-stuffing detection by login + password regex pair. |
| `wallarm_rule_credential_stuffing_point` | Credential-stuffing detection by login + password points. |

**Data handling & response shaping**

| Resource | Description |
|----------|-------------|
| `wallarm_rule_masking` | Mask sensitive request data before storage / display. |
| `wallarm_rule_binary_data` | Treat a request part as binary (skip detect signatures). |
| `wallarm_rule_uploads` | Mark a request part as a file upload. |
| `wallarm_rule_parser_state` | Enable / disable a parser (`json_doc`, `xml`, `jwt`, `gql`, etc.) at a point. |
| `wallarm_rule_set_response_header` | Inject a response header on matched requests. |
| `wallarm_rule_vpatch` | Virtual patch — block specific attack types at a point. |
| `wallarm_rule_overlimit_res_settings` | Per-scope handling of requests exceeding processing limits. |

## Resource ↔ hint type ↔ shape mapping

Probe-derived API ground truth lives in `references/rules_api_fields.md`. The summary table below groups rules by **API shape** — what discriminates two hints on the same action scope. The shape determines whether `existingHintForAction` is wired (single-hint-per-scope shape only) and which mapper helpers apply.

| Shape | Discriminator | Resources | `existingHintForAction` guard |
|-------|--------------|-----------|------------------------------|
| **single hint per `(scope, type)`** | none — only the rule type | `wallarm_rule_mode`, `wallarm_rule_api_abuse_mode`, `wallarm_rule_overlimit_res_settings` | YES |
| **point-discriminated** | `point` field | `wallarm_rule_vpatch`, `_disable_stamp`, `_disable_attack_type`, `wallarm_rule_rate_limit`, `_file_upload_size_limit`, `_set_response_header`, `_masking`, `_binary_data`, `_uploads`, `_parser_state`, `_ignore_regex` | NO |
| **conditions-discriminated** | `enumerated_parameters` / `arbitrary_conditions` | `wallarm_rule_brute`, `_bola`, `_enum`, `_forced_browsing`, `_rate_limit_enum`, `_graphql_detection` | NO |
| **distinct rule identity** | `regex_id` / login+regex pair | `wallarm_rule_regex` (incl. experimental), `_credential_stuffing_regex`, `_credential_stuffing_point` | NO |
| **counter** | bound to `wallarm_trigger` | `wallarm_rule_bruteforce_counter`, `_dirbust_counter`, `_bola_counter` | NO |

## API domain model: Action, Condition & Rule

**CRITICAL: When constructing, reviewing, or testing action conditions — read this section first. Do NOT guess or approximate from memory.**

References:
- `references/action.md` — full path-expansion algorithm, wildcards, headers, query params, special cases, and the **server-side Action/Condition/Hint table model** (§14).
- `spec/actions_examples.json` — 82 representative action condition examples (one per distinct shape; deduped from a larger 343-sample probe).

Rules are stored as **Action + Hint** pairs. An **Action** is a scope (an ordered list of Conditions, max 60); multiple Hints can share the same Action. An **Endpoint** is just an Action with `endpoint: true`. Identical conditions reuse an existing Action via `find_or_create` (deterministic SHA256 `conditions_hash`); auto-cleanup removes Actions whose last Hint is deleted. The provider never calls `ActionDelete` — only `HintCreate`/`HintDelete`, both of which trigger server-side LOM recompilation.

### `existingHintForAction` is a defensive UX layer, not an API constraint

The Wallarm API itself **does not reject** duplicate rules:
- Same scope, **different** discriminator (e.g. mode) → API accepts both; one wins via async last-write-wins.
- Same scope, **same** discriminator → API accepts the second create, then asynchronously removes it.

Either way, HCL declares two rules but only one survives, and TF state diverges from the server on next refresh. `existingHintForAction` (`wallarm/provider/action_helpers.go`) is a **client-side guard**: before `HintCreate` it does an `ActionList` + `HintRead` to detect a same-scope-same-type rule, and on hit fails the apply with `ImportAsExistsError`. Catches both the duplicate and the contradiction (same scope, different mode) cases.

Intentionally mode-agnostic — triggers on `(client_id, hintType, conditions_hash)` without comparing rule-specific fields. **Don't add mode comparison** — that would let the contradiction case through.

Currently wired only for the three rule types in the **single-hint-per-`(scope,type)`** shape (see shape mapping above). Other shapes carry a discriminator (`point`, `enumerated_parameters`, `regex_id`, ...) that legitimately distinguishes multiple Hints on the same Action.

### Condition

A Condition maps to one `action {}` block in Terraform.

| Field | Values |
|-------|--------|
| `type` | `equal` / `iequal` / `regex` / `absent` (default: `equal`) |
| `point` | Proton Point array — `["header","HOST"]`, `["path",0]`, `["method"]`, `["instance"]`, `["action_name"]`, `["action_ext"]`, `["uri"]`, `["query","NAME"]` |
| `value` | Match value (absent for `type=absent`) |

**Server-side normalisation** the provider mirrors client-side to keep state stable:
- `iequal` values are always lowercased — the provider must lowercase domain names in HOST conditions, etc.
- Header **names** are always uppercased — `ExpandSetToActionDetailsList` and `TransformAPIActionToSchema` uppercase on the way in/out.

### Condition points → Terraform scope fields mapping

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

### Instance in action conditions (API has two modes)

The Wallarm API can either include or exclude instance (pool ID) from rule action conditions. This is a per-client setting, not a per-request toggle. Impact on `data.wallarm_hits`:

- `buildActionFromHit` includes instance when `include_instance=true` (default).
- `ActionReadByHitID` may or may not include instance depending on the client's API-side mode.
- Hash validation compares provider-built conditions against `ActionReadByHitID` — if API mode excludes instance but `include_instance=true`, hashes differ. The error prints both condition sets inline so the diff is visible.

### `path = "[multiple]"` special case

Some hits have `path: "[multiple]"` meaning the attack spans multiple URL paths. `buildActionFromHit` skips all `path` / `action_name` / `action_ext` conditions, producing a HOST-header-only scope (`/**/*.*` wildcard). In attack mode, related-hit filtering also skips path comparison when `refPath == "[multiple]"` — matches on domain + poolid only.

### Action-related provider code

**Hashing** (`hash.go`):
- `ConditionsHash([]ActionDetails) string` — deterministic SHA256 matching Ruby's `Action.calculate_conditions_hash`. DB-verified against 4 real examples.
- `PointHash([]interface{}) string` — matches Ruby's `HasPoint.calculate_point_hash`.
- Both use `rawPack()` — port of Ruby's `JSON.raw_pack` deterministic JSON serializer.

**Directory naming** (`action_dir.go`): `ActionDirName([]ActionDetails) string` — 64-char filesystem-safe `{instance}_{domain}_{path}_{hash8}`.

**Validation** (`action_scope.go`): `validateActionBlocks()` in `ActionScopeCustomizeDiff` enforces valid point keys, single key per map, and `uri` ⇄ `path`/`action_name`/`action_ext`/`query` conflicts. `PointValuePoints` (exported) lists point-value types where the value lives in the point map.

**wallarm-go Action API methods:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `ActionList(params)` | `POST /v1/objects/action` | List actions by filter |
| `ActionReadByID(actionID)` | `GET /v3/action/{id}` | Single action by ID |
| `ActionReadByHitID(hitID)` | `POST /v1/objects/action/by_hit` | Action conditions for a hit |

**Resources:** `wallarm_action` (read-only manual tracking), `data.wallarm_actions` (paginated discovery of all non-empty actions).

**Instance type preservation in state** (load-bearing — don't regress): instance conditions preserve `type="equal"` in state instead of forcing `type=""`, enabling future `type="regex"` instance matching. Implementation:
1. `HashActionDetails` — pure hash; normalizes instance type to `""` for hashing so config and state produce the same hash.
2. `TransformAPIActionToSchema` — pure transform (API → config); preserves instance type.
3. `ScopeActionSchema` uses `Set: HashActionDetails` so config and state share the same hash.
4. `type` field is `Computed: true` so omitted type inherits from state after first apply.
5. `ExpandSetToActionDetailsList` keeps `a.Type="equal"` for instance — API always gets explicit type.
6. Old `HashResponseActionDetails` preserved for back-compat but no longer used in Read paths.

## Detection point structure (`point` field)

**CRITICAL: When constructing, reviewing, or testing point values — read the reference first. Do NOT guess or approximate from memory.**

References:
- `references/point.md` — full paired/simple/context-children chaining tables.
- `spec/point_map.json` — raw chaining data (refresh every 30 days via `scripts/fetch_point_refs.py`).
- `references/proton-types.md` — Proton type IDs, simple/keys/array/parser flags, attack type IDs.
- Go authority: `WrapPointElements()` in `wallarm/common/resourcerule/action_expand.go`.

The `point` field is a list of lists representing a path through the request parser chain. Each inner list is either **simple** (`["element"]`, e.g. `["post"]`) or **paired** (`["element","value"]`, e.g. `["header","HOST"]`); some elements are only valid under a specific parent (e.g. `cookie` only under `header`). Get the rules from `references/point.md`.

```hcl
point = [["post"], ["form_urlencoded", "username"]]   # Correct: 2-part paired
point = [["post"], ["json_doc"], ["hash", "password"]] # Correct: simple → simple → paired
point = [["get", "search"]]                            # Correct: paired
point = [["post"], ["form_urlencoded"]]                # WRONG: form_urlencoded is paired, missing field name
```

## Mitigation Controls vs Rules (taxonomy)

The provider manages two categories of rule-like resources that differ in scope and behavior:

**Mitigation Controls** — session-based, real-time threat mitigation. Product overview: https://docs.wallarm.com/about-wallarm/mitigation-controls-overview/

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

**Mode ↔ reaction validation (API-side):** `wallarm_rule_brute`, `wallarm_rule_bola`, `wallarm_rule_forced_browsing`, `wallarm_rule_rate_limit_enum`, and other reaction-bearing mitigation controls enforce a per-mode reaction-key whitelist server-side:
- `mode = "block"` → reaction must contain only `block_by_session` and/or `block_by_ip`
- `mode = "monitoring"` → reaction must contain only `graylist_by_ip`

Mismatches return `HTTP 400 {"reaction":{"error":"keys should contain only [:graylist_by_ip] keys for the mode monitoring"}}` (or the analogous block-mode message). The schema does not enforce this — it's an API runtime check. When writing tests or examples, pair the mode with the matching reaction shape. `wallarm_rule_brute.mode` is also `ForceNew`, so changing mode in HCL forces destroy+recreate; in-place tests should pin a single mode and flip a different mutable field (e.g. `threshold.count`).

**Rules** — request-level, applied per-request during traffic analysis. Both categories share the same underlying API model (Action + Hint).

## Resource pattern (CRUD)

All rule resources merge their schema with `commonResourceRuleFields` via `lo.Assign()`, share `resourcerule.Read/Update/Import` from `wallarm/common/resourcerule/`, and use a 3-part Import ID (`{clientID}/{actionID}/{ruleID}`; the few 4-part exceptions are listed under registration below). For the full build flow (schema, CRUD wiring, registration, tests, docs), use the `create-rule-resource` skill — that is the canonical reference for adding or restructuring a `wallarm_rule_*` resource.

## Rule registration — two required seams

Adding a new `wallarm_rule_*` resource requires two separate registrations. Missing either silently breaks a user workflow.

1. **`provider.go` `ResourcesMap`** — makes the resource callable in HCL. Without this, `terraform apply` with the new resource type errors with "provider does not support resource type".
2. **`wallarm/common/resourcerule/action_reverse_map.go` `APITypeToTerraformResource`** — maps the API-side rule type string (`"wallarm_mode"`, `"vpatch"`, `"api_abuse_mode"`, ...) to the Terraform resource name. Consumed by the `data.wallarm_rules` data source (`data_source_rules.go:238, 262`), which drives the bulk rules-import workflow (see `docs/guides/rules_import.md`). **Unknown types are silently skipped** — a missing entry means existing rules of that type can't be discovered or imported via the generic pipeline.

Related registry: **`FourPartIDTypes`** (same file) — set of rule types whose Import ID is 4-part (`{client_id}/{action_id}/{rule_id}/{suffix}`). Currently `regex`, `experimental_regex` (suffix = `rule_type`), and `wallarm_mode` (suffix = mode value). Add an entry **only if** the resource's Import contract requires a 4th segment. Most resources use the 3-part default and should not appear here.

**Bulk import discovery seam:** when auditing any rule-related change, walk both `provider.go` **and** `action_reverse_map.go` — a resource registered only in the first is invisible to `data.wallarm_rules`.

## Shared schema & defaults (`wallarm/provider/schema_common.go`)

The decision rules for picking schema attributes (`Required`/`Optional`/`Computed`/`Default`/`ForceNew`/etc.) given an API field's characteristics are codified in `references/schema-decisions.md`. Walk the decision tree there when adding a new field; check the anti-patterns section when debugging "sticky on removal", import-trap, or perpetual-diff bugs.

- `defaultPointSchema` — 2D list-of-lists point structure
- `commonResourceRuleFields` — shared fields (rule_id, client_id, comment, active, title, rule_type, etc.)
- `mitigation` field: Optional+Computed, never sent to API
- `variativity_disabled` field: `Optional: true, Default: true`. The provider also hardcodes `VariativityDisabled: true` in shared `resourcerule.Create:193` regardless of the user's value. **Intentional:** denying server-side rule mutations keeps Terraform state stable and synchronized with the API. The Wallarm API can mutate variative rules out-of-band (signature evolution); locking variativity prevents that drift from corrupting Terraform state. Don't "fix" this to honor user input without a separate design discussion — it's load-bearing, not a bug.
- `comment` field: `Optional: true, Default: "Managed by Terraform"`. **`title` and `comment` are independent fields.**

## Common utilities (`wallarm/common/resourcerule/`)

- `const.go` — constants for point keys, match types, `ReadOption`/`CreateOption` types, `ConvertToStringSlice`
- `rule_crud.go` — shared CRUD logic (`Read`, `Create`, `Update`)
- `rule_import.go` — shared 3-part import helper (`Import`)
- `action_expand.go` — `ExpandSetToActionDetailsList()`, `WrapPointElements()`, `ExpandPointsToTwoDimensionalArray()`
- `action_hash.go` — `HashActionDetails()`, `TransformAPIActionToSchema()`, `ActionDetailsToMap()`
- `mapper_tftoapi.go` — Terraform schema to API format conversion (`*ToAPI` functions)
- `mapper_apitotf.go` — API response to Terraform schema conversion (`*ToTF` functions)

## Non-obvious Read patterns (audit trap)

Reads don't always use direct `d.Set(...)` calls. Before concluding a Read "drops fields", check for these patterns:

- **`resourcerule.Read` uses `setIfExists(d, key, val)`** — a schema-aware helper at `wallarm/common/resourcerule/rule_crud.go` that skips fields absent from the resource's schema. Accounts for ~32 rule-specific fields (`mode`, `attack_type`, `regex`, `size`, `burst`, `delay`, `rate`, `stamp`, `file_type`, `parser`, `state`, `values`, `debug_enabled`, `introspection`, `max_*_size_kb`, etc.). A grep for literal `d.Set("...")` in `rule_crud.go` will miss these entirely — use `grep -oE '(d\.Set|setIfExists\(d,)\s*\(?\s*"[a-z_]+"'` to catch both.
- **IP list Reads are config-driven, not API-driven.** See `references/ip-lists.md` for the rationale.
- **Config-only schema fields that are genuinely unreadable:** `wallarm_tenant.prevent_destroy` (provider-side safety flag, not API-persisted), `wallarm_user.password` (sensitive, API never echoes it back). Missing `d.Set` for these is intentional.

## Hint cache (`wallarm/provider/hint_cache.go`)

The `CachedClient` wraps `wallarm.API` and intercepts `HintRead` calls. Lazy-paginated: fetches one page at a time, stops when the requested ID is found. Thread-safe via `sync.Mutex`.

### Credential stuffing cache (`wallarm/provider/credential_stuffing_cache.go`)

Caches credential stuffing configs. Single API call returns all configs. Stored in `ProviderMeta.CredentialStuffingCache`.

### `HintDelete` response semantics

`POST /v1/objects/hint/delete` returns **HTTP 200 in all cases**; only the body distinguishes outcomes:

- `{"status":200,"body":[<hint>]}` — success; deleted hint payload returned.
- `{"status":200,"body":[]}` — no-op (ID doesn't exist OR server-side blocked delete, see counter case below).

wallarm-go treats 200 as success and returns `nil` regardless of body, which is fine for the provider's idempotent Delete path. Anything that wants to *verify* a delete actually happened must check `len(body) > 0`.

**Counter-specific:** `brute_counter`, `dirbust_counter`, and `bola_counter` hints cannot be deleted on demand — the server returns `body: []` for any direct delete request because counters are bound to `wallarm_trigger` resources via `hint_tag` filters. Counters auto-clean ~30s after their last trigger reference is removed. Acceptance tests simulating "console deletion" for drift detection must use a non-counter type.

## HCL generator (`wallarm/provider/hcl_generator.go`)

The `wallarm_rule_generator` resource generates HCL config files from cached rule data or existing API rules.

**Two source modes:**
- `source = "rules"` (default) — generates HCL from pre-built rules via `rules_json`
- `source = "api"` — fetches existing rules from the Wallarm API via `HintRead`

Each rule in `rules_json` carries its own `action` block — rules from different action scopes are correctly generated with their respective action conditions. The `expandedRule` struct has an `Actions` field for per-rule action conditions.

**Templates** (`hcl_generator_templates.go`): use `hclwrite` + `cty` for proper HCL generation with correct escaping.

## Regex syntax (Pire engine)

Rules using regex fields are executed by the **Pire** engine. See `references/regex.md` — covers supported/unsupported constructs, anchoring, HCL/JSON escaping, and practical examples for `wallarm_rule_regex`, `condition type = "regex"`, `enumerated_parameters.*_regexps`, and credential-stuffing rules.

## Known issues / SDK gotchas

### nil vs empty string in action TypeSet
`actionValueString()` normalizes nil→"" for TypeSet hash consistency. Any code producing action conditions must ensure string values are never nil.

### Header name case sensitivity
The API stores header names in uppercase. `ExpandSetToActionDetailsList()` and `TransformAPIActionToSchema()` uppercase header values.

### `enumerated_parameters` mode↔fields validation
`mapEnumeratedParameterExactToAPI` and `mapEnumeratedParameterRegexpToAPI` (`wallarm/common/resourcerule/mapper_tftoapi.go`) silently drop fields that don't apply to the chosen `mode`. Without plan-time validation this produces a perpetual diff loop. `EnumeratedParamsCustomizeDiff` (`wallarm/common/resourcerule/enumerated_params_diff.go`) rejects mismatches at plan time. Wired via `customdiff.All(ActionScopeCustomizeDiff, EnumeratedParamsCustomizeDiff)` on `wallarm_rule_brute`, `_bola`, `_enum`. Allowed combinations:
- `mode = "exact"` → only `points`
- `mode = "regexp"` → only `name_regexps`, `value_regexps`, `additional_parameters`, `plain_parameters`
- **Regexp mode requires both `name_regexps` AND `value_regexps` populated at plan time** (API constraint enforced by `EnumeratedParamsCustomizeDiff`). Use `[""]` to opt out of one filter while satisfying the API.

### `arbitrary_conditions.point` flat→2D round-trip
The API stores `arbitrary_conditions[].point` as a flat array (e.g. `["post","json_doc","hash","user_id"]`); HCL exposes it as 2D paired/simple chunks (e.g. `[["post"],["json_doc"],["hash","user_id"]]`). `ArbitraryConditionsToTF` (`wallarm/common/resourcerule/mapper_apitotf.go`) re-chunks the flat list via `WrapPointElements`. Anyone adding a new rule type with a `point` field on a non-action-scope nested block must do the same — otherwise `terraform plan` will force-replace every run because the state shape mismatches the user's HCL shape.

### `GetPointerIfConfigured[T]` — Optional+Computed primitives with API defaults
When a schema is `Optional+Computed` and the API has a non-zero default, Create must NOT send a literal zero for unset fields. The generic helper in `wallarm/common/resourcerule/rule_crud.go` (`GetPointerIfConfigured[T any]`) reads `d.GetRawConfig()` to detect whether the user actually wrote the field; returns `nil` if not, so `*T+omitempty` drops it on the wire and the API default wins. Works for any primitive (`int`, `bool`, `string`). For bool fields with stable API defaults, prefer flipping schema to `Optional+Default(<value>)` directly — symmetric remove-restores-default UX without needing the pointer round-trip. Use the helper for fields that stay Optional+Computed.

## API validation notes

### Comment field
API rejects empty strings on `comment`. Provider should never send `comment: ""`.

### Threshold/Reaction rules
Reaction values must be in range **600..315569520**. Mode `"block"` requires `block_by_session` or `block_by_ip`.

### GraphQL detection rule
`max_value_size_kb` range: 1..100. Other fields API-enforced.

### `attack_type` allowlists (per resource)

Validators on `wallarm_rule_disable_attack_type.attack_type`, `wallarm_rule_vpatch.attack_type`, and `wallarm_rule_regex.attack_type` are curated subsets of the upstream Proton attack-types enum. Per-resource lists differ — `disable_attack_type`/`vpatch` accept the special wildcard `any` and `invalid_xml`; `regex` accepts `vpatch` instead. Refresh sources when the API evolves:

- **`GET /v2/attack_types`** — runtime API endpoint that returns the canonical attack-type set the API currently accepts. Authoritative for "which values does the live API recognise today?".
- **`references/proton-types.md`** — committed reference extracted from the upstream Proton types snapshot (`gl.wallarm.com/wallarm-node/meganode/.../proton/types.rb`). Useful for offline lookup; refresh by re-syncing from upstream. Provider validators don't include all values from this file (many are internal: `warn`, `marker`, `bot`, the API-Policy-Enforcement group, the GraphQL-parser group).

When updating the validator lists, cross-check both sources, then audit per-resource semantics (e.g. `any` makes no sense for a regex rule).
