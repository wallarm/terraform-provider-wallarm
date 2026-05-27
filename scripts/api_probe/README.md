# api_probe

Discovery tool for Wallarm rule-creation API constraints. For every rule type listed in the `probes` table inside `main.go`, the script:

1. **POSTs `/v1/objects/hint/create`** with a starting body (`type` + minimal action), retrying with field additions when the API returns "can't be blank" / "should be in N..M" errors. Records the minimal Create body, required fields, range constraints, and the API-defaulted fields the API filled in.
2. **PUTs `/v3/hint/{id}`** once with the same body it just succeeded on, to surface Update-side schema drift (fields the Update endpoint requires/rejects that Create doesn't).
3. *(Optional)* If `API_PROBE_MUTABILITY=1`, sends per-field PUTs with each API-defaulted field flipped to a different value, then compares the response body to confirm whether the field is mutable, immutable-but-silently-dropped, or rejected.
4. **DELETEs** the rule it just created via `/v1/objects/hint/delete` so probing leaves no state behind.

Output: a markdown report listing per-rule-type results — handy ground truth for schema decisions (see `references/schema-decisions.md`).

## Required env vars

| Var | Purpose |
|-----|---------|
| `WALLARM_API_HOST` | API endpoint, e.g. `https://api.wallarm.com` |
| `WALLARM_API_TOKEN` | API token with rule-create + rule-delete permission |
| `WALLARM_API_CLIENT_ID` | Numeric client/tenant ID to probe under |

The repo's `acc_test.env` already exports these for the test tenant; sourcing it is the easiest setup.

## Optional env vars

| Var | Effect |
|-----|--------|
| `API_PROBE_RULES` | Comma-separated allowlist of rule types to probe (default: all). Supports both raw rule types (`graphql_detection`) and labelled probes (`brute_exact`). |
| `API_PROBE_MUTABILITY` | `1` enables the per-field mutability probe (PUT each API-defaulted field flipped to an alternate value). Off by default — extra API calls per rule type. |
| `API_PROBE_VERBOSE` | `1` logs every HTTP request and response. |
| `API_PROBE_OUT` | Output path for the markdown report. Default: `api_probe_results.md` in the current working directory. |

## Build

```sh
go build ./scripts/api_probe/
```

Or run directly without building:

```sh
go run ./scripts/api_probe/
```

## Typical runs

**Full probe across every rule type, default report path:**

```sh
. ./acc_test.env
go run ./scripts/api_probe/
```

**Single rule type, with per-field mutability probe and a custom output path:**

```sh
. ./acc_test.env
API_PROBE_RULES=graphql_detection \
API_PROBE_MUTABILITY=1 \
API_PROBE_OUT=/tmp/graphql.md \
  go run ./scripts/api_probe/
```

**Verbose dry-run for debugging:**

```sh
API_PROBE_VERBOSE=1 API_PROBE_RULES=disable_stamp go run ./scripts/api_probe/
```

## Adding a new rule type

Append an entry to the `probes` slice in `main.go`. Each entry needs:

- `RuleType` — the API hint-type string (e.g. `"graphql_detection"`).
- `Base` — `map[string]any` of fields to send on the initial Create attempt. Usually just enough to prevent immediate "can't be blank" rejections (e.g. `{"mode": "block"}`).
- *(optional)* `Label` — disambiguator for cases where one type has multiple meaningful body shapes (e.g. `brute_exact` vs `brute_regexp`).
- *(optional)* `TokenEnv` — env var name carrying a token with elevated permissions, for rule types that require Administrator-extended (e.g. `disable_stamp`).

The retry loop's `candidateValues` map (also in `main.go`) provides values for fields the API reports as "can't be blank" — extend it if a new rule type requires a field that's not already covered.

## Output format

The script writes a markdown report with two sections:

1. **Summary table** — one row per rule type with success/failure, required fields, API-defaulted fields, and any range errors encountered.
2. **Per-rule-type detail** — full Create/Update outcomes, the wire payload that succeeded, and (when `API_PROBE_MUTABILITY=1`) the per-field mutability classifications.

The script writes results to `api_probe_results.md` in the working directory by default (override via the `API_PROBE_OUT` env var). Treat the output as a working snapshot of API ground truth — refresh by re-running the probe after any API update.

## Known limitations of the per-field mutability probe

`API_PROBE_MUTABILITY=1` sends per-field PUTs and compares the response. It's informative for fields where the server treats single-field flips the same as full-body updates, but the Wallarm API is sensitive to body shape in ways that produce noise:

- **`immutable_silent` on common fields** (`set`, `attack_type`, etc.) often means "the API silently drops this field when sent in a partial body" — NOT that the field is server-immutable. Real provider Updates send a fuller body via `wallarm.HintUpdateV3Params` (always including `comment`, `title`, `active`, `set`, `variativity_disabled` plus per-resource customizers); those updates DO mutate the field. Treat probe `immutable_silent` for common fields as likely noise; verify via an acceptance test before acting on it.
- **`rejected` on a field that has no API default** (showed as `<nil>` in the API echo) is generally trustworthy — the field-flip target is rejected outright, distinct from silent-drop.
- **`untested`** — the probe couldn't pick a safe alternate (the field's current value is `nil` and the field name isn't in `enumAlternates` or `candidateValues`). To improve coverage for a specific field, add it to one of those maps with a known-good probe value.

Trust the probe most for resource-specific int/bool/enum fields with concrete API defaults (e.g. `graphql_detection.max_*`, `rate_limit.delay/burst/time_unit`); be skeptical when it claims `immutable_silent` for the shared common-field surface.
