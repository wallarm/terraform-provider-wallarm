---
layout: "wallarm"
page_title: "Wallarm: wallarm_hits_index"
subcategory: "Common"
description: |-
  Tracks fetched request IDs for the hits-to-rules workflow.
---

# wallarm_hits_index

Tracks which request IDs have had their hits fetched. Used to gate `data.wallarm_hits` so only uncached request IDs trigger API calls.

This is a **state-only** resource -- it makes no API calls. On first create, `ready` is `false` (triggers fetching all request IDs). After create, `ready` becomes `true` and `cached_request_ids` reflects the current `request_ids` set -- enabling gating to only fetch new IDs.

## Example Usage

```hcl
resource "wallarm_hits_index" "this" {
  request_ids = keys(var.request_ids)
}

locals {
  # On first create (ready=false): fetch all request_ids.
  # After create (ready=true): only fetch IDs not in cached_request_ids.
  _request_ids_to_fetch = wallarm_hits_index.this.ready ? toset([
    for id in keys(var.request_ids) : id
    if !contains(wallarm_hits_index.this.cached_request_ids, id)
  ]) : toset(keys(var.request_ids))
}

data "wallarm_hits" "new" {
  for_each   = local._request_ids_to_fetch
  request_id = each.key
}
```

For complete usage including rule creation and caching, see the [Hits to Rules Guide](../guides/hits_to_rules).

## Argument Reference

* `client_id` - (Optional) Client ID. Defaults to the provider's client ID.
* `request_ids` - (Required) Set of request IDs to track.

## Attributes Reference

* `ready` - Boolean. `false` on first create (known at plan time), `true` after. Use to control gating: when `false`, fetch all request IDs; when `true`, only fetch IDs not in `cached_request_ids`.
* `cached_request_ids` - Set of request IDs currently tracked. Synced to match `request_ids` on each refresh.
