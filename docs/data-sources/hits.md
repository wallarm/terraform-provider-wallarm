---
layout: "wallarm"
page_title: "Wallarm: wallarm_hits"
subcategory: "Common"
description: |-
  Fetches attack hits from the Wallarm API for a given request ID.
---

# wallarm_hits

Fetches attack hits from the Wallarm API and provides them in an aggregated, rule-ready structure. Used to build automatic false-positive suppression rules.

~> **Important:** Hits are **ephemeral** — they have a retention period and can be dropped from the API at any time. This data source should only be used for the initial fetch. Rules derived from hits must be cached in Terraform state (e.g., via `terraform_data` with `ignore_changes`) to survive after the source hits expire. Re-fetching on every plan will cause rules to be destroyed when hits are no longer available. See the [Hits to Rules Guide](../guides/hits_to_rules) for the recommended caching pattern.

A single HTTP request can generate multiple hits when different attack vectors are detected in different parts of the request. All hits for a request share the same action conditions (Host header + URI path).

## Example Usage

### Request Mode (default)

Fetch hits for a specific request:

```hcl
data "wallarm_hits" "example" {
  request_id = "d4184a2f138b73c7ef4f7090deb5dfe1"
}
```

### Attack Mode

Expand to all related hits in the same attack campaigns:

```hcl
data "wallarm_hits" "example" {
  request_id = "d4184a2f138b73c7ef4f7090deb5dfe1"
  mode       = "attack"
}
```

### With Custom Time Range and Attack Types

```hcl
data "wallarm_hits" "example" {
  request_id   = "d4184a2f138b73c7ef4f7090deb5dfe1"
  mode         = "attack"
  time         = [1672531200, 1704067199]
  attack_types = ["sqli", "xss", "rce"]
}
```

### Only Generate Stamp Rules for SQLi Hits

```hcl
data "wallarm_hits" "sqli_stamps" {
  request_id   = "d4184a2f138b73c7ef4f7090deb5dfe1"
  attack_types = ["sqli"]
  rule_types   = ["disable_stamp"]
}
```

## Argument Reference

* `client_id` - (Optional) ID of the client to query. Defaults to the provider's default client ID.
* `request_id` - (Required) The unique request identifier to fetch hits for.
* `mode` - (Optional) Fetch mode. `"request"` (default) fetches hits for the request_id only. `"attack"` expands to all related hits sharing the same `attack_id`, filtered by allowed attack types and matching action (Host + path).
* `attack_types` - (Optional) Allowed attack types for filtering. In attack mode, controls which types to fetch from the API. In all modes, only hits matching these types produce rules. Defaults to: `xss`, `sqli`, `rce`, `xxe`, `ptrav`, `crlf`, `redir`, `nosqli`, `ldapi`, `scanner`, `mass_assignment`, `ssrf`, `ssi`, `mail_injection`, `ssti`.
* `rule_types` - (Optional) Rule types to generate. Valid values: `disable_stamp`, `disable_attack_type`. Defaults to both.
* `time` - (Optional) Time range as `[from, to]` unix timestamps. Defaults to 6 months ago to now.
* `include_instance` - (Optional) Include instance (pool ID) in action conditions. When `true` (default), rules are scoped to the hit's application instance. Set to `false` if your Wallarm account is configured to exclude instance from action conditions — otherwise action hash mismatches will occur.

## Attributes Reference

* `action` - Rule action conditions derived from the hit's domain, path, and pool ID. Uses the same schema as `wallarm_rule_*` action blocks, so the output can be passed directly to rule resources.
* `action_hash` - SHA256 hash of the sorted action conditions, used for grouping rules with the same scope.
* `aggregated` - JSON-encoded compact representation of the grouped hits data. Structure: `{action_hash (16 hex chars), action (conditions list), groups (list)}`. Two kinds of groups: stamp groups (keyed by `point_hash`, containing stamps for `disable_stamp` rules) and attack_type groups (keyed by `point_hash_attack_type`, for `disable_attack_type` rules). Use this for caching in `terraform_data` with `ignore_changes`. See the [Hits to Rules Guide](../guides/hits_to_rules) for the recommended caching pattern.
* `hits` - List of hit objects, each containing:
  * `id` - Hit ID components.
  * `type` - Attack type (e.g., `sqli`, `xss`, `rce`).
  * `ip` - Source IP address.
  * `statuscode` - HTTP response status code.
  * `time` - Detection timestamp (unix).
  * `value` - Hit value / payload.
  * `stamps` - Detection stamp values.
  * `stamps_hash` - Hash of stamps.
  * `point` - Detection point as a flat string list.
  * `point_wrapped` - Detection point in 2D nested list structure (matches rule point format).
  * `poolid` - Application pool ID.
  * `attack_id` - Attack ID components.
  * `block_status` - Whether the request was blocked.
  * `request_id` - Request ID.
  * `domain` - Request domain (Host header).
  * `path` - Request URI path.
  * `protocol` - Request protocol.
  * `known_attack` - Known attack type classifications.
  * `node_uuid` - Wallarm node UUID(s) that detected the hit.
