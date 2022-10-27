---
layout: "wallarm"
page_title: "Wallarm: wallarm_node"
subcategory: "Common"
description: |-
  Provides the resource to manage nodes of the account.
---

# wallarm_node

Provides the resource to manage [Wallarm nodes][1] of the account.

## Example Usage

```hcl
# Creates a new Wallarm node

resource "wallarm_node" "terraform" {
  client_id = 6039
  hostname = "Terraform Tests"
}

```

## Argument Reference

* `hostname` - (**required**) node name. The value must be unique across other nodes in the account.
* `client_id` - (optional) ID of the client (tenant). The value is required for [multi-tenant scenarios][2].

## Attributes Reference

* `node_id` - Unique ID (numerical) of the created node.
* `node_uuid` - Unique UUID of the created node.
* `token` - Token of the Wallarm node.

[1]: https://docs.wallarm.com/user-guides/nodes/nodes/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
