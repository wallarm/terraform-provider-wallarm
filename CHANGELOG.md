## [Unreleased] - v2.3.7

### Bug Fixes

* Fixed `data.wallarm_rules` double-counting `wallarm_rule_credential_stuffing_regex` and `wallarm_rule_credential_stuffing_point` rules when Create and the data source read happen in the same `terraform apply`. The v2.3.2 dedup fix added the `isCredentialStuffingType` filter at three cache-population sites but missed `HintCache.Insert`; this completes the fix at the fourth site so credential_stuffing rules never enter the generic hint cache.
* `terraform destroy` of `wallarm_rule_bruteforce_counter`, `wallarm_rule_dirbust_counter`, and `wallarm_rule_bola_counter` is now state-only. The Wallarm API rejects on-demand counter deletes (returns HTTP 200 with empty body) and counters auto-clean ~30 seconds after their last trigger reference is removed; the previous Delete implementation issued a no-op API call and falsely reported destroy success while the counter persisted server-side. The new behavior drops the resource from state and emits an `[INFO]` log line directing operators to the auto-clean lifecycle.
* `terraform destroy` on every other rule resource now emits a `[WARN]` log line when the API returns an empty response body — the rule was already absent server-side (deleted out-of-band, never existed, or silently rejected). The destroy still succeeds, but operators get a signal that their delete may not have changed anything. Previously this case was indistinguishable from a successful delete because wallarm-go discarded the response body.

* Fixed perpetual destroy+recreate plan diff for `iequal`-typed action conditions with mixed-case values. The Wallarm API downcases `iequal` values server-side, so any mixed-case literal in HCL would drift against the lowercased state on every plan. The fix covers both shapes the matched string can take:
  - **`value` field** for paired-element points (`header`, `query`): e.g. `point = { header = "HOST" }, value = "Example.com"`.
  - **point map value** for value-bearing points (`action_name`, `action_ext`, `method`, `instance`, `proto`, `scheme`, `uri`): e.g. `point = { action_name = "TEST" }`.
  The action TypeSet hash now lowercases iequal values in both shapes (set membership stable), and `DiffSuppressFunc`s on `action.value` and `action.point` treat case-only differences as equivalent when the sibling `type` is `iequal`. Affects every rule resource using `ScopeActionSchema` / `ScopeActionSchemaMutable`.

* `terraform plan` now reports **update-in-place** (instead of destroy+recreate) for many rule schema fields whose API supports PUT updates. Previously every non-`comment`/`variativity_disabled` change required replacement because the schema declared the field `ForceNew` or the provider's shared Update path never sent it. The audit at `.claude/manual-testing/audit-update-fields/results.md` enumerates all 45 field/resource pairs migrated from `ForceNew` to mutable, plus 8 fields explicitly kept `ForceNew` (where PUT is rejected or silently no-op). Affected resources: `wallarm_rule_mode`, `wallarm_rule_api_abuse_mode`, `wallarm_rule_vpatch`, `wallarm_rule_disable_stamp`, `wallarm_rule_disable_attack_type`, `wallarm_rule_regex`, `wallarm_rule_parser_state`, `wallarm_rule_set_response_header`, `wallarm_rule_uploads`, `wallarm_rule_overlimit_res_settings`, `wallarm_rule_graphql_detection`, `wallarm_rule_file_upload_size_limit`, `wallarm_rule_rate_limit`, `wallarm_rule_rate_limit_enum`, `wallarm_rule_bola`, `wallarm_rule_brute`, `wallarm_rule_enum`, `wallarm_rule_forced_browsing`, `wallarm_rule_credential_stuffing_regex`, `wallarm_rule_credential_stuffing_point`. Existing state files keep working — the schema change is purely additive (drops `ForceNew`).

### Other Changes

* Bumped `wallarm-go` to `v0.12.0` for the new `HintDelete` response surface (`*HintDeleteResp{Status, Body []ActionBody}`) and the extended `HintUpdateV3Params` (~30 new pointer fields covering the full mutable rule surface). End users writing only HCL are unaffected; downstream Go consumers calling `wallarm.API.HintDelete(...)` see the breaking signature change documented in the wallarm-go changelog.
* Extracted shared `resourcerule.Delete` factory used by 21 rule resources (replaces ~17 lines of identical Delete boilerplate per resource). The 2 `wallarm_rule_credential_stuffing_*` resources keep custom Deletes because they additionally invalidate `CredentialStuffingCache` after the API call.
* Extended `resourcerule.Update` with variadic `UpdateCustomizer` parameters and a library of `With*` helpers (one per mutable field). Per-resource Update wiring is now a one-liner like `UpdateContext: resourcerule.Update(apiClient, resourcerule.WithMode, resourcerule.WithThreshold, resourcerule.WithReaction)`.

## [v2.3.6] - 2026-04-24

> Full Terraform support for Wallarm's API Specification Enforcement: expanded `wallarm_api_spec` schema, new `wallarm_api_spec_policy` resource for per-violation block/monitor/ignore controls, and wallarm-go v0.11.0 with the backing endpoints.

### Upgrade Steps

* Bump the provider to `v2.3.6`. No HCL migration is required — existing `wallarm_api_spec` configurations continue to apply unchanged.
* `wallarm_api_spec.domains` and `wallarm_api_spec.instances` changed from Required to Optional. If you relied on Terraform failing fast when these were missing, add `validation` blocks on your input variables instead.
* [ACTION REQUIRED] Provider developers consuming `wallarm-go` directly must update imports: the `Api*` prefix is gone. Rename `ApiSpec` → `APISpec`, `ApiSpecBody` → `APISpecBody`, `ApiDetection` → `APIDetection`, etc. JSON tags are unchanged, so API payloads are wire-compatible. End users writing only HCL are unaffected.

### New Features

* New resource `wallarm_api_spec_policy` — manages the [API Specification Enforcement][api-spec-enf] policy attached to an uploaded spec. Six violation modes (`undefined_endpoint_mode`, `undefined_parameter_mode`, `missing_parameter_mode`, `invalid_parameter_value_mode`, `missing_auth_mode`, `invalid_request_mode`), two threshold modes (`timeout_mode`, `max_request_size_mode`), and a repeatable `condition` block that reuses the rule action-scope schema. Destroy is a soft-delete (PUT `enabled: false`) — the policy record is removed only when the parent spec is deleted.
* `wallarm_api_spec`: extended schema with `auth_headers` (list of `{key, value}` blocks, sensitive values) used when Wallarm fetches the spec URL, plus API-computed attributes — `status`, `spec_version`, `openapi_version`, `endpoints_count`, `shadow_endpoints_count`, `orphan_endpoints_count`, `zombie_endpoints_count`, `format`, `version`, `node_sync_version`, `last_synced_at`, `last_compared_at`, `updated_at`, `created_at`, `file_changed_at`, and a nested `file` block (`name`, `signed_url`, `checksum`, `mime_type`, `version`). `file.signed_url` is Sensitive and regenerates on every Read (short-lived, ~10 min).
* `wallarm_api_spec`: Update is now wired via `APISpecUpdate` (partial PUT), so fields can be mutated in place instead of forcing destroy/recreate.

### Other Changes

* **build(deps):** bump `wallarm-go` to `v0.11.0` — adds `APISpecReadByID`, `APISpecUpdate`, `APISpecPolicyPut` endpoints; extends `APISpecBody` with `AuthHeaders` and the new computed fields; renames `Api*` Go identifiers to `API*` for idiomatic Go initialisms (JSON tags unchanged).
* **refactor(api_spec):** `domains` and `instances` relaxed from Required to Optional. The Wallarm console now hides these legacy fields and Wallarm computes the applicable scope from other signals, so forcing a value in HCL was misleading.
* **feat(api_spec):** `title`, `description`, `file_remote_url`, `regular_file_update`, `api_detection`, `domains`, `instances` are now mutable (no more `ForceNew`). Edits apply as an in-place PUT via `APISpecUpdate` instead of destroy/recreate. `client_id` remains `ForceNew`.
* **test(api_spec):** acceptance-test suite migrated to v2.3.5 patterns (`ProtoV5ProviderFactories`, `testAccNewAPIClient`, unique per-test `title`). Added Update-in-place, Import-round-trip, and `auth_headers` lifecycle coverage. New `resource_api_spec_policy_test.go` covers Create, Update-mode-flip, scoped `condition` blocks, soft-delete, and import.
* **docs(api_spec):** resource doc rewritten against the v2.3.6 schema — full attribute list including the nested `file` block with `signed_url` caveat; example updated.
* **docs(api_spec_policy):** new resource doc with Example Usage, Argument Reference split into Scope / Violation Modes / Threshold Limits, Import (3-part ID `{client_id}/{api_spec_id}/policy`), and a Limitations section documenting soft-delete behaviour and the one-policy-per-spec constraint.
* **docs(examples):** new `examples/wallarm_api_spec_policy.tf` with three scenarios (observation-mode rollout, block on undefined endpoints with host scope, paused policy preserving settings).
* **docs(readme):** added `wallarm_api_spec_policy` to the Infrastructure & Tooling table; resource count bumped 11 → 12.
* **chore(schema):** added `Description` strings to 15+ previously-undocumented `wallarm_api_spec` fields so `terraform providers schema` and registry docs surface helpful text.

[api-spec-enf]: https://docs.wallarm.com/api-specification-enforcement/overview/

# v2.3.5 (Apr 22, 2026)

## FEATURES:

* New resource `wallarm_rule_api_abuse_mode` — toggles [API Abuse Prevention](https://docs.wallarm.com/api-abuse-prevention/overview/) for requests matching an action scope. Primary use case: allowlist trusted crawlers (Pinterest, Google, monitoring agents) with `mode = "disabled"`. `mode` is `ForceNew` — changing it destroys and recreates the rule.

## IMPROVEMENTS:

* Extracted shared `existingHintForAction` helper for duplicate-rule detection on Create. Used by `wallarm_rule_mode` and `wallarm_rule_api_abuse_mode`; replaces the per-resource `existsAction` + `existsHint` pair.
* Replaced hand-rolled action-conditions comparison (`equalWithoutOrder`, `compareActionDetails`, `actionPointsEqual`, `convertToStringSlice`) with a single `resourcerule.ConditionsHash` compare. Matches the Wallarm API's own canonical action identity and removes ~85 lines of per-field equality logic.
* Added `testAccNewAPIClient()` test helper for CheckDestroy in tests that use `ProtoV5ProviderFactories`. Constructs an API client from `WALLARM_API_TOKEN` / `WALLARM_API_HOST` without going through the shared `testAccProvider`, avoiding a `Configure` race under `-race`.

## BUG FIXES:

* `CachedClient.HintDelete` now invalidates the hint cache on success. Previously the cache could return stale entries for just-deleted rules, surfacing in acceptance tests as "dangling resource" errors in CheckDestroy after a Create path that populated the cache.

## DOCUMENTATION:

* New `docs/resources/rule_api_abuse_mode.md` with Pinterest allowlist example, Argument/Attributes reference, and 3-part (`{client_id}/{action_id}/{rule_id}`) Import section.
* New `examples/wallarm_rule_api_abuse_mode.tf` with three scope patterns (enable per instance, enable per host, Pinterest allowlist).
* README Rules table: added `wallarm_rule_api_abuse_mode` entry; Rules count bumped 20 → 21.

# v2.3.4 (Apr 20, 2026)

## IMPROVEMENTS:

* Bumped `wallarm-go` dependency to v0.10.0 — adds Attack, Activity Log, and Security Issues API methods; nil-input guards on request helpers; IP-list search query encoding fix; cursor pagination for `AttackRead`; hit block-status filter.
* Extracted shared `ResourceRuleWallarmImport` / `ResourceRuleWallarmUpdate` helpers in `wallarm/common/resourcerule/` — 42 rule resources migrated, ~850 net LOC of duplicated boilerplate removed.
* Shortened `resourcerule` public API: `ResourceRuleWallarm{Read,Create,Update,Import}` → `{Read,Create,Update,Import}`. Callers use `resourcerule.Read(...)` etc.
* Added unit tests for `validateActionSet` (6 cases), `EnumeratedParametersToTF/ToAPI`, `ArbitraryConditionsToTF/ToAPI`, `mapEnumeratedParameter{Regexp,Exact}ToAPI`, plus 5 cases for the new `importIntegration` helper. `resourcerule` coverage 55.7% → 69.3%.
* Extracted `modeExact` / `modeRegexp` constants in `resourcerule/const.go`, replacing repeated string literals.

## BUG FIXES:

* `terraform import` now works correctly for `wallarm_api_spec` and `wallarm_user`. Previously these used `schema.ImportStatePassthroughContext` which did not populate the fields `Read` requires, so import commands failed with zero-ID API calls. Real `StateContextFunc` parsers have been added.
* `data.wallarm_security_issues` no longer panics when an issue has no vpatch mitigation. `SecurityIssueMitigations.Vpatch` (now a pointer in `wallarm-go` v0.10.0) is nil-checked before dereference.
* `wallarm_api_spec` import ID format changed to `{client_id}/{api_spec_id}` (was single integer `{api_spec_id}`) to align with the `{client_id}/{resource_id}` convention used by other resources. Existing state with bare-integer IDs remains functional — `Read` does not parse `d.Id()`, so no destroy/recreate occurs on upgrade.

## BREAKING CHANGES:

* Removed the `ignore_existing` provider argument and the `WALLARM_IGNORE_EXISTING_RESOURCES` environment variable. The field was defined in the schema but never read by any code, so existing configurations that set it had no effect. Remove the assignment from your `provider "wallarm" {}` block after upgrading.

## DOCUMENTATION:

* Added `## Import` sections to 4 resource docs (`api_spec`, `node`, `tenant`, `user`) with concrete ID-format examples.
* Renamed `docs/resources/integration_ms_teams.md` → `integration_teams.md` to match the resource name `wallarm_integration_teams`.

# v2.3.3 (Apr 19, 2026)

## IMPROVEMENTS:

* Bumped `wallarm-go` dependency to v0.9.1 — adds unit test coverage (79.5%), removes unused `vuln_prefix` and `get_vulns` endpoints
* Restructured `wallarm/common/` — collapsed `mapper/` packages into `resourcerule/`, deleted `common` package, all constants and helpers consolidated
* Split monolithic `resource_rule.go` (647 lines) into focused files: `rule_crud.go`, `action_hash.go`, `action_expand.go`
* Dissolved `utils.go` — moved functions to domain files (`integration_helpers.go`, `action_helpers.go`, `resource_user.go`, etc.)
* Renamed files for clarity: `default.go` → `schema_common.go`, `utils.go` → `provider_helpers.go`, `resource_hcl_generator.go` → `hcl_generator.go`
* Added `make dev` / `make dev-clean` targets for local provider development without version hardcoding
* CI: dedicated test tenant per acceptance test run — eliminates orphaned resources from failed pipelines
* Removed `vuln_prefix` from tenant creation (field removed from Wallarm API)
* Unit test coverage: `resourcerule` raised to 55.7%, added ~50 unit tests across `resourcerule` and `provider` packages

## BUG FIXES:

* Removed `generateVulnPrefix` — sending `vuln_prefix` to the API now causes errors
* Fixed counter resources (`bola_counter`, `bruteforce_counter`, `dirbust_counter`) failing with 403 on import — removed Update (counters are immutable), made `comment` and `variativity_disabled` computed-only

## DOCUMENTATION:

* `docs/index.md` — documented `hint_prefetch` and `require_explicit_client_id` provider arguments; corrected `max_backoff` default from 30 to 5
* `README.md` — Terraform required version raised to >= 1.5

# v2.3.2 (Apr 10, 2026)

## IMPROVEMENTS:

* Instance action conditions now preserve `type="equal"` in state instead of clearing to empty string, enabling future `type="regex"` support
* Added `admin_ext` role to `wallarm_user` resource for Administrator (extended) support

## BUG FIXES:

* Fixed credential stuffing rules appearing twice in `data.wallarm_rules` when API token has elevated permissions
* Fixed instance condition `type` field lost after resource Read, causing perpetual drift when explicitly set
* Fixed wrong field name `resource_type` → `terraform_resource` in 5 resource docs and `rules_import` guide
* Fixed incorrect `rule_type` examples in `rule_bola`, `rule_brute`, `rule_forced_browsing`, `rule_graphql_detection`, `rule_rate_limit_enum` docs

## DOCUMENTATION:

* Rewrote `rules_import` guide with complete configuration, all workflows (native import, generator fallback), sync status, and filtering
* Added `terraform {}` block with `required_version >= 1.5` to provider example
* Added Administrator (extended) role documentation to provider and user resource docs
* Fixed `action` guide: `type` field documented as optional (was incorrectly marked required), added missing common fields (`title`, `active`, `set`)
* Fixed `mitigation_controls` guide: corrected file upload parameters (`size` not `max_size`, removed non-existent `file_types`), clarified `safe_blocking` is `rule_mode` only

# v2.3.1 (Apr 7, 2026)

## IMPROVEMENTS:

* Stamps grouped per attack type for traceability — each group now has `attack_type` always set
* New `disable_attack_type` bool field in aggregated groups — controls whether `disable_attack_type` rule is created, decoupled from `attack_type` key
* `rule_types` filter correctly controls both stamp and attack_type rule creation

## BUG FIXES:

* Fixed `rule_types = ["disable_stamp"]` still creating `disable_attack_type` rules
* Fixed nil stamps for stampless types (`xxe`, `invalid_xml`) causing HCL null errors
* Fixed stamp group key truncation dropping attack_type suffix

# v2.3.0 (Apr 6, 2026)

## FEATURES:

* **`data.wallarm_hits`: `rule_types` filter** — controls which rule types to generate (`disable_stamp`, `disable_attack_type`, or both). Validated at plan time.
* **`data.wallarm_hits`: `include_instance` parameter** — controls whether instance (pool ID) is included in action conditions
* **`data.wallarm_hits`: `aggregated` output** — compact JSON representation for efficient caching in `terraform_data`
* **`wallarm_rule_generator`: `source = "rules"` mode** — generates HCL from pre-built rules via `rules_json` input, with per-rule action conditions (different action scopes generate correct HCL)
* **`wallarm_hits_index`: `ready` flag** — enables single-apply workflow (no double-apply on first run)
* **`wallarm_hits_index`: `cached_request_ids` as TypeSet** — replaces comma-separated string
* **Hits-to-rules deduplication** — multiple request_ids with identical actions are deduplicated in HCL locals
* **16-char hash prefixes** — action_hash and point_hash in `for_each` keys use 16 hex chars (was 8) for collision safety

## BREAKING CHANGES:

* **`data.wallarm_hits`: `attack_types` now filters in all modes** — previously only filtered API fetch in attack mode. Now also filters which hits produce rules in request mode.
* **`data.wallarm_hits`: `rules` output removed** — use `aggregated` output with `terraform_data` caching instead. Also removed: `rules_count`, `rules_stamp_count`, `rules_attack_type_count`. See the [Hits to Rules Guide](docs/guides/hits_to_rules.md).
* **`wallarm_hits_index`: `cached_request_ids` type changed** — from comma-separated `string` to `TypeSet`. HCL using `split(",", ...)` must be updated to use the set directly.
* **`wallarm_rule_generator`: `source = "hits"` removed** — use `source = "rules"` (now default) with `rules_json` input. The `requests_json` field is removed.

## IMPROVEMENTS:

* Stamps grouped per attack type in aggregated output for traceability
* Error propagation in `buildAggregatedJSON` (returns error instead of logging)
* `PointValuePoints` exported from `resourcerule` package for shared use
* `containsInt`/`containsStr` moved to `utils.go`
* Removed dead code: `generateFromHits`, `groupHitsByPoint`, `parseActionConditionsJSON`, `hitJSON`, `requestEntry`
* New unit tests for `common.ConvertToStringSlice`, `apitotf`, `tftoapi` mappers
* New integration tests for `generateFromRulesJSON` with multiple action scopes

## DOCUMENTATION:

* New: `docs/data-sources/actions.md`
* Updated: `docs/resources/rule_parser_state.md` — added `jwt` and `gql` parsers
* Updated: `docs/guides/hits_to_rules.md` — single-apply flow, resource naming, deduplication, stampless types
* Updated: `docs/resources/hits_index.md` — `ready` flag, TypeSet `cached_request_ids`
* Updated: `docs/resources/rule_generator.md` — `source = "rules"` default, removed `source = "hits"`

# v2.2.1 (Apr 2, 2026)

## DOCUMENTATION:

* Reorganized registry sidebar: resources and data sources grouped into Common, Integrations, IP Lists, and Rules subcategories
* Fixed guide page titles for registry display (`action` → "Wallarm Rule Action", `point` → "Wallarm Rule Point")
* Removed redundant "Guide" subcategory from guides

# v2.2.0 (Apr 1, 2026)

## FEATURES:

* **New resource: `wallarm_rule_generator`** — generates Terraform rule files from `wallarm_hits` data source output or existing API rules (`source = "hits"` / `source = "api"`)
* **New resource: `wallarm_action`** — read-only resource for manual action scope tracking
* **New resource: `wallarm_hits_index`** — persistent index for tracking fetched request IDs, enables hits-to-rules caching workflow
* **New data source: `wallarm_actions`** — discovers all non-empty action scopes with pagination
* **New example: `hits-to-rules`** — creates false positive suppression rules from hit data with persistent state caching
* **New example: `import-ip-lists`** — imports all existing IP list entries into Terraform state
* **IP list import by application scope** — new import ID format `{clientID}/subnet/{expiredAt}/apps/{appIDs}` groups subnets by both expiration and application IDs
* **Credential stuffing cache** — shared cache at `ProviderMeta` level, eliminates redundant API calls
* **Gzip compression** — all API requests use `Accept-Encoding: gzip` (~19x payload reduction)

## IMPROVEMENTS:

* IP list cache at `ProviderMeta` level with per-rule-type fetching and Create serialization
* Centralized API limits in `constants.go`
* Action scope validation in `CustomizeDiff`
* `variativity_disabled` changed to `Optional+Default:true` for consistent import behavior
* `[multiple]` path handling — hits spanning multiple paths produce HOST-header-only scope
* Action conditions mismatch error prints both condition sets inline
* Lint fixes: extracted string constants, pre-allocated slices, checked `d.Set()` returns

## SECURITY:

* Added `Sensitive: true` to node `token` field in resource and data source
* Added `Sensitive: true` to webhook integration `headers` field
* Safe type assertions in `tftoapi` mapper — returns errors instead of panicking on corrupted state

## BUG FIXES:

* Fixed IP list import merging entries with different application scopes into one resource
* Fixed IP list import chunk index: chunk 0 now correctly includes `/0` suffix when chunking is needed
* Fixed `context.TODO()` in 52 CRUD functions — now passes actual `ctx`
* Fixed unchecked `d.Set()` on aggregate types
* Fixed hint cache pagination slice reuse bug
* Fixed `HitFilter.AttackID` type from `[]string` to `[][]string`

## DOCUMENTATION:

* New guide: `docs/guides/hits_to_rules.md`
* New docs: `hits_index`, `rule_generator` resources
* Added ephemeral hits warning to `data.wallarm_hits` doc
* Added common fields section to action guide
* Fixed `integration_ms_teams.md` resource name, `rule_file_upload.md` name and field optionality
* Rewrote `global_mode.md` and `rules_settings.md` with complete field references

# v2.1.0 (Mar 17, 2026)

## NOTES:

* Completed migration to Terraform Plugin SDKv2
* Added IP Lists Auto Import
* Improved logging system
* Implemented APIError type
* Added errors handling for 5xx and some 4xx responses
* Added 'attack' mode to fetch hits by data source
* Fixed bugs
* Updated tests
* Updated documentation
* Added a module for interaction with Wallarm API: `./examples/terraform-wallarm-api`

# v2.0.0 (Mar 12, 2026)

## NOTES:

* Updated Terraform Plugin SDK to v2
* Updated IP List to support all current API source types
* Updated existing integrations
* Removed Scanner
* Added data_source_rules for bulk rules imports
* Added data_source_applications for bulk applications imports
* Added retries for rules CREATE/UPDATE/DELETE methods to handle snapshot 423 error
* Fixed tests
* Fixed bugs

# v1.9.0 (Mar 6, 2026)

## NOTES:

* Added new data source `data_source_wallarm_hits.go`
* Added middleware between API and terraform provider `hint_cache.go`
* Bug fixes

# v1.8.6 (Feb 12, 2026)

## NOTES:

* Added new resource `wallarm_rule_disable_stamp`

# v1.8.5 (Feb 06, 2026)

## NOTES:

* Add updating comment and variativity_disabled
* Patch wrapPointElements -> support gql params

# v1.8.4 (Jan 28, 2026)

## NOTES:

* Resource rule cred stuff is removed

# v1.8.3 (Jan 28, 2026)

## NOTES:

* Resource rules import refactored
* Go mod updated 1.23 -> 1.24

# v1.8.2 (Jul 24, 2025)

## NOTES:

* Added new rules
* Code refactored

# v1.8.1 (Jul 24, 2025)

## NOTES:

* Added fields 'set', 'active', 'title', 'mitigation' for all resource_rules
* Code refactored

# v1.8.0 (April 17, 2025)

## NOTES:

* Code refactored
* Go mod updated 1.15 -> 1.23

# v1.7.0 (April 15, 2025)

## NOTES:

* Updated event_type: 'vuln' -> 'vuln_high' in docs/resources
* Removed unused resources 'attack_rechecker', 'attack_rechecker_rewrite'
* Extended validation for resourceWallarmTrigger, added check 'forced_browsing_started'
* Fixed lock_time for triggers, now it fills only for action_id 'block_ips' or 'add_to_graylist'

# v1.6.0 (December 6, 2024)

## NOTES:

* Added import support
* Added OverlimitResSettingsRule resource
* Changed rule vpatch resource according to api
* Changed rule set response header resource according to api
* Fixed some api methods

# v1.5.0 (September 1, 2024)

## NOTES:

* Added integration with DataDog
* Added integration with Telegram
* Added integration with MS Teams
* Added API Spec resource for managing API specifications
* Added Rate Limit rule
* Added support for all triggers that we have in the UI

# v1.1.0 (May 24, 2023)

## BUG FIXES:

* Fixed intergation api methods
* Fixed rule api methods
* Fixed trigger api methods

## NOTES:

* Added support for the following new resources: `wallarm_allowlist`, `wallarm_graylist`
* Added support of the query field in rules
* Change api authentication method using X-WallarmAPI-Token

# v1.0.0 (November 3, 2022)

## BUG FIXES:

* Fixed the following resources that were broken by changes in the API: `wallarm_blacklist`, `wallarm_rule_bruteforce_counter`, `wallarm_rule_dirbust_counter`, `wallarm_global_mode`
* Fixed various other bugs

## NOTES:

* Added support for the following new resources: `wallarm_rule_binary_data`, `wallarm_rule_disable_attack_type`, `wallarm_rule_parser_state`, `wallarm_rule_uploads`, `wallarm_rule_variative_keys`, `wallarm_rule_variative_values`
* Renamed the resource `wallarm_blacklist` to `wallarm_denylist`
* Updated the docs

# v0.0.10 (September 10, 2021)

## BUG FIXES:

* Fixed Opsgenie integration resource

## NOTES:

* Minor changes in the dependencies
* Fixed broken links and typos in the docs

# v0.0.9 (January 8, 2021)

## NOTES:

* Fix bug with the incorrect state
* Now `PASSPHRASE` for Release workflow is taken from `secrets.PASSPHRASE` as per best practices
* Added support of `go 1.15` in tests
* Two newly created resources: `wallarm_rule_bruteforce_counter` and `wallarm_rule_dirbust_counter`. The source code and documentation have been included altogether.
* `Trigger` arbitrary time value replaced by enum type
* New `wallarm-go` library structure approach

# v0.0.8 (December 16, 2020)

## NOTES:

* Minor changes in markdown

# v0.0.7 (October 23, 2020)

## NOTES:

* Documentation layout have been modified

# v0.0.3 (October 10, 2020)

## NOTES:

* Automatic GO release have been added

# v0.0.2 (September 16, 2020)

## FEATURES:

**New Resource:** `wallarm_blacklist` - is used to manage the blacklist

## ENHANCEMENTS:

**Resource:** `wallarm_rule_*` - bumped a client library version to comply with the new struct for `[][]interface{}`

## BUG FIXES:

**Resource:** - fixed 400 HTTP response code when there is an incorrect JSON body for requests on the update call

# v0.0.1 (September 10, 2020)

## NOTES:

* The first public release
