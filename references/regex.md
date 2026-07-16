# Pire regex syntax for Wallarm rules

Reference for writing regular expressions in Wallarm rules. Wallarm uses the
**Pire** engine (Perl Incompatible Regular Expressions): fast (~400 MB/s, fixed
cost per byte) but deliberately limited, so many PCRE / Python `re` / Go
`regexp` constructs do not work.

## 1. Overview

The same Pire engine backs every regex field in the provider. Knowing its
capabilities and escaping rules is what keeps a rule both correct and
compile-friendly. HCL and JSON each consume one level of backslash escaping, so
a literal `.` reaches the engine as `\.` only when written `\\.` in HCL.

## 2. Model

Pire matches against the **whole** input string by default (PCRE_ANCHORED): the
pattern must consume the entire value. Some Wallarm contexts internally wrap
`.*` on each side for substring semantics, but that cannot be relied on - be
explicit with `^`/`$` for a full match or `.*...*` for "contains".

Two engine features have non-PCRE semantics:

- `~a` - **negation / complement**: accepts strings NOT matching `a`. When
  concatenated (`~A B`), the constraint applies only to the matched *prefix*, not
  the whole input (the split-point trap, §6.4 Example 1).
- `a&b` - **intersection**: matches both `a` and `b` (rare).

## 3. Elements

Regex appears in four places, all the same engine:

| Site | Field | Matches |
|---|---|---|
| Condition `type = "regex"` | `action { value }` | the point's value against `value` |
| Custom signature | `wallarm_rule_regex.regex` | detection-point content |
| Enumerated params (regexp mode) | `name_regexps` / `value_regexps` on `wallarm_rule_brute` / `_bola` / `_enum` | parameter names / values |
| Credential stuffing | `login_regex` / `regex` on `wallarm_rule_credential_stuffing_regex` | the login field / the credential value |

`name_regexps` / `value_regexps` are string lists; in regexp mode **both** are
required (each with >=1 element). Use `[""]` to opt out of one filter while
still satisfying the requirement.

## 4. Behavior

### 4.1 Anchoring

- `condition type = "regex"` matches the full point value: wrap with `.*` for
  "contains" or anchor with `^...$` for an explicit full match.
- `wallarm_rule_regex.regex` matches the entire detection-point content.

### 4.2 Escaping across layers

| Layer | `\` | `"` |
|---|---|---|
| Terraform HCL string | `\\` | `\"` |
| JSON over the wire | `\\` | `\"` |
| Pire pattern itself | `\\` (must reach the engine as `\\`) | bare in `[]`; `"` otherwise |

The user types two backslashes either way; HCL and JSON each consume one.
Recount backslashes whenever copying a pattern between HCL and JSON.

### 4.3 Unsupported constructs

- No lookahead / lookbehind (`(?=...)`, `(?!...)`, `(?<=...)`, `(?<!...)`).
- No backreferences (`\1`, ...).
- No usable capture groups (matched substring is not extractable).
- No conditionals, possessive quantifiers, atomic groups.
- No non-greedy quantifiers (`*?`, `+?`, `??`).

If you need any of these, split into narrower rules or switch to `equal` /
`iequal`.

### 4.4 Compile-friendliness

`equal` / `iequal` are cheaper than `regex`: smaller compile contribution and a
different selector branch on the same point. Default to a literal `equal` /
`iequal`; reach for `regex` only when the value set is unbounded or genuinely
pattern-shaped. Mixing types on one point is an antipattern.

| Want to match | Use |
|---|---|
| `/api/v1/users` exactly | `equal` |
| same, case-insensitive | `iequal` |
| `/api/v1/<numeric_id>` | `regex` (`^/api/v1/\d+$`) |
| any of `/login`, `/signin`, `/auth` | one rule per scope with `equal` |

Patterns that will not deploy:

| Shape | Why | Mitigation |
|---|---|---|
| `x.{50}` | bounded `.` repetition explodes states | use `x.+` or split rules |
| `(a\|b\|c\|...){20,}` | combinatorial state explosion | flatten into multiple rules |
| `~` + deep `.*` chains | negation grows enormous | drop `~` or simplify the inner pattern |

## 5. Parameters

The regex-bearing fields are parameters of their resources (`rules_api_fields.md`
for per-hint field ground truth); this doc governs the value syntax, not the
schema.

## 6. Reference data

### 6.1 Cheat sheet

| Construct | Meaning | Example |
|---|---|---|
| `abc` | literal | `admin` |
| `a\|b` | alternation | `(get\|post)` |
| `a*` `a+` `a?` | 0+, 1+, 0-1 | `\d+` |
| `a{n}` `a{n,m}` `a{n,}` | bounded repetition | `\d{4}` |
| `.` | any single character | `a.b` |
| `[abc]` `[a-z]` `[^abc]` | character classes | `[A-Za-z0-9]` |
| `\w \W \d \D \s \S` | predefined classes | `\w+@\w+\.\w+` |
| `^` `$` | start / end anchors | `^login$` |
| `(...)` | grouping (no capture in Pire) | `(foo\|bar)+` |
| `~a` | negation / complement (see §2, §6.4) | Example 1 |
| `a&b` | intersection | rare |

### 6.2 Worked examples

**Example 1 - match `admin`/`root` in path, ignore in query string.**
Recommended (character-class negation): `^[^?]*(admin|root)` - from start,
consume non-`?` chars, then the keyword, so it must be in the path; no `$` since
the trailing query string is irrelevant. The complement form
`^((~(.*[?].*))(admin|root).*)$` matches the same set but hits the split-point
trap: the no-`?` constraint binds only the prefix, and the trailing `.*` can
consume `?q=root`. It fails only when no valid split exists (the keyword is
absent or always right of a `?`). Verified with `cpire-runner`.

**Example 2 - User-Agent families:** `^(curl|wget|python-requests|libwww-perl)/[0-9.]+`

**Example 3 - path traversal in URI** (HCL, note double-escape):

```hcl
resource "wallarm_rule_regex" "ptrav" {
  attack_type = "ptrav"
  regex       = "\\.\\./"
  point       = [["uri"]]
  action {
    type  = "iequal"
    point = { header = "HOST" }
    value = "api.example.com"
  }
}
```

API equivalent: `{"type":"regex","regex":"\\.\\./","point":[["uri"]],"attack_type":"ptrav","action":[{"type":"iequal","point":["header","HOST"],"value":"api.example.com"}]}`.

**Example 4 - SQL-injection probe:** `'\s*or\s*'?\d+'?\s*=\s*'?\d+` (matches
`' or 1=1`, `' or '1'='1`).

**Example 5 - digit-only credential:** on `wallarm_rule_credential_stuffing_regex`,
`login_regex = ".*"`, `regex = "^[0-9]{1,12}$"`, `case_sensitive = true`
(credential matchers operate on the entire field, so anchor both ends).

**Example 6 - contains-any:** `.*(login|signin|auth).*` (a bare alternation is
anchored both ends and matches only exact equals). `(.*login.*)|(.*signin.*)|(.*auth.*)`
is equivalent and can prune branches earlier on long alternation lists.

## 7. References

- `rules-core.md` - Action/Condition/Hint model.
- `point.md` - detection points a `regex` condition targets.
- `rules_api_fields.md` - regex-bearing fields per hint type.
- Wallarm docs: https://docs.wallarm.com/user-guides/rules/rules/#condition-type-regex
