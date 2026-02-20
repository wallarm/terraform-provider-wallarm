---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_false_positive"
subcategory: "Rule"
description: |-
  Provides the "False positive suppression" rule resource.
---

# wallarm_rule_false_positive

Provides a resource to suppress false-positive detections using Wallarm's stamp-based mechanism. Given a `request_id` observed in the Wallarm Console, the resource:

1. Looks up the hit via the Wallarm API.
2. Extracts the detected attack stamps and request point(s).
3. Creates one `disable_stamp` rule per unique **(point × stamp)** combination, scoped to the same domain and URI path as the original request.

This is a fully Terraform-managed alternative to clicking "Mark as false positive" in the Console, and produces rules that are visible under **Rules → Disable specific attack stamp** in the UI.

**Important:** This resource requires an API token with **Global Admin (Extended)** permissions. Standard Admin tokens do not have permission to create `disable_stamp` rules.

**Supported attack types:** `sqli`, `nosqli`, `rce`, `ssi`, `ssti`, `ldapi`, `mail_injection`, `ssrf`, `ptrav`, `xxe`, `scanner`, `xss`, `redir`, `crlf`. Requests with other detection types (e.g. `blocked_source`, `bot`) will produce an error.

**Important:** Rules made with Terraform can't be altered by other rules that usually change how rules work (middleware, variative_values, variative_by_regex).
This is because Terraform is designed to keep its configurations stable and not meant to be modified from outside its environment.

## Example Usage

```hcl
# Suppress a single false-positive XSS detection
resource "wallarm_rule_false_positive" "xss_fp" {
  request_id = "0c34fdea81f746a7ab5d765a4fc4fa3e"
  days_back  = 30
}

output "created_rules" {
  value = wallarm_rule_false_positive.xss_fp.created_rules
}
```

```hcl
# Multi-tenant: specify the client explicitly
resource "wallarm_rule_false_positive" "sqli_fp" {
  client_id  = 12345
  request_id = "690b88b3d6664ecb2281b34452567a46"
  days_back  = 100
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][1]. Defaults to the client associated with the API token.
* `request_id` - (**required**, ForceNew) The request ID (hit ID) as shown in the Wallarm Console. This is the 32-character hexadecimal identifier visible in the Events view. Changing this value destroys and recreates the resource.
* `days_back` - (optional, ForceNew) How many days back to search for the hit. Defaults to `89`. Must be between `1` and `365`. Increase this value for older events. Changing this value destroys and recreates the resource.

## Attributes Reference

* `attack_id` - The Wallarm internal attack ID that the hit belongs to.
* `attack_type` - The detected attack type (e.g. `xss`, `sqli`).
* `domain` - The HTTP `Host` header value of the original request (used to scope the created rules).
* `path` - The URI path of the original request (used to scope the created rules).
* `created_rules` - List of `disable_stamp` rules created by this resource. Each element contains:
  * `rule_id` - ID of the created hint rule.
  * `action_id` - ID of the action (matching conditions) shared by rules with the same scope.
  * `stamp` - The attack stamp value that is suppressed.
  * `point` - JSON-encoded request point where suppression applies (e.g. `["get","myquery"]`).

## Notes on Multiple Rules

A single HTTP request may trigger detection at multiple request points (e.g. two different POST body parameters), or a single point may match multiple stamps. The resource creates one `disable_stamp` rule for every unique **(point, stamp)** pair found in the hit, so `length(created_rules)` may be greater than 1. All rules share the same `action_id` when their domain and path scope is identical.

[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
