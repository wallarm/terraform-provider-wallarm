# Rules Engine — YAML Config Examples

## Simple mode rule

```yaml
name: block_admin
resource_type: wallarm_rule_mode
comment: "Block admin panel"
path: "/api/v1/admin/*"
domain: "example.com"
method: "POST"
mode: "block"
```

## Masking with detection point

```yaml
name: mask_passwords
resource_type: wallarm_rule_masking
comment: "Mask password fields in POST body"
path: "/api/v1/auth/login"
domain: "example.com"
point:
  - ["post"]
  - ["json_doc"]
  - ["hash", "password"]
```

## Grouped stamps (one config = 3 rules)

```yaml
name: disable_sqli_login
resource_type: wallarm_rule_disable_stamp
comment: "FP from request abc123"
path: "/auth/login"
domain: "example.com"
stamps:
  - 1001
  - 1002
  - 1003
point:
  - ["post"]
  - ["form_urlencoded"]
  - ["hash", "query"]
```

## Grouped attack types (one config = 2 rules)

```yaml
name: disable_xss_sqli_health
resource_type: wallarm_rule_disable_attack_type
comment: "Health endpoint FP"
path: "/api/health"
domain: "example.com"
attack_types:
  - sqli
  - xss
point:
  - ["post"]
  - ["form_urlencoded"]
```

## Virtual patch with attack types

```yaml
name: vpatch_api
resource_type: wallarm_rule_vpatch
path: "/api/v1/users"
domain: "example.com"
attack_types:
  - sqli
  - xxe
point:
  - ["uri"]
```

## Multiple file types (one config = 3 rules)

```yaml
name: allow_uploads
resource_type: wallarm_rule_uploads
path: "/api/documents"
domain: "example.com"
file_types:
  - docs
  - images
  - music
point:
  - ["post"]
```

## Disable multiple parsers (one config = 3 rules)

```yaml
name: disable_parsers
resource_type: wallarm_rule_parser_state
path: "/api/raw"
domain: "example.com"
parsers:
  - json_doc
  - xml
  - base64
point:
  - ["post"]
```

## Wildcards in path

```yaml
# * matches any single segment
name: any_version_users
resource_type: wallarm_rule_mode
path: "/api/*/users"
domain: "example.com"
mode: "monitoring"

# ** allows any depth (last directory only)
name: deep_admin
resource_type: wallarm_rule_mode
path: "/api/**/admin"
domain: "example.com"
mode: "block"

# *.ext — any filename with specific extension
name: block_json
resource_type: wallarm_rule_mode
path: "/api/v1/*.json"
domain: "example.com"
mode: "block"
```

## Headers and query parameters

```yaml
name: block_with_conditions
resource_type: wallarm_rule_mode
path: "/api/v1/action"
domain: "super-example.com"
scheme: "https"
method: "POST"
query:
  - key: "key1"
    value: "value_1"
  - key: "key2"
    value: "value_2"
headers:
  - name: "X-Custom"
    value: "test"
  - name: "Content-Type"
    value: "application/json"
    type: "iequal"
mode: "block"
```

## Brute force protection

```yaml
name: brute_login
resource_type: wallarm_rule_brute
path: "/auth/login"
domain: "example.com"
mode: "block"
threshold:
  period: 300
  count: 10
reaction:
  block_by_ip: 600
enumerated_parameters:
  mode: "regexp"
  name_regexps:
    - "^password$"
  value_regexps:
    - ""
  additional_parameters: false
  plain_parameters: false
```

## Generated FP rule with metadata

```yaml
name: a1b2c3d4_e5f6g7h8_disable_stamp
resource_type: wallarm_rule_disable_stamp
comment: "FP from request abc12345-def6-7890-abcd-ef1234567890"
path: "/api/v1/auth/login"
domain: "example.com"
instance: "1"
stamps:
  - 1001
  - 1002
point:
  - ["post"]
  - ["form_urlencoded"]
  - ["hash", "query"]
metadata:
  source: "hits"
  request_id: "abc12345-def6-7890-abcd-ef1234567890"
  point_hash: "e5f6g7h8..."
  hit_ids: [456, 789]
```
