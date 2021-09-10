---
layout: "wallarm"
page_title: "Wallarm: wallarm_node"
subcategory: "Common"
description: |-
  Provides the resource to manage nodes of the account.
---

# wallarm_user

Provides the resource to manage cloud nodes of the account.

## Example Usage

```hcl
# Creates a new cloud node

resource "wallarm_node" "terraform" {
  client_id = 6039
  hostname = "Terraform Tests"
}

```

## Argument Reference

* `hostname` - (Required) Node name. The value must be unique across other nodes in the account.
* `client_id` - (Optional) ID of the client to apply the rules to. The value is required for multi-tenant scenarios.

## Attributes Reference

* `node_id` - Unique ID (numerical) of the created node.
* `node_uuid` - Unique UUID of the created node.
* `token` - Token of the cloud node.
