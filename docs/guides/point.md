---
layout: "wallarm"
page_title: "Wallarm Rule Point"
description: |-
  Provides examples of the point argument.
---

# Point

`point` — (**required**) request parameter to apply the rules to. Additional information is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

## Base points (level 1)

Base points are the top-level request parameters available in the Wallarm Console UI.

| POINT | POSSIBLE VALUES |
|-------|----------------|
| `action_ext`, `action_name`, `get_name`, `header_name`, `path`, `path_all`, `uri` | `base64`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `get`, `get_all` | `array`, `array_all`, `base64`, `gql`, `gzip`, `hash`, `hash_all`, `hash_name`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `header`, `header_all` | `array`, `array_all`, `base64`, `cookie`, `cookie_all`, `cookie_name`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `post` | `base64`, `form_urlencoded`, `form_urlencoded_all`, `form_urlencoded_name`, `gql`, `grpc`, `grpc_all`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `multipart`, `multipart_all`, `multipart_name`, `percent`, `xml` |

## Nested elements (level 2+)

Elements that appear inside a parser chain. The ELEMENT column shows the parser/structure type; POSSIBLE VALUES shows what can follow it at the next level.

| ELEMENT | POSSIBLE VALUES |
|---------|----------------|
| `base64`, `gzip`, `htmljs`, `jwt`, `jwt_all`, `hash_name`, `form_urlencoded_name`, `cookie_name`, `multipart_name` | `base64`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `percent` | `base64`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `json_doc` | `array`, `array_all`, `hash`, `hash_all`, `hash_name` |
| `xml` | `array`, `array_all`, `hash`, `hash_all`, `hash_name`, `xml_comment`, `xml_comment_all`, `xml_dtd`, `xml_dtd_entity`, `xml_dtd_entity_all`, `xml_pi`, `xml_pi_all`, `xml_tag`, `xml_tag_all`, `xml_tag_array`, `xml_tag_array_all`, `xml_tag_name` |
| `form_urlencoded`, `form_urlencoded_all`, `cookie`, `cookie_all`, `hash`, `hash_all` | `array`, `array_all`, `base64`, `gzip`, `hash`, `hash_all`, `hash_name`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `array`, `array_all` | `base64`, `cookie`, `cookie_all`, `cookie_name`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `multipart`, `multipart_all` | `array`, `array_all`, `base64`, `file`, `gzip`, `hash`, `hash_all`, `hash_name`, `header`, `header_all`, `header_name`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `xml` |
| `grpc`, `grpc_all` | `base64`, `gzip`, `htmljs`, `json_doc`, `jwt`, `jwt_all`, `percent`, `protobuf`, `protobuf_all`, `protobuf_name`, `xml` |
| `gql` | `gql_fragment`, `gql_fragment_all`, `gql_mutation`, `gql_mutation_all`, `gql_query`, `gql_query_all`, `gql_subscription`, `gql_subscription_all` |
| `gql_query`, `gql_mutation`, `gql_subscription`, `gql_query_all`, `gql_mutation_all`, `gql_subscription_all` | `array`, `array_all`, `gql_dir`, `gql_dir_all`, `gql_inline`, `gql_spread`, `gql_spread_all`, `gql_var`, `gql_var_all`, `hash`, `hash_all`, `hash_name` |
| `gql_fragment`, `gql_fragment_all` | `array`, `array_all`, `gql_dir`, `gql_dir_all`, `gql_inline`, `gql_spread`, `gql_spread_all`, `gql_type`, `gql_type_all`, `gql_var`, `gql_var_all`, `hash`, `hash_all`, `hash_name` |

~> **Note:** Some elements have context-dependent children. For example, `json_doc` under `post` also allows `gql`; `form_urlencoded` under `post` also allows `gql`. The tables above show the most common children set. Context-specific additions apply when deeper in a `post` body chain.

## Paired vs simple elements

Point elements are either **paired** (take a second value) or **simple** (standalone).

### Paired elements

Format: `["element", "value"]`. The second value is a key name, field name, or integer index:

| Element | Value type | Example |
|---------|-----------|---------|
| `header` | Header name | `["header", "HOST"]` |
| `cookie` | Cookie name | `["cookie", "session_id"]` |
| `get` | Query parameter name | `["get", "search"]` |
| `hash` | Object key name | `["hash", "password"]` |
| `form_urlencoded` | Form field name | `["form_urlencoded", "username"]` |
| `multipart` | Multipart field name | `["multipart", "file"]` |
| `content_disp` | Content-Disposition field | `["content_disp", "filename"]` |
| `response_header` | Response header name | `["response_header", "Set-Cookie"]` |
| `jwt` | JWT section | `["jwt", "payload"]` |
| `json`, `json_obj` | JSON object key | `["json_obj", "data"]` |
| `xml_tag` | Tag name | `["xml_tag", "root"]` |
| `xml_attr` | Attribute name | `["xml_attr", "id"]` |
| `protobuf` | Protobuf field name | `["protobuf", "field"]` |
| `gql_query` | Query/field name | `["gql_query", "getUser"]` |
| `gql_mutation` | Mutation name | `["gql_mutation", "createUser"]` |
| `gql_subscription` | Subscription name | `["gql_subscription", "onUpdate"]` |
| `gql_fragment` | Fragment name | `["gql_fragment", "UserFields"]` |
| `gql_dir` | Directive name | `["gql_dir", "deprecated"]` |
| `gql_spread` | Spread target | `["gql_spread", "UserFields"]` |
| `gql_type` | Type name | `["gql_type", "User"]` |
| `gql_var` | Variable name | `["gql_var", "userId"]` |
| `path` | Path segment index (integer) | `["path", 0]` |
| `array`, `json_array` | Array index (integer) | `["array", 0]` |
| `grpc` | gRPC field number (integer) | `["grpc", 1]` |
| `xml_pi` | Processing instruction index | `["xml_pi", 0]` |
| `xml_dtd_entity` | DTD entity index | `["xml_dtd_entity", 0]` |
| `xml_tag_array` | Tag array index | `["xml_tag_array", 0]` |
| `xml_comment` | Comment index | `["xml_comment", 0]` |
| `viewstate_array`, `viewstate_pair`, `viewstate_triplet` | Index (integer) | `["viewstate_array", 0]` |
| `viewstate_dict`, `viewstate_sparse_array` | Key name | `["viewstate_dict", "key"]` |

### Simple elements

Format: `["element"]`. No second value:

`post`, `json_doc`, `xml`, `uri`, `action_name`, `action_ext`, `route`, `remote_addr`, `response_body`, `file`, `base64`, `gzip`, `htmljs`, `percent`, `pollution`, `gql`, `gql_alias`, `gql_arg`, `gql_inline`, `viewstate`, `viewstate_dict_key`, `viewstate_dict_value`, `protobuf_int32`, `protobuf_int64`, `protobuf_varint`, `xml_dtd`

## Examples

### 1. Form POST (`application/x-www-form-urlencoded`)

```
p1=1&p2[a]=2&p2[b]=3&p3[]=4&p3[]=5&p4=6&p4=7
```

| Point | Matches |
|-------|---------|
| `point = [["post"], ["form_urlencoded", "p1"]]` | `1` |
| `point = [["post"], ["form_urlencoded", "p2"], ["hash", "a"]]` | `2` |
| `point = [["post"], ["form_urlencoded", "p2"], ["hash", "b"]]` | `3` |
| `point = [["post"], ["form_urlencoded", "p3"], ["array", 0]]` | `4` |
| `point = [["post"], ["form_urlencoded", "p3"], ["array", 1]]` | `5` |
| `point = [["post"], ["form_urlencoded", "p4"], ["array", 0]]` | `6` |
| `point = [["post"], ["form_urlencoded", "p4"], ["array", 1]]` | `7` |
| `point = [["post"], ["form_urlencoded", "p4"], ["pollution"]]` | `6,7` |

### 2. JSON body

```json
{
  "p1": "value",
  "p2": ["v1", "v2"],
  "p3": { "somekey": "somevalue" }
}
```

| Point | Matches |
|-------|---------|
| `point = [["post"], ["json_doc"], ["hash", "p1"]]` | `value` |
| `point = [["post"], ["json_doc"], ["hash", "p2"], ["array", 0]]` | `v1` |
| `point = [["post"], ["json_doc"], ["hash", "p2"], ["array", 1]]` | `v2` |
| `point = [["post"], ["json_doc"], ["hash", "p3"], ["hash", "somekey"]]` | `somevalue` |

### 3. GET parameters

`/?q=some+text&check=yes`

| Point | Matches |
|-------|---------|
| `point = [["get", "q"]]` | `some text` |
| `point = [["get", "check"]]` | `yes` |

### 4. URL path

`/blogs/123/index.php?q=aaa`

| Point | Matches |
|-------|---------|
| `point = [["uri"]]` | `/blogs/123/index.php?q=aaa` |
| `point = [["path", 0]]` | `blogs` |
| `point = [["path", 1]]` | `123` |
| `point = [["action_name"]]` | `index` |
| `point = [["action_ext"]]` | `php` |
| `point = [["get", "q"]]` | `aaa` |

### 5. Headers

```
GET / HTTP/1.1
Host: example.com
```

| Point | Matches |
|-------|---------|
| `point = [["header", "HOST"]]` | `example.com` |

### 6. Cookies

```
GET / HTTP/1.1
Cookie: session=abc123; theme=dark
```

| Point | Matches |
|-------|---------|
| `point = [["header", "COOKIE"], ["cookie", "session"]]` | `abc123` |
| `point = [["header", "COOKIE"], ["cookie", "theme"]]` | `dark` |

### 7. JWT

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2xvY2FsaG9zdCIsImlzcyI6Imh0dHBzOi8vbG9jYWxob3N0In0.signature
```

| Point | Matches |
|-------|---------|
| `point = [["header", "AUTHORIZATION"], ["jwt", "payload"], ["base64"], ["json_doc"], ["hash", "aud"]]` | `https://localhost` |

### 8. XML body

```xml
<root id="1"><item>value</item></root>
```

| Point | Matches |
|-------|---------|
| `point = [["post"], ["xml"], ["xml_tag", "root"], ["xml_attr", "id"]]` | `1` |
| `point = [["post"], ["xml"], ["xml_tag", "root"], ["xml_tag", "item"]]` | `value` |

### 9. gRPC / Protobuf

| Point | Description |
|-------|-------------|
| `point = [["post"], ["grpc", 1], ["protobuf", "name"]]` | Protobuf field `name` in gRPC field 1 |

### 10. GraphQL

```graphql
query getUser($id: ID!) {
  user(id: $id) { name email }
}
```

| Point | Description |
|-------|-------------|
| `point = [["post"], ["gql"], ["gql_query", "getUser"]]` | The `getUser` query operation |
| `point = [["post"], ["gql"], ["gql_query", "getUser"], ["hash", "user"], ["hash", "name"]]` | The `name` field in `user` selection |
| `point = [["post"], ["gql"], ["gql_query", "getUser"], ["gql_var", "id"]]` | The `$id` variable |

### 11. Multipart form upload

```
POST /upload HTTP/1.1
Content-Type: multipart/form-data; boundary=----boundary

------boundary
Content-Disposition: form-data; name="description"

My file
------boundary
Content-Disposition: form-data; name="file"; filename="report.pdf"
Content-Type: application/pdf

<binary data>
------boundary--
```

| Point | Matches |
|-------|---------|
| `point = [["post"], ["multipart", "description"]]` | `My file` |
| `point = [["post"], ["multipart", "file"], ["file"]]` | Binary content of `report.pdf` |
| `point = [["post"], ["multipart", "file"], ["header", "CONTENT-TYPE"]]` | `application/pdf` |
| `point = [["post"], ["multipart", "file"], ["header", "CONTENT-DISPOSITION"], ["content_disp", "filename"]]` | `report.pdf` |

### 12. ViewState (ASP.NET)

ViewState is a base64-encoded structure found in form fields. The parser chain requires `base64` before `viewstate`:

| Point | Description |
|-------|-------------|
| `point = [["post"], ["form_urlencoded", "__VIEWSTATE"], ["base64"], ["viewstate"]]` | Decoded ViewState root |
| `point = [["post"], ["form_urlencoded", "__VIEWSTATE"], ["base64"], ["viewstate"], ["viewstate_pair", 0]]` | First pair in ViewState |
| `point = [["post"], ["form_urlencoded", "__VIEWSTATE"], ["base64"], ["viewstate"], ["viewstate_dict", "key"]]` | Dictionary entry by key |
| `point = [["post"], ["form_urlencoded", "__VIEWSTATE"], ["base64"], ["viewstate"], ["viewstate_array", 0]]` | First array element |

More details on how it works are available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/).
