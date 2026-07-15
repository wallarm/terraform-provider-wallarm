# Rule-type API contract (probe ground truth)

Reference for what the Wallarm rule API actually requires, defaults, and
accepts per rule type. The provider's view (schema types, CRUD) is in
`rules-core.md`; this doc is the raw API contract used when designing a schema
field or choosing a wallarm-go struct tag.

**Source:** `POST /v1/objects/hint/create` + `PUT /v3/hint/{id}` probes against
`https://audit.api.wallarm.com`, client `110557`, via `scripts/api_probe/main.go`
(raw JSON, no wallarm-go struct in the way). **Probe date:** 2026-05-01.

## 1. Overview

Use this when: deciding Required vs Optional vs Computed for a rule field;
choosing the default Read should expect; deciding whether to validate a range at
plan time or defer to the API; or picking a wallarm-go field tag (`int` vs
`*int`, with/without `omitempty`).

## 2. Model

A Create body carries the discriminating fields for the rule type plus a scope
`action`; the API fills the rest with defaults and echoes them back. Two facts
shape schema design: the per-type `variativity_disabled` default (§4.1) and
which numeric fields treat `0` as a valid value (§4.2).

## 3. Elements

The probe records, per rule type: the **minimal body** (fields the API demands),
the **API-defaults filled** (echoed when omitted), any **range/enum**
constraints, **permissions**, and whether an idempotent **update with the same
body** succeeds. These are the reference tables in §6.

## 4. Behavior

### 4.1 Common defaults and the variativity split

When omitted from Create, the API echoes: `active=true`, `comment=null`,
`set=null`, `title=null`. `variativity_disabled` defaults **per rule type**:

| Default `true` | Default `false` |
|---|---|
| vpatch, wallarm_mode, api_abuse_mode, regex, experimental_regex, sensitive_data, set_response_header, brute, bola, enum, forced_browsing, rate_limit_enum, graphql_detection, rate_limit, brute_counter, dirbust_counter, bola_counter | disable_attack_type, disable_stamp, binary_data, uploads, parser_state, overlimit_res_settings, file_upload_size_limit |

The provider hardcodes `variativity_disabled=true` on Create regardless of type
(`rules-core.md §4.3`), a deliberate synchronization choice.

### 4.2 Zero-valid numeric fields (wallarm-go tag design)

`omitempty` on a non-pointer `int` drops a literal `0` from the wire, so a field
where `0` is a valid input must be `*int` in wallarm-go (else the API rejects
`can't be blank`). Classification from the probe:

| Field | Range | `0` valid? | Tag |
|---|---|---|---|
| `Rate`, `Burst`, `Delay` | 0..1000 | yes | `*int` |
| `OverlimitTime` | 0..MaxInt | yes | `*int` |
| `Size` | 1..2^64 | no | `int` |
| `RspStatus` | 400..599 | no | `int` |
| `Stamp` | >=1 | no | `int` |
| `MaxDepth`, `MaxValueSizeKb`, `MaxDocSizeKb`, `MaxDocPerBatch`, `MaxAliases` | 1..* | no | `int` (omitempty lets the API default win) |

The provider sends the zero-valid fields via pointers (`lo.ToPtr` for Required,
`GetPointerIfConfigured` for Optional), so a configured `0` reaches the wire.

## 5. Parameters

Per-rule fields are the reference data in §6; provider-side schema decisions for
them follow `schema-decisions.md`.

## 6. Reference data

Legend: minimal body = fields the API requires (all also take a scope `action`);
all types fill `active=true`, `comment=null`, `set=null`, `title=null` unless
noted.

| Rule type (provider resource) | Minimal body | Notable defaults / constraints |
|---|---|---|
| `vpatch` | `attack_type`, `point` | `variativity_disabled=true` |
| `wallarm_mode` | `mode` | `variativity_disabled=true` |
| `api_abuse_mode` | `mode` | `variativity_disabled=true` |
| `disable_attack_type` | `attack_type`, `point` | `variativity_disabled=false` |
| `disable_stamp` | `stamp`, `point` | `variativity_disabled=false`; **Administrator (extended)** - plain Admin gets 403 |
| `regex` | `attack_type`, `point`, `regex` | `variativity_disabled=true` |
| `experimental_regex` | `attack_type`, `point`, `regex` | same as `regex`, differs by `type` discriminator |
| `disable_regex` (`ignore_regex`) | `regex_id` (existing experimental_regex, same client), `point` | probe blocked - needs a real `regex_id` |
| `binary_data` | `point` | `variativity_disabled=false` |
| `sensitive_data` (`masking`) | `point` | `variativity_disabled=true` |
| `uploads` | `point`, `file_type` | `file_type` enum: `docs`/`html`/`images`/`music`/`video`; `variativity_disabled=false` |
| `set_response_header` | `mode`, `name`, `values` | `mode` enum: `append`/`replace`; `variativity_disabled=true` |
| `parser_state` | `parser`, `state`, `point` | `variativity_disabled=false` |
| `overlimit_res_settings` | `overlimit_time` | `mode` optional (default `monitoring`); `variativity_disabled=false` |
| `rate_limit` | `point`, `rate` (0..1000), `rsp_status` (400..599) | `burst`/`delay` optional (0..1000); `time_unit` optional (default `rps`); `suffix` read-only random echo; `variativity_disabled=true` |
| `brute` / `bola` / `enum` | `mode`, `enumerated_parameters`, `reaction`, `threshold` | regexp mode: all 5 `enumerated_parameters` fields required; exact mode: only `points`; mixing 400s; `variativity_disabled=true` |
| `forced_browsing` | `mode`, `reaction`, `threshold` | no `enumerated_parameters`; `attack_type=dirbust`; `variativity_disabled=true` |
| `rate_limit_enum` | `mode`, `reaction`, `threshold` | `attack_type=rate_limit`; `variativity_disabled=true` |
| `graphql_detection` | `mode` | see §6.1; `variativity_disabled=true` |
| `file_upload_size_limit` | `size` (1..2^64, `0` rejected) | `mode` optional (default `monitoring`); `size_unit` optional (default `b`); `variativity_disabled=false` |
| `brute_counter` / `dirbust_counter` / `bola_counter` | none (type + action only) | `point` not allowed; `counter=<prefix>:<client>:<hex>` (`b:`/`d:`/`i:`); cannot be updated or deleted directly (trigger-bound, API-lifecycle) |

### 6.1 `graphql_detection` defaults and ranges

API defaults: `debug_enabled=true`, `introspection=true`, `max_aliases=5`,
`max_depth=10`, `max_doc_per_batch=10`, `max_doc_size_kb=100`,
`max_value_size_kb=10`. Ranges: `max_depth` 1..100, `max_value_size_kb` 1..100,
`max_doc_size_kb` 1..1024, `max_doc_per_batch` 1..100.

- The GraphQL alias limit is the API field `max_aliases` (default 5); the
  provider schema and wallarm-go use the same name (aligned in wallarm-go
  v0.12.2). Roadmap **R6** tracks a create-rejection investigated under this
  field.
- Making the omitted bool defaults (`introspection`, `debug_enabled`) round-trip
  cleanly is roadmap **R5**.

### 6.2 Known follow-ups

- `rate_limit.suffix` - read-only API field, not in the provider schema (Read
  filters it via `setIfExists`). Whether to expose it is roadmap **R3**.

## 7. References

- `rules-core.md` - the provider-side rule model and CRUD.
- `schema-decisions.md` - how these API facts map to schema attributes.
- `proton-types.md` - `attack_type` and `file_type` enum sources.
- `scripts/api_probe/main.go` - the probe that produced this data.
