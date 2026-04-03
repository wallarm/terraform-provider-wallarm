---
layout: "wallarm"
page_title: "Wallarm: wallarm_hits_data_cache"
subcategory: "Common"
description: |-
  Deduplicated cache for hits-derived rule data, keyed by action_hash.
---

# wallarm_hits_data_cache

Stores deduplicated rule data from `data.wallarm_hits` in Terraform state, keyed by action_hash. Multiple request IDs sharing the same action (same host and path) produce a single cache entry -- stamps are merged and new point groups are added.

This is a **state-only** resource -- it makes no API calls. Rule data is passed in via `new_entries` and persisted in the `cache` output for downstream rule expansion.

## Example Usage

```hcl
resource "wallarm_hits_data_cache" "this" {
  request_ids = keys(var.request_ids)

  new_entries = {
    for id in local._new_request_ids :
    id => data.wallarm_hits.new[id].aggregated
  }
}

locals {
  _cache = try(jsondecode(wallarm_hits_data_cache.this.cache), {})
}
```

For complete usage including gating, rule expansion, and cleanup, see the [Hits to Rules Guide](../guides/hits_to_rules).

## Argument Reference

* `client_id` - (Optional) Client ID. Defaults to the provider's client ID.
* `request_ids` - (Required) Set of all active request IDs. Used for cleanup -- cache entries with no remaining references are removed.
* `new_entries` - (Optional) Map of `request_id` to aggregated JSON string from `data.wallarm_hits`. New entries are merged into the cache with deduplication by action_hash and detection point. Consumed during apply, not stored in state.

## Attributes Reference

* `cache` - JSON string containing the deduplicated rule cache. Keyed by action_hash, each entry contains action conditions and point groups. Two kinds of groups: stamp groups (keyed by point_hash, containing stamps) and attack_type groups (keyed by point_hash + attack_type).
* `request_to_action` - JSON string mapping each request_id to its action_hash. Enables tracing which request IDs contributed to which cache entries. Used for cleanup when request IDs are removed.

## How Deduplication Works

When `new_entries` contains data for a request_id whose action_hash already exists in the cache:

1. Groups are matched by key (point_hash for stamp groups, point_hash + attack_type for attack_type groups)
2. Matching groups: stamps are unioned (sorted, deduplicated)
3. New groups (key not in existing): added to the cache entry

When a request_id is removed from `request_ids`:

1. Its entry is removed from `request_to_action`
2. If no other request_id references the same action_hash, the cache entry is removed
3. Downstream rules using that cache entry are destroyed on the next apply
