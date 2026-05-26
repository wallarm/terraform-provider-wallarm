# Proton Types — Attack Types & Point Types

Reference extracted from Wallarm's Proton library type definitions. Two stable enumerations the API and provider depend on:

1. **Attack types** — numeric IDs paired with attack-name symbols, grouped by purpose.
2. **Point types** — request-parser elements with classification flags (simple / keys / array / parser / immutable / pollution) and a rating.

Upstream source: `gl.wallarm.com/wallarm-node/meganode/.../proton/types.rb`. Snapshot lives in `.claude/types.rb` (gitignored); re-sync from upstream when the API evolves and update this file.

---

## Attack types

| ID | Name |
|----|------|
| 0  | `warn` |
| 1  | `xss` |
| 2  | `sqli` |
| 3  | `rce` |
| 4  | `xxe` |
| 5  | `ptrav` |
| 6  | `crlf` |
| 7  | `redir` |
| 8  | `nosqli` |
| 9  | `infoleak` |
| 10 | `brute` |
| 11 | `dirbust` |
| 12 | `marker` |
| 13 | `ldapi` |
| 14 | `scanner` |
| 15 | `bot` |
| 16 | `bola` |
| 17 | `mass_assignment` |
| 18 | `apileak` |
| 19 | `ssrf` |
| 20 | `blocked_source` |
| 21 | `api_abuse` |
| 22 | `vectors` |
| 23 | `unused0` |
| 24 | `credential_stuffing` |
| 25 | `ssi` |
| 26 | `mail_injection` |
| 27 | `ssti` |
| 28 | `invalid_xml` |
| 29 | `overlimit_res` |
| 30 | `data_bomb` |
| 31 | `vpatch` |

**API Policy Enforcement group (32–38):**

| ID | Name |
|----|------|
| 32 | `undefined_endpoint` |
| 33 | `undefined_parameter` |
| 34 | `missing_auth` |
| 35 | `missing_parameter` |
| 36 | `invalid_parameter_value` |
| 37 | `invalid_request` |
| 38 | `processing_overlimit` |

**GraphQL parser group (39–45):**

| ID | Name |
|----|------|
| 39 | `gql_depth` |
| 40 | `gql_value_size` |
| 41 | `gql_aliases` |
| 42 | `gql_doc_size` |
| 43 | `gql_docs_per_batch` |
| 44 | `gql_introspection` |
| 45 | `gql_debug` |

**ACL group (46–58):** additional to `21: api_abuse`.

| ID | Name |
|----|------|
| 46 | `account_takeover` |
| 47 | `security_crawlers` |
| 48 | `scraping` |
| 49 | `rate_limit` |
| 50 | `enum` |
| 51 | `file_upload_violation` |
| 52 | `resource_consumption` |
| 53 | `session_anomaly` |
| 54 | `query_anomaly` |
| 55 | `ai_attack` |
| 56 | `mcp_schema_violation` |
| 57 | `mcp_parameter_violation` |
| 58 | `mcp_acl_violation` |

> The provider's `attack_type` validators (`wallarm_rule_disable_attack_type`, `_vpatch`, `_regex`) are curated subsets — many entries above are internal (`warn`, `marker`, `bot`, the API-Policy-Enforcement group, the GraphQL-parser group) and not user-settable.

---

## Point types

Each point element has:
- **ID** — numeric type id (used in low-level structures).
- **Flags** — categorisation:
  - `simple` — non-paired terminal (e.g. `["post"]`).
  - `keys` — paired with a string key (e.g. `["header", "HOST"]`).
  - `array` — paired with an integer index (e.g. `["path", 0]`).
  - `parser` — element drives a parser (consumed by `Proton.valid_parser?`).
  - `immutable` — protected from mutation.
  - `pollution` — eligible for HTTP-parameter-pollution analysis.
- **Rating** — relative priority (higher = checked first).

### Core / structural

| Name | ID | Flags | Rating |
|---|---|---|---|
| `hash` | 3 | keys, pollution | 92 |
| `array` | 4 | array | 90 |
| `path` | 39 | array | 36 |
| `pollution` | 5 | simple | 1 |

### ViewState

| Name | ID | Flags | Rating |
|---|---|---|---|
| `viewstate` | 25 | simple, immutable, parser | 88 |
| `viewstate_sparse_array` | 32 | keys, immutable | 87 |
| `viewstate_array` | 26 | array, immutable | 86 |
| `viewstate_pair` | 27 | array, immutable | 85 |
| `viewstate_triplet` | 28 | array, immutable | 84 |
| `viewstate_dict` | 29 | keys, immutable | 82 |
| `viewstate_dict_key` | 30 | simple, immutable | 81 |
| `viewstate_dict_value` | 31 | simple, immutable | 80 |

### Request line / URI

| Name | ID | Flags | Rating |
|---|---|---|---|
| `action_name` | 37 | simple | 78 |
| `action_ext` | 38 | simple | 74 |
| `uri` | 34 | simple | 72 |
| `route` | 40 | simple | 71 |
| `remote_addr` | 52 | simple | 72 |

### HTTP

| Name | ID | Flags | Rating |
|---|---|---|---|
| `header` | 18 | keys, pollution | 70 |
| `cookie` | 6 | keys, pollution, parser | 30 |
| `get` | 35 | keys, pollution | 44 |
| `post` | 36 | simple | 76 |
| `multipart` | 1 | keys, pollution, parser | 48 |
| `file` | 19 | simple | 46 |
| `content_disp` | 21 | keys | 42 |
| `form_urlencoded` | 7 | keys, pollution, parser | 40 |
| `response_body` | 54 | simple | 76 |
| `response_header` | 53 | keys, pollution | 70 |

### XML

| Name | ID | Flags | Rating |
|---|---|---|---|
| `xml` | 9 | simple, parser | 67 |
| `xml_pi` | 11 | array | 66 |
| `xml_dtd` | 13 | simple | 65 |
| `xml_dtd_entity` | 12 | array | 64 |
| `xml_tag_array` | 15 | array | 63 |
| `xml_tag` | 14 | keys | 62 |
| `xml_attr` | 10 | keys | 61 |
| `xml_comment` | 17 | array | 60 |

### JSON

| Name | ID | Flags | Rating |
|---|---|---|---|
| `json_doc` | 22 | simple, parser | 56 |
| `json_obj` | 23 | keys | 54 |
| `json_array` | 24 | array | 52 |
| `json` | 2 | keys | 50 |

### JWT

| Name | ID | Flags | Rating |
|---|---|---|---|
| `jwt` | 51 | keys, parser | 91 |

### gRPC / Protobuf

| Name | ID | Flags | Rating |
|---|---|---|---|
| `grpc` | 46 | array, parser | 59 |
| `protobuf` | 47 | keys, parser | 58 |
| `protobuf_int32` | 48 | simple | 57 |
| `protobuf_int64` | 49 | simple | 57 |
| `protobuf_varint` | 50 | simple | 57 |

### GraphQL

| Name | ID | Flags | Rating |
|---|---|---|---|
| `gql` | 55 | simple, parser | 89 |
| `gql_query` | 56 | keys | 89 |
| `gql_mutation` | 57 | keys | 89 |
| `gql_subscription` | 58 | keys | 89 |
| `gql_alias` | 59 | simple | 89 |
| `gql_arg` | 60 | simple | 89 |
| `gql_dir` | 61 | keys | 89 |
| `gql_spread` | 62 | keys | 89 |
| `gql_fragment` | 63 | keys | 89 |
| `gql_type` | 64 | keys | 89 |
| `gql_inline` | 65 | simple | 89 |
| `gql_var` | 66 | keys | 89 |

### Encoding parsers

| Name | ID | Flags | Rating |
|---|---|---|---|
| `gzip` | 8 | simple, parser | 38 |
| `percent` | 20 | simple, parser | 32 |
| `base64` | 0 | simple, parser | 28 |
| `htmljs` | 33 | simple, immutable, parser | 28 |
| `hex` | 67 | simple, parser | 28 |

---

## Derived sets (computed in `types.rb`)

These are exposed by `Proton.simple_types`, `Proton.key_types`, etc., and used by validators:

- **`SIMPLE_TYPES`** — all entries with `simple: true`.
- **`KEY_TYPES`** — all entries with `keys: true`.
- **`ARRAY_TYPES`** — all entries with `array: true`.
- **`IMMUTABLE_TYPES`** — all entries with `immutable: true`.
- **`PARSER_TYPES`** — all entries with `parser: true`. Validated by `Proton.valid_parser?`.
- **`TYPES_BY_ID`** — inverse map.
- **`TYPES_BY_RATING`** — names sorted by descending rating.
