---
layout: "wallarm"
page_title: "Wallarm: action"
description: |-
  Describes the action argument shared across Wallarm rule resources.
---

# action

The `action` argument defines the scope (conditions) for where a rule applies. There are two ways to define action conditions:

## Option 1: Scope fields (recommended)

Use `action_path`, `action_domain`, and other scope fields to define conditions. The provider automatically expands them into the correct action conditions.

```hcl
resource "wallarm_rule_mode" "example" {
  mode          = "block"
  action_path   = "/api/v1/admin/*"
  action_domain = "example.com"
  action_method = "POST"
  action_scheme = "https"
}
```

Available scope fields:

| Field | Description | Example |
|-------|-------------|---------|
| `action_path` | URL path with wildcard support (`*`, `**`) | `"/api/v1/users"`, `"/admin/*"`, `"/api/**/data"` |
| `action_domain` | Domain (HOST header, case-insensitive match) | `"example.com"` |
| `action_instance` | Application instance (pool) ID | `"17"` |
| `action_method` | HTTP method | `"POST"` |
| `action_scheme` | URL scheme | `"https"` |
| `action_proto` | HTTP protocol version | `"1.1"` |
| `action_query` | Query parameter conditions (block) | See below |
| `action_header` | Custom header conditions (block) | See below |

Query and header blocks:

```hcl
  action_query {
    key   = "token"
    value = "secret"
    type  = "equal"  # equal (default), iequal, regex, absent
  }

  action_header {
    name  = "X-Custom"
    value = "test"
    type  = "equal"
  }
```

Path wildcards:
- `*` — matches any single path segment: `/api/*/users`
- `**` — matches any depth (last directory only): `/api/**/admin`
- `*` as filename — matches any endpoint: `/articles/*`
- `*.*` — matches any file with any extension: `/**/*.*`

## Option 2: Explicit action blocks

Use `action {}` blocks to define conditions directly. This gives full control over condition types and values.

```hcl
resource "wallarm_rule_mode" "example" {
  mode = "block"

  action {
    type  = "iequal"
    value = "example.com"
    point = { header = "HOST" }
  }

  action {
    type  = "equal"
    value = "api"
    point = { path = "0" }
  }

  action {
    type  = "absent"
    point = { action_ext = "" }
  }
}
```

~> **Note:** You cannot mix scope fields and explicit action blocks in the same resource. Use one or the other.

### Action block arguments

* `type` - (**required**) condition type. Must be one of:
  - `equal` — exact match (case-sensitive)
  - `iequal` — case-insensitive match. **Values are automatically lowercased by the API.** Always used for HOST header.
  - `regex` — regular expression match (Pire engine syntax)
  - `absent` — the parameter must not exist
  - `""` (empty string) — used only for `instance` condition

* `value` - (conditionally required) value to match:
  - **Required** for `header` and `query` conditions — the actual matched value (e.g., domain name, query param value).
  - **Empty string `""`** for point-value types (`action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri`, `instance`) — the value goes in `point` instead.
  - **Omittable** for `absent` conditions.
  - For `iequal` type, the API automatically lowercases this value.

* `point` - (**required**) request parameter that triggers the condition:

  | POINT | VALUE TYPE | EXAMPLE |
  |-------|-----------|---------|
  | `header` | Header name (uppercase) | `header = "HOST"` |
  | `method` | HTTP method | `method = "POST"` |
  | `path` | Segment index (string) | `path = "0"` |
  | `action_name` | Endpoint name | `action_name = "login"` |
  | `action_ext` | File extension | `action_ext = "php"` |
  | `query` | Query parameter name | `query = "user"` |
  | `proto` | Protocol version | `proto = "1.1"` |
  | `scheme` | URL scheme | `scheme = "https"` |
  | `uri` | Full URI | `uri = "/api/login"` |
  | `instance` | Application ID | `instance = "42"` |

### Special cases

**Absent conditions:** When `type` is `absent`, the `point` value should be `""` (empty string):

```hcl
  action {
    type  = "absent"
    point = { action_ext = "" }
  }

  action {
    type  = "absent"
    point = { path = "0" }
  }
```

**Value placement:** For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri`, and `instance`, the actual value goes in the `point` map and `value` is set to `""`. For `instance`, `type` must also be `""`:

```hcl
  # action_name = "login" → value goes in point
  action {
    type  = "equal"
    value = ""
    point = { action_name = "login" }
  }

  # instance = "42" → value goes in point, type must be ""
  action {
    type  = ""
    value = ""
    point = { instance = "42" }
  }

  # header = "HOST" → value is the domain, point is the header name
  action {
    type  = "iequal"
    value = "example.com"
    point = { header = "HOST" }
  }
```

## Action reuse

The Wallarm API reuses Actions (scope definitions). If two rules have identical conditions, they share the same Action. This means:
- Creating a rule with the same conditions as an existing one will not create a duplicate Action.
- The provider checks for existing Actions before creating (the `existsAction` check).
- If the Action already exists, the provider returns an error suggesting to import the existing resource.

## Common fields on all rule resources

In addition to `action` and resource-specific fields, all rule resources support these common arguments:

* `client_id` - (Optional) Client ID. Defaults to the provider's client ID.
* `comment` - (Optional) Comment stored with the rule. Default: `"Managed by Terraform"`. On import, the provider sets this default, which may trigger an update if the existing rule has a different comment.
* `variativity_disabled` - (Optional) Disable variativity for the rule. Default: `true`. On import, the provider sets this default, which may trigger an update if the existing rule has variativity enabled.

Common computed attributes:

* `rule_id` - ID of the created rule.
* `action_id` - The action ID (scope conditions).
* `rule_type` - Type of the created rule.
* `mitigation` - Type of the created mitigation (for mitigation controls only).
