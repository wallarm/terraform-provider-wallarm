---
layout: "wallarm"
page_title: "Wallarm: wallarm_hits_index"
subcategory: "Common"
description: |-
  Provides a persistent index of fetched request IDs for the hits-to-rules workflow.
---

# wallarm_hits_index

Tracks which request IDs have had their hits fetched. Used with `data.wallarm_hits` and `terraform_data` to implement a persistent cache for the [hits-to-rules workflow](../guides/hits_to_rules).

This is a **state-only** resource -- it makes no API calls. It stores the set of request IDs as a comma-separated string in Terraform state, enabling gated fetching of hit data.

## Example Usage

```hcl
resource "wallarm_hits_index" "this" {
  request_ids = ["abc123", "def456"]
}
```

```hcl
# Multi-tenant
resource "wallarm_hits_index" "this" {
  client_id   = 8649
  request_ids = keys(var.request_ids)
}
```

For complete usage including hit fetching and rule creation, see the [Hits to Rules Guide](../guides/hits_to_rules).

## Argument Reference

* `client_id` - (Optional) Client ID. Defaults to the provider's client ID.
* `request_ids` - (Required) Set of request IDs to track.

## Attributes Reference

* `cached_request_ids` - Comma-separated string of request IDs currently in the index. Used downstream to gate `data.wallarm_hits` fetches.
