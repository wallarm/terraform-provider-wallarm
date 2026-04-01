---
layout: "wallarm"
page_title: "Wallarm: wallarm_ip_lists"
subcategory: "IP Lists"
description: |-
  Reads existing IP list entries from the Wallarm API.
---

# wallarm_ip_lists

Reads all active IP list entries for a given list type (allowlist, denylist, or graylist) from the Wallarm API. Expired entries are automatically filtered out.

Used by the import module to discover existing IP list entries and generate Terraform import blocks.

## Example Usage

```hcl
data "wallarm_ip_lists" "deny" {
  list_type = "denylist"
}

output "denylist_entries" {
  value = data.wallarm_ip_lists.deny.entries
}
```

### Query All Three List Types

```hcl
data "wallarm_ip_lists" "deny" {
  list_type = "denylist"
}

data "wallarm_ip_lists" "allow" {
  list_type = "allowlist"
}

data "wallarm_ip_lists" "gray" {
  list_type = "graylist"
}
```

## Argument Reference

* `client_id` - (Optional) ID of the client to query. Defaults to the provider's default client ID.
* `list_type` - (Required) IP list type. Must be one of: `"allowlist"`, `"denylist"`, `"graylist"`.

## Attributes Reference

* `entries` - List of IP list group objects, each containing:
  * `id` - (Int) API group ID. Used for import and delete operations.
  * `rule_type` - (String) Entry type: `subnet`, `location` (country), `datacenter`, or `proxy_type`.
  * `values` - (List of String) Values in this group. For subnets: single IP with CIDR (e.g., `1.2.3.4/32`). For countries: list of country codes. For datacenters/proxy_type: list of identifiers.
  * `reason` - (String) Description/reason for the entry.
  * `expired_at` - (Int) Expiration unix timestamp.
  * `created_at` - (Int) Creation unix timestamp.
  * `application_ids` - (List of Int) Application IDs this entry applies to. `[0]` means all applications.
  * `status` - (String) Entry status (e.g., `active`).

~> **Note:** The API groups entries differently by type. Countries, datacenters, and proxy types are returned as one group with multiple values. Subnets are returned as one group per IP address.
