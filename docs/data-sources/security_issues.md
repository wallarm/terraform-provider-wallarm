---
layout: "wallarm"
page_title: "Wallarm: wallarm_security_issues"
subcategory: "Common"
description: |-
  Get details on security issues detected by Wallarm.
---

# wallarm_security_issues

Use this data source to get details on security issues detected by Wallarm.

## Example Usage

```hcl
# Looks up for 1000 (max) security issues.

data "wallarm_security_issues" "issues" {
  limit = 1000
}
```

```hcl
# Looks up for security issues with an offset of 1000.
# Issues from 1001 to 2000 will be returned if they exist.

data "wallarm_security_issues" "issues" {
  limit  = 1000
  offset = 1000
}
```

```hcl
# Looks up for all security issues without pagination limits.

data "wallarm_security_issues" "all_issues" {
  unlimited = true
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to retrieve security issues for. If not set, the default client is used.
* `limit` - (optional) number of security issues to return. Possible values: 0–1000. Defaults to `1000`.
* `offset` - (optional) number of security issues to skip before returning results. Defaults to `0`.
* `unlimited` - (optional) boolean flag to retrieve all security issues regardless of pagination limits. Defaults to `false`.

## Attributes Reference

`issues` - list of security issue objects. Each object contains the following attributes:

* `id` - integer security issue ID.
* `client_id` - ID of the client associated with the security issue.
* `severity` - severity level of the security issue (e.g., `low`, `medium`, `high`).
* `state` - current state of the security issue (e.g., `open`, `closed`, `falsepositive`).
* `volume` - number of times the security issue was detected.
* `name` - name of the security issue.
* `created_at` - Unix timestamp when the security issue was created.
* `discovered_at` - Unix timestamp when the security issue was first discovered.
* `discovered_by` - identifier of the entity that discovered the security issue.
* `discovered_by_display_name` - display name of the entity that discovered the security issue.
* `url` - full URL where the security issue was detected.
* `host` - host where the security issue was detected.
* `path` - path where the security issue was detected.
* `parameter_display_name` - display name of the parameter associated with the security issue.
* `parameter_position` - position of the parameter (e.g., `query`, `body`, `header`).
* `parameter_name` - name of the parameter associated with the security issue.
* `http_method` - HTTP method of the request that exposed the security issue.
* `aasm_template` - AASM template identifier for the security issue state machine.
* `verified` - boolean indicating whether the security issue has been verified.
* `mitigations` - mitigation details for the security issue. Contains:
  * `vpatch` - virtual patch mitigation. Contains:
    * `rule_id` - integer ID of the virtual patch rule applied to mitigate the issue.
* `issue_type` - type classification of the security issue. Contains:
  * `id` - string identifier of the issue type.
  * `name` - display name of the issue type.
* `owasp` - list of OWASP categories associated with the security issue. Each entry contains:
  * `id` - string OWASP category ID.
  * `name` - short name of the OWASP category.
  * `full_name` - full name of the OWASP category.
  * `link` - URL to the OWASP category documentation.
* `tags` - list of tags associated with the security issue. Each entry contains:
  * `id` - integer tag ID.
  * `name` - tag name.
  * `slug` - tag slug identifier.
