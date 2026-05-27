# Hits, Attacks & False Positive Suppression

## Domain model

A **hit** represents a single detected threat within an HTTP request. Hits sharing the same HTTP request are linked by `request_id`. Hits from the same attack campaign share an `attack_id`.

**IMPORTANT: Hits are ephemeral** — they have a retention period and can be dropped from the API at any time.

## False positive workflow

1. **Fetch**: `wallarm_hits` data source retrieves hits for given `request_id`(s)
2. **Group by Action**: Hits grouped by Host header + URI path (the Action scope)
3. **Group by Point**: Within each action, grouped by detection point
4. **Generate Rules**: Two rule types for FP suppression:
   - **`disable_stamp`** — allows specific attack signatures (stamps) at a given point
   - **`disable_attack_type`** — allows specific attack types at a given point
5. **One resource per rule**: Each stamp and each attack_type is a separate Terraform resource, matching the API 1:1. The `for_each` key is `{action_hash}_{point_hash}_{attack_type}_{stamp}` for stamp rules or `{action_hash}_{point_hash}_{attack_type}` for attack_type rules. Hash prefixes are 16 hex chars.

**Stampless attack types:** `xxe` and `invalid_xml` do not produce stamps. Hits of these types can only be suppressed via `disable_attack_type` rules.

## Data source: `wallarm_hits`

**Input**: `request_id` (single string) + `mode` variable (`"request"` or `"attack"`). Called per-request_id via `for_each` in HCL.

**Hit filtering — allowed attack types:**
`xss`, `sqli`, `rce`, `ptrav`, `crlf`, `redir`, `nosqli`, `ldapi`, `scanner`, `mass_assignment`, `ssrf`, `ssi`, `mail_injection`, `ssti`, `xxe`, `invalid_xml`

**Key computed outputs:**
- `aggregated` — compact JSON with `action_hash` (16 chars), `action` conditions, and `groups` (each keyed by `point_hash_16 + "_" + attack_type`, containing `stamps`, `attack_type`, and `disable_attack_type` bool controlled by `rule_types` filter)
- `action_hash` — Ruby-compatible `ConditionsHash`
- Action validation via `ActionReadByHitID` hash comparison

## Hits-to-rules flow

Three components: `wallarm_hits_index` (gating), `data.wallarm_hits` (fetching), `terraform_data.cache` (persistence). Deduplication by action_hash in HCL locals. See `docs/guides/hits_to_rules.md`.
