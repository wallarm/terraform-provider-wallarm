# Regex guide — Pire syntax for Wallarm rules

Wallarm uses the **Pire** engine (Perl Incompatible Regular Expressions) — fast (~400 MB/s, fixed cost per byte) but deliberately limited. Many PCRE / Python `re` / Go `regexp` constructs don't work.

---

## Where regex appears

Same Pire engine in four places:

1. **Condition `type = "regex"`** — `value` matched against a point's value:
   ```hcl
   action {
     type  = "regex"
     point = { header = "USER-AGENT" }
     value = "(curl|wget|python-requests)/.*"
   }
   ```
2. **`regex` field on `wallarm_rule_regex`** — custom attack signature:
   ```hcl
   resource "wallarm_rule_regex" "block_path_traversal" {
     attack_type = "ptrav"
     regex       = "\\.\\./"
     point       = [["uri"]]
   }
   ```
3. **`name_regexps` / `value_regexps`** on enumerated_parameters (`wallarm_rule_brute`, `_bola`, `_enum`). String lists; `""` means "skip this filter"; at least one of the two lists must be populated in regexp mode.
4. **`login` / `password` regex** on credential-stuffing rules.

---

## Cheat sheet

| Construct | Meaning | Example |
|---|---|---|
| `abc` | literal | `admin` |
| `a\|b` | alternation | `(get\|post)` |
| `a*` `a+` `a?` | 0+, 1+, 0–1 | `\d+` |
| `a{n}` `a{n,m}` `a{n,}` | bounded repetition | `\d{4}` |
| `.` | any single character | `a.b` |
| `[abc]` `[a-z]` `[^abc]` | character classes | `[A-Za-z0-9]` |
| `\w \W \d \D \s \S` | predefined classes | `\w+@\w+\.\w+` |
| `^` `$` | start / end anchors | `^login$` |
| `(...)` | grouping (no capture in Pire) | `(foo\|bar)+` |
| `~a` | **negation / complement** — accepts strings NOT matching `a`. Concatenation has a non-obvious split-point trap (see Example 1 below). | Example 1 |
| `a&b` | intersection — matches both | rare |

### Doesn't work

- No lookahead / lookbehind (`(?=...)`, `(?!...)`, `(?<=...)`, `(?<!...)`)
- No backreferences (`\1`, …)
- No usable capture groups (matched substring not extractable)
- No conditionals, possessive quantifiers, atomic groups
- No non-greedy quantifiers (`*?`, `+?`, `??`) — don't write them

If you need any of these, restructure into multiple narrower rules or switch to `equal` / `iequal`.

---

## Anchoring

By default Pire matches against the **whole** input string (PCRE_ANCHORED) — your pattern must consume the entire value. Some Wallarm contexts internally wrap `.*` on each side for substring semantics, but you can't rely on it.

For Wallarm rules:
- **`condition type = "regex"`** matches the full point value. Wrap with `.*` for "contains", or anchor with `^…$` for explicit full-match.
- **`wallarm_rule_regex.regex`** same — pattern matches entire detection-point content.

When in doubt, be explicit with `^`/`$` or `.*…\.*`.

---

## Practical examples

### Example 1 — match `admin`/`root` in path, ignore in query string

> | Input | Match? | Reason |
> |---|---|---|
> | `/admin` | yes | `admin` in path |
> | `/test/admin` | yes | `admin` in path |
> | `/test/admin?q=root` | yes | `admin` in path before `?` |
> | `/test?q=admin` | no | only in query string |

**Approach A — character class negation (recommended):**

```
^[^?]*(admin|root)
```

"From start, consume zero or more non-`?` chars, then match `admin`|`root`." `[^?]*` stops at `?`, so the keyword must be in the path. No `$` — trailing query string is irrelevant.

| Input | Matches? | How |
|---|---|---|
| `/admin` | yes | `[^?]*` matches `/`, then `admin` |
| `/test/admin` | yes | `[^?]*` matches `/test/`, then `admin` |
| `/test/admin?q=root` | yes | `[^?]*` matches `/test/`, then `admin`; `?q=root` left unconsumed |
| `/test?q=admin` | no | `[^?]*` can't cross `?`; nothing inside ends with the keyword |

**Approach B — using `~` (complement) operator:**

```
^((~(.*[?].*))(admin|root).*)$
```

Same set of matches as A. **The naive reading "string without `?`, then keyword, then anything" is wrong**: when `~` is concatenated with the rest of the pattern, the no-`?` constraint applies only to the matched **prefix**, not the whole input. The engine looks for a split where the prefix matches `~(.*[?].*)` AND the suffix matches `(admin|root).*`. The trailing `.*` is unconstrained — it can consume `?q=root`.

Input fails only if **no valid split** exists — i.e. the keyword either doesn't appear at all, or every occurrence is to the right of a `?`.

| Input | Matches? | Valid split |
|---|---|---|
| `/admin` | yes | `~(...)` matches `/`, then `admin`, then empty |
| `/test/admin` | yes | `~(...)` matches `/test/`, then `admin`, then empty |
| `/test/admin?q=root` | yes | `~(...)` matches `/test/`, then `admin` at position 6, then `.*` consumes `?q=root` |
| `/test?q=admin` | no | every no-`?` prefix ends ≤ position 5; `admin` only at position 8. No valid split. |

Verified with `cpire-runner` on these inputs.

When to prefer B: rarely. A reads more naturally. B is useful for language-level negation that character classes can't express (e.g. "not containing the substring `XXX`"). Beware the split-point trap above: `~A B` does NOT mean "the whole string isn't A AND matches B" — it means "find a split where the prefix is in `~A` and the suffix is in B".

### Example 2 — User-Agent families

```
^(curl|wget|python-requests|libwww-perl)/[0-9.]+
```

### Example 3 — path traversal in URI

```hcl
resource "wallarm_rule_regex" "ptrav" {
  attack_type = "ptrav"
  regex       = "\\.\\./"     # double-escape for HCL
  point       = [["uri"]]
  action {
    type  = "iequal"
    point = { header = "HOST" }
    value = "api.example.com"
  }
}
```

API equivalent:
```json
{
  "type": "regex",
  "regex": "\\.\\./",
  "point": [["uri"]],
  "attack_type": "ptrav",
  "action": [
    {"type": "iequal", "point": ["header", "HOST"], "value": "api.example.com"}
  ]
}
```

Always recount backslashes when copying between HCL and JSON. Both layers consume one level of escape.

### Example 4 — SQL-injection probe

```
'\s*or\s*'?\d+'?\s*=\s*'?\d+
```

Matches `' or 1=1`, `' or '1'='1`. Use on post-body or query-string detection point.

### Example 5 — digit-only password

```hcl
login_regex    = ".*"
password_regex = "^[0-9]{1,12}$"
```

Anchored both ends — credential-stuffing matchers operate on the entire field.

### Example 6 — `.*` wrap trick

`(login|signin|auth)` alone matches **only** strings exactly equal to one of those (anchored both ends). For "URI contains any of those":

```
.*(login|signin|auth).*
```

Or, sometimes faster on large alternations:

```
(.*login.*)|(.*signin.*)|(.*auth.*)
```

First form is cleaner. Second can prune branches earlier on long alternation lists.

---

## Patterns that won't deploy

| Shape | Why | Mitigation |
|---|---|---|
| `x.{50}` | bounded `.` repetition explodes states | use `x.+` or split rules |
| `(a\|b\|c\|...){20,}` | combinatorial state explosion | flatten into multiple rules |
| `~` + deep `.*` chains | negation gets enormous | drop `~` or simplify inner |

---

## Regex vs `equal` / `iequal`

`equal` and `iequal` are cheaper — smaller compile contribution, different selector branch from `regex` on the same point. Mixing types on one point is an antipattern.

**Default:** literal → `equal`/`iequal`. Reach for `regex` only when the value-set is unbounded or genuinely pattern-shaped.

| Want to match | Use |
|---|---|
| `/api/v1/users` exactly | `equal` |
| same, case-insensitive | `iequal` |
| `/api/v1/<numeric_id>` | `regex` (`^/api/v1/\d+$`) |
| Any of `/login`, `/signin`, `/auth` | one rule per scope with `equal` (preserves type-consistency) |

---

## Escaping

| Layer | `\` | `"` |
|---|---|---|
| Terraform HCL string | `\\` | `\"` |
| JSON over the wire | `\\` | `\"` |
| Pire pattern itself | `\\` (must reach engine as `\\`) | bare in `[]`; `"` otherwise |

A literal `.` must reach the engine as `\.`. In HCL: `"\\."`. In curl JSON: `"\\."`. User types two backslashes either way; HCL and JSON each consume one.

---

## See also

- Wallarm docs: https://docs.wallarm.com/user-guides/rules/rules/#condition-type-regex
