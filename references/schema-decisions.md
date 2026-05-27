# Schema Decision Rules

How to choose Terraform Plugin SDKv2 schema attributes (`Required`, `Optional`, `Computed`, `Default`, `ForceNew`, `Sensitive`, `ValidateDiagFunc`, `CustomizeDiff`, etc.) given the underlying API field's characteristics. Codified from lessons learned during provider development (notably the v2.3.8 / v2.3.9 schema-shape audits).

Use this file:
- When **adding a new resource field** — walk the decision tree in §E.
- When **auditing existing fields** for consistency — use §A–D as the reference matrix and §F to spot anti-patterns.
- When **debugging plan/apply UX bugs** — symptoms like "removing the HCL line doesn't restore default", "import triggers destroy+recreate", or "perpetual diff after API echo" usually trace to one of the anti-patterns in §F.

The rules are deliberately API-agnostic. Concrete Wallarm-specific examples appear only in §F to illustrate failure modes.

---

## A. The primary axis — who supplies the value?

| API characteristic | Schema attributes | Why |
|--------------------|-------------------|-----|
| API **rejects** Create without it | `Required: true` | SDK fails plan if omitted. No state, no Default, no Computed allowed. |
| API **accepts** Create without it; has a **stable, documented default** | `Optional: true, Default: <value>` | SDK fills the schema default at plan time → wire payload includes the value → symmetric "remove HCL line restores default". |
| API **accepts** Create without it; default is **undocumented / unstable / unknown** | `Optional: true, Computed: true` (no Default) | API echo populates state. HCL omission preserves state (sticky). |
| Field is **purely server-computed** (user cannot set it) | `Computed: true` alone (no Optional) | Pure output. SDK rejects HCL containing this attribute. |
| Server-derived ID / timestamp / read-only metadata | `Computed: true` | Exposed in state for cross-resource references. |

## B. Lifecycle / mutability modifiers

| API characteristic | Schema modifier | Combine with |
|--------------------|----------------|--------------|
| Field is **immutable after Create** (mutating requires destroy+recreate) | `ForceNew: true` | Required, or Optional+Computed (rarely Optional+Default — see anti-pattern 3) |
| Field is a **secret / credential** | `Sensitive: true` | Any shape — masks the value in plan output |

## C. Value-shape modifiers

| API characteristic | Schema attribute | When |
|--------------------|------------------|------|
| API has a **range or enum** constraint | `ValidateDiagFunc` | Plan-time fail instead of API rejection — better UX |
| Two fields are **mutually exclusive** | `ConflictsWith: []string{"other"}` | Pairwise constraint |
| Constraint is **conditional** (field A allowed only when field B has value X) | `CustomizeDiff` | Plan-time validation; can also `SetNew` to force defaults |
| API **normalizes** case / format (lowercases, uppercases, trims) | `StateFunc` (normalize on store) and/or `DiffSuppressFunc` (ignore semantically-equal diffs) | Prevents perpetual diff |

## D. The "zero value is meaningful" rule (per primitive type)

The treatment differs by type because SDKv2 has type-specific normalisation quirks. Pick the row that matches your field's type AND whether the zero value is a legitimate user input.

> **Top-level vs nested-block caveat (applies to all of D.1–D.3).** SDKv2 represents top-level resource fields with the modern proto state, which supports `cty.NullVal` — so an Optional-only top-level int / bool / string can be genuinely "absent" in state, and `terraform import` + `-generate-config-out` will emit no line for it. **Sub-fields inside a nested `TypeList`/`TypeSet`-of-`Resource` block use the legacy flat-state model** (`block.0.field = "0"`), which has no null slot for `TypeInt`/`TypeBool`. Every defined sub-field always materialises with the type's zero value when the API omits it, regardless of whether the schema is Optional, Optional+Computed, or Optional+Default. None of the rows below change this; if the field is a nested-block sub-field with a strict-range validator, see anti-pattern 9.

### D.1 — Bool

| Scenario | Schema | Why |
|----------|--------|-----|
| `false` is a meaningful user input | **`Optional + Default(<value>)`** — pick the API default if stable, else a chosen policy default | Symmetric `remove line → restore default` UX. Side-steps any potential SDKv2 normalisation of `field = false` to null on `Optional+Computed` (untested for top-level bool, but the documented string-list normalisation makes the analogous risk plausible). |
| `false` is NOT a meaningful user input (opt-in toggle, user only ever sets `true`) | `Optional + Computed` | Rare. Field is "user opts in or accepts API value" — Computed lets API echo win on omit. |
| Pure server-computed | `Computed` alone | Read-only. |

**Avoid `Optional+Computed` for bools when `false` is a valid user value.** Default-based gives strictly better UX (symmetric removal) at the cost of legacy-CLI-import asymmetry, which the modern `import{}+-generate-config-out` flow doesn't hit.

### D.2 — String

| Scenario | Schema | Why |
|----------|--------|-----|
| `""` (empty) is a meaningful "clear" user input | **`Optional` alone** (no Computed, no Default) | SDKv2 normalises HCL `field = ""` to `cty.NullVal`. With `Optional+Computed`, Computed semantics then preserve state — `field = ""` silently fails to clear. Documented as anti-pattern 7. |
| `""` is NOT meaningful (user always sets a non-empty value, or API default isn't `""`) | `Optional + Default(<value>)` if API has stable non-empty default, else `Optional + Computed` | Same as bool D.1. |
| Pure server-computed (read-only echo) | `Computed` alone | E.g. `rule_type`, `mitigation`. |

### D.3 — Int

| Scenario | Schema | Wire | Provider helper |
|----------|--------|------|-----------------|
| `0` is a meaningful user input AND API has a non-zero stable default (provider must not overwrite it) | **`Optional` alone** (no Default; Computed if user-omit should preserve API value) | wallarm-go field `*int` with `json:",omitempty"` | `GetPointerIfConfigured[int]` — uses `d.GetRawConfig()` to distinguish "user wrote 0" from "user omitted". Send nil on omit so omitempty drops the field and the API default wins. |
| `0` is NOT a meaningful user input (range starts at 1+) AND API has a stable default | `Optional + Default(<API default>)` | wallarm-go field `int` (no pointer needed) | `d.Get` is fine. |
| `0` is meaningful AND no stable API default | Same as row 1 (pointer + helper). | | |
| Pure server-computed | `Computed` alone | | |

**Why ints are different from bools/strings:** the `Optional+Default(0)` shape would send `0` on every Create/Update, masking the API's non-zero default. Use `*int+omitempty` + `GetPointerIfConfigured` to distinguish "user omitted" from "user wrote 0", so omit-on-wire when omitted-in-HCL.

`d.Get` cannot tell these apart for an Optional field — `RawConfig` is the only authoritative source. The pointer-on-wire trick lets you transmit a literal zero when the user wants it.

## E. Decision tree (apply in order)

```
1. Can the user set the value?
   NO  → Computed: true (alone). Done.
   YES → continue.

2. Will the API reject Create without it?
   YES → Required: true. (Add ForceNew / Sensitive / Validate as needed.) Done.
   NO  → continue.

3. Is the type's zero value (false / "" / 0) a legitimate user input?
   NO  → continue to step 4.
   YES → branch by type:
         BOOL   → Optional + Default(<value>). Done.
         STRING → Optional alone (no Computed; no Default unless policy-driven). Done.
         INT    → Optional + GetPointerIfConfigured[int] + *int+omitempty wire. Done.

4. Does the API have a stable, documented default?
   YES → Optional + Default(<value>). (Avoid ForceNew here — see anti-pattern 3.)
   NO  → Optional + Computed.

5. Always-applicable layer:
   ForceNew, Sensitive, ValidateDiagFunc, ConflictsWith, CustomizeDiff,
   DiffSuppressFunc, StateFunc.
```

**Step 3 reads from §D** — see the per-type tables for nuance (bool with default, string with policy default, int with vs. without API default).

## F. Anti-patterns (each has bitten us at least once)

1. **`Optional + Computed + Default` together.** SDKv2 rejects this combo; Default and Computed are mutually exclusive.
2. **`Optional + Computed` for fields with stable API defaults.** Leads to "sticky last user value" UX — removing the HCL line doesn't restore the default. Use `Optional + Default` instead. *Example: v2.3.9 fix on `wallarm_rule_graphql_detection.introspection` and `active`.*
3. **`Optional + Default + ForceNew` when state can disagree with the schema default.** Import trap. After importing a resource whose API value ≠ schema default, removing the line from HCL plans `state → default` and triggers ForceNew destroy+recreate. Use `Optional + Computed + ForceNew` when ForceNew applies, even at the cost of asymmetric "remove restores default" UX (destroy is more expensive than asymmetry). *Example: v2.3.9 fix on `wallarm_rule_regex.experimental`.*
4. **`int + omitempty` for fields where 0 is a valid value.** Go's `encoding/json` drops literal zero from `int+omitempty`, so the wire payload is missing the field and the API rejects with "can't be blank". Use `*int + omitempty` (and a `GetPointerIfConfigured`-style helper). *Example: wallarm-go v0.12.1 on `ActionCreate.Rate/Burst/Delay/OverlimitTime`.*
5. **Validating field-vs-mode constraints in the API mapper alone.** Silent drops produce perpetual diffs because state and HCL stay misaligned. Validate at plan time via `CustomizeDiff` so the user gets a clean error. *Example: `EnumeratedParamsCustomizeDiff` (v2.3.8).*
6. **Read that doesn't populate every Optional+Computed field.** `ImportStateVerify` fails because state is incomplete. Read must mirror every settable schema attribute. *Example: v2.3.9 fix on `wallarm_rule_regex.experimental` Read.*
7. **`Optional+Computed` on `TypeString` fields users might want to clear.** SDKv2 normalises an explicit empty string in HCL (`field = ""`) to `cty.NullVal`. Computed semantics then preserve state instead of clearing. Result: setting the field to `""` after a non-empty value silently fails. For string fields where "clear" is a valid user intent, prefer **`Optional` alone** (accept the legacy-import-CLI asymmetry; the modern `import{}+-generate-config-out` flow doesn't hit it). *Example: v2.3.9 revert on `set` / `title` after `TestAccRuleParserState_UpdateSetToEmpty` regressed.*

8. **`StringInSlice` validators sourced only from schema Description text.** Schema descriptions in code often lag behind both `docs/resources/<rule>.md` and CLAUDE.md. A validator derived from a single source — especially the schema's inline Description — frequently misses values the API actually accepts (in this codebase: `regex.attack_type=vpatch`, `disable_attack_type` accepts `ssi`/`mail_injection`/`ssti`/`invalid_xml`/`vpatch`). Don't skip the validator — **users want the plan-time enum check**. Source the allowlist from the **union** of: docs/resources/<rule>.md, the schema's old Description text (often has values the docs miss), `references/` cross-references, and any values referenced by acc tests. When in doubt, expand the list. Document where each value came from in a code comment so a future maintainer can audit + extend.

9. **Strict-range validator (`IntBetween(N, M)` with `N > 0`) on a nested-block `TypeInt` sub-field whose API can omit the key.** Because the legacy flat-state model can't carry `cty.NullVal` for nested `TypeInt`, an API-omitted key materialises in state as `0`. `terraform import` + `-generate-config-out` then emits a literal `field = 0` line that the validator rejects at plan time — the round-trip is broken regardless of whether the schema is Optional, Optional+Computed, or Optional+Default. **Fix:** the validator must accept `0` (state-encoding artifact, not real data) alongside the valid range, e.g. `validation.Any(IntInSlice([]int{0}), IntBetween(N, M))`; the API mapper must drop `0` on the wire so it never reaches the API. (Top-level Optional ints are not affected — modern proto state supports null. See the §D nested-block caveat.) *Example: v2.3.9 fix on `reactionSchema.{block_by_session, block_by_ip, graylist_by_ip}` after the `IntBetween(600, 315569520)` validator added in the same release blocked partial-reaction imports.*

## G. Tie-breaker rules of thumb

- **Default vs Computed:** prefer `Default` when the value is stable and you want symmetric remove-restores-default UX. Prefer `Computed` when the API value is dynamic / unknown or when the field is `ForceNew` (where asymmetric stickiness is safer than spurious destroys).
- **Required vs Optional+Default:** prefer `Required` when the value is semantically meaningful enough that you want users to think about it (e.g., a security-relevant flag), even if the API has a default. UX choice, not a hard rule.
- **CustomizeDiff vs ValidateDiagFunc:** `ValidateDiagFunc` for single-field constraints (per-attribute, clearer error message); `CustomizeDiff` for cross-field / conditional logic.

---

## See also

- `sdkv2-patterns` skill — canonical SDKv2 mechanics (schema types, behaviors, validators, CustomizeDiff helpers, ImportState, retries, testing).
- `references/action.md` — provider-internal reference for action conditions and the server-side data model.
- `references/point.md` — point-element chaining tables.
- `references/rules_api_fields.md` — probe-derived API ground truth per rule type (defaults, required fields, range constraints).
