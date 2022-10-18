---
layout: "wallarm"
page_title: "Wallarm: wallarm_node"
subcategory: "Common"
description: |-
  Get Wallarm node details.
---

# wallarm_node

Use this data source to get the [Wallarm node][1] details.

## Example usage

In the given example, it is assumed that there are three cloud Wallarm nodes with the following names:

- gcp-production
- aws-staging
- azure-development

```hcl
# Looks up for the Wallarm node details by its exact name which is unique over an account

data "wallarm_node" "aws_staging" {
  filter {
    hostname = "aws-staging"
  }
}
```

```hcl
# Looks up for details on Wallarm nodes with the specific type (can be "cloud_node" or "node")

data "wallarm_node" "cloud_nodes" {
  filter {
    type = "cloud_node"
  }
}
```

```hcl
# Looks up for the Wallarm node details by its UUID which is unique over an account

data "wallarm_node" "example" {
  filter {
    uuid = "b161e6f9-33d2-491e-a584-513522d312db"
  }
}
```

## Argument Reference

`filter` - (Required) Filters set in the `key=value` format used to look up for Wallarm node details. Possible keys:

- `uuid` - (Optional) Wallarm node UUID.
- `hostname` - (Optional) Wallarm node name.
- `type` - (Optional) Wallarm node type. Can be: `cloud_node` for cloud Wallarm nodes, `node` for regular Wallarm nodes.
- `enabled` - (Optional) Indicator of the Wallarm node status. Can be: `true` for enabled Wallarm nodes and `false` for disabled Wallarm nodes.

To get details on all created Wallarm nodes, specify an empty set of the filters (`filter {}`).

## Attributes Reference

`nodes` - Wallarm node attributes in the `key=value` format. Possible keys:

- `id` - Internal Wallarm node ID.
- `hostname` - Wallarm node name.
- `type` - Wallarm node type. Can be: `cloud_node` for cloud Wallarm nodes, `node` for regular Wallarm nodes.
- `uuid` - Wallarm node UUID.
- `enabled` - Indicator of the Wallarm node status. Can be: `true` for enabled Wallarm nodes and `false` for disabled Wallarm nodes.
- `client_id` - ID of the client installed the Wallarm node.
- `active` - Node syncronisation status
- `instance_count` - Number of instances with the installed Wallarm node. Only for the `cloud_node` type.
- `active_instance_count` - Number of active instances with the installed Wallarm node. Only for the `cloud_node` type.
- `token` - Wallarm node token. Only for the `cloud_node` type.
- `requests_amount` - Number of requests processed by the Wallarm node.
- `ip` - IP address of the Wallarm node at the time of the last synchronization.
- `proton` - Installed `proton.db` version. Can be an integer value or `null`.
- `lom` - Installed `lom` version. Can be an integer value or `null`.

[1]: https://docs.wallarm.com/user-guides/nodes/nodes/
