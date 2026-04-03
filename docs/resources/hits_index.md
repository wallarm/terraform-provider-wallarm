---
layout: "wallarm"
page_title: "Wallarm: wallarm_hits_index"
subcategory: "Common"
description: |-
  Tracks fetched request IDs for the hits-to-rules workflow.
---

# wallarm_hits_index

Tracks which request IDs have had their hits fetched. Used to gate `data.wallarm_hits` so only uncached request IDs trigger API calls.

This is a **state-only** resource -- it makes no API calls. The `cached_request_ids` output reflects the current `request_ids` input after each refresh.

## Example Usage

```hcl
resource "wallarm_hits_index" "this" {
  request_ids = keys(var.request_ids)
}

locals {
  _cached = toset(compact(split(",", wallarm_hits_index.this.cached_request_ids)))

  _new_request_ids = toset([
    for id in keys(var.request_ids) : id
    if !contains(local._cached, id)
  ])
}

data "wallarm_hits" "new" {
  for_each   = local._new_request_ids
  request_id = each.key
}
```

For complete usage including rule creation and caching, see the [Hits to Rules Guide](../guides/hits_to_rules).

## Argument Reference

* `client_id` - (Optional) Client ID. Defaults to the provider's client ID.
* `request_ids` - (Required) Set of request IDs to track.

## Attributes Reference

* `cached_request_ids` - Comma-separated string of request IDs currently tracked. Used downstream to gate `data.wallarm_hits` so only uncached request IDs trigger API calls.
