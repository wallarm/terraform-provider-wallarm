# Wallarm API — Rule-Type Reference (ground truth)

**Source:** `POST /v1/objects/hint/create` + `PUT /v3/hint/{id}` probes against `https://audit.api.wallarm.com`, client `110557`.
**Probe script:** `scripts/api_probe/main.go` (run with `go run`).
**Probe date:** 2026-05-01.

This document captures what the API actually does — required vs optional, value ranges, server-side defaults, and the minimal body needed to create each rule type. The probe sends raw JSON via `net/http` (no wallarm-go struct in the way), so what's recorded here is the API contract, not the provider's view of it.

Use this when:
- Designing schema changes for a rule resource (Required vs Optional vs Computed).
- Choosing default values to expect from Read.
- Validating ranges at plan time vs leaving to the API.
- Verifying a wallarm-go struct field tag (`int` vs `*int`, with/without `omitempty`).

---

## Common defaults across all rule types

When the field is omitted from the Create body, the API echoes back:

| Field | API default | Notes |
|---|---|---|
| `active` | `true` | Provider's `resourcerule.Create` was sending `false` — fixed in v2.3.8 Batch 9. |
| `comment` | `null` | API allows null; provider sends `"Managed by Terraform"` from `commonResourceRuleFields.Default`. |
| `set` | `null` | |
| `title` | `null` | |
| `variativity_disabled` | varies (see per-rule table) | Some rule types default to `true`, others to `false`. Provider hardcodes `true` in `resourcerule.Create:193`. |

`variativity_disabled` defaults per rule type:

| Default `true` | Default `false` |
|---|---|
| vpatch, wallarm_mode, api_abuse_mode, regex, experimental_regex, sensitive_data, set_response_header, brute, bola, enum, forced_browsing, rate_limit_enum, graphql_detection, rate_limit, brute_counter, dirbust_counter, bola_counter | disable_attack_type, disable_stamp, binary_data, uploads, parser_state, overlimit_res_settings, file_upload_size_limit |

---

## Per-rule reference

### `vpatch`

- **Minimal body:** `attack_type`, `point` (+ scope `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `wallarm_mode`

- **Minimal body:** `mode` (+ `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `api_abuse_mode`

- **Minimal body:** `mode` (+ `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `disable_attack_type`

- **Minimal body:** `attack_type`, `point` (+ `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=false`
- **Update with same body:** ✅ ok

### `disable_stamp`

- **Minimal body:** `stamp`, `point` (+ `action`)
- **Permissions:** Administrator (extended) — Admin alone returns 403.
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=false`
- **Update with same body:** ✅ ok

### `regex`

- **Minimal body:** `attack_type`, `point`, `regex` (+ `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `experimental_regex`

- **Minimal body:** `attack_type`, `point`, `regex` (+ `action`)
- Same shape as `regex`; differs only via the `type` discriminator.

### `disable_regex` (provider: `wallarm_rule_ignore_regex`)

- **Minimal body:** `regex_id` referencing an existing experimental_regex rule of the same client (+ `point`, `action`).
- **Probe blocked:** the probe doesn't have a real `regex_id` to point at — extend the probe with `DISABLE_REGEX_ID` env var to retest.

### `binary_data`

- **Minimal body:** `point` (+ `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=false`
- **Update with same body:** ✅ ok

### `sensitive_data` (provider: `wallarm_rule_masking`)

- **Minimal body:** `point` (+ `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `uploads`

- **Minimal body:** `point`, `file_type` (+ `action`)
- **`file_type` enum:** `docs`, `html`, `images`, `music`, `video`
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=false`
- **Update with same body:** ✅ ok

### `set_response_header`

- **Minimal body:** `mode`, `name`, `values` (+ `action`)
- **`mode` enum:** `append`, `replace` (different from rule-mode enum used by other types).
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `parser_state`

- **Minimal body:** `parser`, `state`, `point` (+ `action`)
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=false`
- **Update with same body:** ✅ ok

### `overlimit_res_settings`

- **Minimal body:** `overlimit_time` (+ `action`)
- **`mode` is OPTIONAL** — API defaults to `monitoring` when omitted.
- **API-defaults filled:** `active=true`, `comment=null`, `mode=monitoring`, `set=null`, `title=null`, `variativity_disabled=false`
- **Update with same body:** ✅ ok

### `rate_limit`

- **Minimal body:** `point`, `rate`, `rsp_status` (+ `action`).
- **`rate` is REQUIRED, range 0..1000** (`should be in 0..1000, can't be blank`).
- **`rsp_status` is REQUIRED, range 400..599** (`should be in 400..599, can't be blank`).
- **`burst` is OPTIONAL** — API didn't ask for it. Provider currently marks it Required.
- **`time_unit` is OPTIONAL** — API defaults to `rps` when omitted. Provider marks it Required.
- **`delay` is OPTIONAL**, range 0..1000.
- **API-defaults filled:** `active=true`, `comment=null`, `set=null`, `suffix=<random>`, `time_unit=rps`, `title=null`, `variativity_disabled=true`
- **Read-only field `suffix`** echoed by API; not in provider schema.
- **Update with same body:** ✅ ok
- **Critical bug:** with provider's wallarm-go struct (`Rate int json:"rate,omitempty"`), user-typed `rate=0` is silently dropped on the wire — API rejects "can't be blank". Same for `burst=0`, `delay=0`. Direct API probe with `rate=0` succeeds, confirming the wire-shape is the bug, not the API.

### `brute` — `mode = "regexp"`

- **Minimal body:** `mode`, `enumerated_parameters{mode,name_regexps,value_regexps,additional_parameters,plain_parameters}`, `reaction`, `threshold` (+ `action`).
- **All five `enumerated_parameters` fields are REQUIRED** in regexp mode.
- **API-defaults filled:** `active=true`, `advanced_conditions=[]`, `arbitrary_conditions=[]`, `attack_type=brute`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `brute` — `mode = "exact"`

- **Minimal body:** `mode`, `enumerated_parameters{mode,points}`, `reaction`, `threshold` (+ `action`).
- Sending `name_regexps`/`value_regexps`/`additional_parameters`/`plain_parameters` in exact mode → 400 "fields for regexp mode should contain only [...]". Same for `points` in regexp mode (unverified directly here but enforced by the v2.3.8 Batch 8 plan-time validator).
- **Update with same body:** ✅ ok

### `bola` (regexp + exact)

- Same shape as `brute`.

### `enum` (regexp + exact)

- Same shape as `brute`.

### `forced_browsing`

- **Minimal body:** `mode`, `reaction`, `threshold` (+ `action`).
- No `enumerated_parameters` block.
- **API-defaults filled:** `active=true`, `advanced_conditions=[]`, `arbitrary_conditions=[]`, `attack_type=dirbust`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `rate_limit_enum`

- **Minimal body:** `mode`, `reaction`, `threshold` (+ `action`).
- **API-defaults filled:** `active=true`, `advanced_conditions=[]`, `arbitrary_conditions=[]`, `attack_type=rate_limit`, `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ✅ ok

### `graphql_detection`

- **Minimal body:** `mode` (+ `action`).
- **API-defaults filled:**
  - `active=true`
  - `debug_enabled=true`
  - `introspection=true`
  - `max_aliases=5`
  - `max_depth=10`
  - `max_doc_per_batch=10`
  - `max_doc_size_kb=100`
  - `max_value_size_kb=10`
  - `comment=null`, `set=null`, `title=null`, `variativity_disabled=true`
- **Range constraints (from Batch 10 attempted-update bug):**
  - `max_depth` 1..100
  - `max_value_size_kb` 1..100
  - `max_doc_size_kb` 1..1024
  - `max_doc_per_batch` 1..100
  - `max_alias_size_kb` (alias for API's `max_aliases`) — range unknown
- **API field name `max_aliases` ≠ provider/wallarm-go field `max_alias_size_kb`.** Worth verifying this is a deliberate alias.
- **Update with same body:** ✅ ok

### `file_upload_size_limit`

- **Minimal body:** `size` (+ `action`).
- **`size` is REQUIRED, range 1..2^64 bytes.** `size=0` rejected as `should be in range of (1..18446744073709551616) bytes`. `size=999999999` accepted. **0 is never valid → keep `int+omitempty` in wallarm-go.**
- **API-defaults filled:** `active=true`, `comment=null`, `mode=monitoring`, `set=null`, `size_unit=b`, `title=null`, `variativity_disabled=false`
- **`mode` is OPTIONAL** — API defaults to `monitoring`. Provider marks it Required.
- **`size_unit` is OPTIONAL** — API defaults to `b` (bytes).
- **Update with same body:** ✅ ok

### `brute_counter` (provider: `wallarm_rule_bruteforce_counter`)

- **Minimal body:** `(none — type and action only)`.
- **`point` is NOT ALLOWED** at the API level for this counter.
- **API-defaults filled:** `active=true`, `comment=null`, `counter=b:<client>:<hex>`, `set=null`, `title=null`, `variativity_disabled=true`
- **Update with same body:** ❌ 400 — `comment: "must contain at least one of keys"`. Counters can't be updated directly via this route; they're trigger-bound and lifecycle-managed by the API. Provider correctly never calls Update.

### `dirbust_counter`

- Same shape as `brute_counter`. `counter` prefix `d:`.

### `bola_counter`

- Same shape as `brute_counter`. `counter` prefix `i:`.

---

## Provider schema deltas (audit)

Where the provider currently disagrees with the API. ✅ = matches. ⚠️ = drift.

### `wallarm_rule_rate_limit`

| Field | Provider | API | Verdict |
|---|---|---|---|
| `rate` | Required, IntBetween(0,1000) | Required, 0..1000 | ✅ but **wire bug** — `rate=0` silently dropped (`int+omitempty`) |
| `burst` | Required, IntBetween(0,1000) | Optional | ⚠️ over-restricts; **wire bug** for `burst=0` |
| `delay` | Optional, IntBetween(0,1000) | Optional | ✅ but **wire bug** for `delay=0` |
| `rsp_status` | Optional, IntBetween(400,599) | Required, 400..599 | ⚠️ should be Required (apply-time error today) |
| `time_unit` | Required, StringInSlice([rps,rpm]) | Optional, default `rps` | ⚠️ over-restricts |

### `wallarm_rule_overlimit_res_settings`

| Field | Provider | API | Verdict |
|---|---|---|---|
| `mode` | Required | Optional, default `monitoring` | ⚠️ over-restricts |

### `wallarm_rule_file_upload_size_limit`

| Field | Provider | API | Verdict |
|---|---|---|---|
| `mode` | (need to check) | Optional, default `monitoring` | ⚠️ likely over-restricts |
| `size_unit` | (need to check) | Optional, default `b` | ⚠️ likely over-restricts |

### `wallarm_rule_graphql_detection`

| Field | Provider | API | Verdict |
|---|---|---|---|
| `max_depth` | Optional, no Computed | Optional, default 10 | ⚠️ needs `Computed: true` (Batch 10 plan) |
| `max_value_size_kb` | Optional, no Computed | Optional, default 10 | ⚠️ needs `Computed: true` |
| `max_doc_size_kb` | Optional, no Computed | Optional, default 100 | ⚠️ needs `Computed: true` |
| `max_doc_per_batch` | Optional, no Computed | Optional, default 10 | ⚠️ needs `Computed: true` |
| `max_alias_size_kb` | Optional+Computed+ForceNew | Optional, default 5 (API field name `max_aliases`) | ✅ already Computed; verify wire-tag mapping |
| `introspection` | Optional, no Computed | Optional, default `true` | ⚠️ needs `Computed: true` + helper change so omitted bool isn't sent as `false` |
| `debug_enabled` | Optional, no Computed | Optional, default `true` | ⚠️ same |

### Common (`commonResourceRuleFields`)

| Field | Provider | API | Verdict |
|---|---|---|---|
| `active` | Optional+Computed | Optional, default `true` | ✅ matches; Create-side default `true` fixed in Batch 9 |
| `variativity_disabled` | Optional, Default `true` | Default varies per rule type (see common-defaults table above) | ⚠️ provider hardcodes `true` in `resourcerule.Create`; for rule types where API default is `false`, this silently overrides |

---

## wallarm-go struct deltas (audit)

The wire bug for `rate=0`/`burst=0`/`delay=0` is rooted in `wallarm-go/action.go:101`-onward. The relevant `ActionCreate` fields use `int` (non-pointer) with `,omitempty`:

```go
Delay     int    `json:"delay,omitempty"`
Burst     int    `json:"burst,omitempty"`
Rate      int    `json:"rate,omitempty"`
RspStatus int    `json:"rsp_status,omitempty"`
OverlimitTime int `json:"overlimit_time,omitempty"`
Stamp     int    `json:"stamp,omitempty"`
Size      int    `json:"size,omitempty"`
MaxDepth  int    `json:"max_depth,omitempty"`
// ... more max_* fields
```

`omitempty` drops `0` for non-pointer numeric types. Fields where `0` is a valid user value need `*int`:

| Field | Range | 0 valid? | Action |
|---|---|---|---|
| `Rate` | 0..1000 | ✅ | change to `*int` |
| `Burst` | 0..1000 | ✅ | change to `*int` |
| `Delay` | 0..1000 | ✅ | change to `*int` |
| `OverlimitTime` | 0..MaxInt | ✅ | change to `*int` |
| `Size` | unknown | likely | change to `*int` (verify range) |
| `RspStatus` | 400..599 | ❌ | leave as `int` (0 invalid by definition) |
| `Stamp` | ≥1 | ❌ | leave as `int` |
| `MaxDepth`, `MaxValueSizeKb`, `MaxDocSizeKb`, `MaxDocPerBatch`, `MaxAliasesSizeKb` | 1..* | ❌ | leave as `int` (omitempty correctly lets API default win on omission) |

`HintUpdateV3Params` already uses `*int` for these fields — only `ActionCreate` is broken.

The corresponding provider Create paths need to switch from `Rate: rate` (passing int) to `Rate: lo.ToPtr(rate)` (passing *int). For Required fields the provider can always send a pointer; for Optional fields, the provider should send pointer only when the user configured the field (via `GetRawConfig` or equivalent).

---

## Summary of action items

### v2.3.8 — minimum to ship safely
1. Block on the `rate=0` / `burst=0` / `delay=0` wire bug — **needs wallarm-go release** for `*int` fields. Either bump wallarm-go to v0.12.1 with the struct change + provider plumbing, or document the limitation and ship.

### v2.3.9 — schema actualisation (full audit)
Bring schemas in line with API:
- `wallarm_rule_rate_limit`: make `rsp_status` Required, `burst` Optional, `time_unit` Optional+Computed.
- `wallarm_rule_overlimit_res_settings`: `mode` Optional+Computed.
- `wallarm_rule_file_upload_size_limit`: `mode` Optional+Computed, `size_unit` Optional+Computed.
- `wallarm_rule_graphql_detection`: add `Computed: true` to `max_depth`, `max_value_size_kb`, `max_doc_size_kb`, `max_doc_per_batch`, `introspection`, `debug_enabled`.
- `resourcerule.Create:193`: stop hardcoding `variativity_disabled: true`; honor user value with API-default fallback per rule type.
- Rename/clarify `max_alias_size_kb` ↔ API `max_aliases` mapping; verify wire tag.

### Hidden-field decisions
- `rate_limit.suffix` — read-only API field. Decide whether to expose as Computed or continue ignoring. Ignoring is safer; Read already filters via `setIfExists`.

### Probe extensions for next time
- Run `delay`/`burst` boundary tests (lower=0, upper=1000, out-of-range=1001).
- Add `disable_regex` support via `DISABLE_REGEX_ID` env var.
- Stress-test counter Update behavior for completeness.
