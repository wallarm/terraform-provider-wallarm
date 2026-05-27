# Detection Point Structure — Internal Reference

This file is the authoritative reference for `point` field chaining (paired vs simple elements, base points, context-specific children) used in Wallarm rules.

The Go-side authoritative reference is the `WrapPointElements()` function in `wallarm/common/resourcerule/action_expand.go`.

**Data sources:**
- `spec/point_map.json` — full chaining data (fetched by `scripts/fetch_point_refs.py`; refresh every 30 days).
- `references/proton-types.md` — Proton type IDs, simple/keys/array/parser flags, attack type IDs (verify against upstream Proton source every 30 days).

The `point` field in HCL is a list of lists of strings representing a path through the request parser chain.

---

## 1. Base points (level 1)

| Base point(s) | Allowed children |
|---------------|-----------------|
| `action_ext`, `action_name`, `get_name`, `header_name`, `path`, `path_all`, `uri` | `base64`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `get`, `get_all` | `array`, `array_all`, `base64`, `gql`, `gzip`, `hash`, `hash_all`, `hash_name`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `header`, `header_all` | `array`, `array_all`, `base64`, `cookie`, `cookie_all`, `cookie_name`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `post` | `base64`, `form_urlencoded`, `form_urlencoded_all`, `form_urlencoded_name`, `gql`, `grpc`, `grpc_all`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `multipart`, `multipart_all`, `multipart_name`, `percent`, `xml` |

---

## 2. Paired elements (2-part: `["element", "value"]`)

| Element | Value type | Example |
|---------|-----------|---------|
| `header`, `cookie`, `get`, `hash`, `form_urlencoded`, `multipart`, `content_disp`, `response_header` | String (key/field name) | `["header", "HOST"]` |
| `jwt`, `json`, `json_obj`, `xml_tag`, `xml_attr`, `protobuf` | String (key/field name) | `["jwt", "payload"]` |
| `gql_query`, `gql_mutation`, `gql_subscription`, `gql_fragment`, `gql_dir`, `gql_spread`, `gql_type`, `gql_var` | String (operation/field name) | `["gql_query", "getUser"]` |
| `viewstate_dict`, `viewstate_sparse_array` | String (key name) | `["viewstate_dict", "key"]` |
| `path`, `array`, `json_array`, `grpc` | Integer (index) | `["path", 0]`, `["grpc", 1]` |
| `xml_pi`, `xml_dtd_entity`, `xml_tag_array`, `xml_comment` | Integer (index) | `["xml_pi", 0]` |
| `viewstate_array`, `viewstate_pair`, `viewstate_triplet` | Integer (index) | `["viewstate_array", 0]` |

---

## 3. Simple elements (1-part: `["element"]`)

`post`, `json_doc`, `xml`, `uri`, `action_name`, `action_ext`, `route`, `remote_addr`, `response_body`, `file`, `base64`, `gzip`, `htmljs`, `percent`, `pollution`, `gql`, `gql_alias`, `gql_arg`, `gql_inline`, `viewstate`, `viewstate_dict_key`, `viewstate_dict_value`, `protobuf_int32`, `protobuf_int64`, `protobuf_varint`, `xml_dtd`

---

## 4. Context-specific children

| Element | Context required | Example chain |
|---------|-----------------|---------------|
| `cookie`, `cookie_all`, `cookie_name` | Under `header` or `header_all` | `[["header", "COOKIE"], ["cookie", "session"]]` |
| `form_urlencoded`, `multipart`, `grpc`, `gql` | Under `post` | `[["post"], ["form_urlencoded", "field"]]` |
| `gql` in `json_doc` | Under `post > json_doc` | `[["post"], ["json_doc"], ["gql"]]` |
| `gql` in `percent` | Under `post > form_urlencoded > percent` or `get > percent` | `[["get", "q"], ["percent"], ["gql"]]` |
| `protobuf`, `protobuf_all`, `protobuf_name` | Under `grpc` (which is under `post`) | `[["post"], ["grpc", 1], ["protobuf", "field"]]` |
| `viewstate` and sub-elements | Under `base64` after a parser context | `[["post"], ["form_urlencoded", "f"], ["base64"], ["viewstate"]]` |
| `file`, `header` (nested) | Under `multipart` | `[["post"], ["multipart", "upload"], ["file"]]` |
| `content_disp` | Under `multipart > header` | `[["post"], ["multipart", "f"], ["header", "Content-Disposition"], ["content_disp", "filename"]]` |
| Post-context parsers in `gzip` | Under `post > gzip` | `post > gzip > json_doc` adds `form_urlencoded`, `grpc`, `multipart`, `gql` as children |
