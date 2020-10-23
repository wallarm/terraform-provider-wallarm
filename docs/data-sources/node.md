---
layout: "wallarm"
page_title: "Wallarm: wallarm_node"
subcategory: "Common"
description: |-
  Get Wallarm WAF node details.
---

# wallarm_node

Use this data source to get the [WAF node][1] details.

## Example usage

In the given example, it is assumed that there are three cloud WAF nodes with the following names:

- gcp-production
- aws-staging
- azure-development

```hcl
# Looks up for the WAF node details by its exact name which is unique over an account

data "wallarm_node" "aws_staging" {
  filter {
    hostname = "aws-staging"
  }
}
```

```hcl
# Looks up for details on WAF nodes with the specific type (can be "cloud_node" or "node")

data "wallarm_node" "cloud_nodes" {
  filter {
    type = "cloud_node"
  }
}
```

```hcl
# Looks up for the WAF node details by its UUID which is unique over an account

data "wallarm_node" "example" {
  filter {
    uuid = "b161e6f9-33d2-491e-a584-513522d312db"
  }
}
```

## Argument Reference

`filter` - (Required) Filters set in the `key=value` format used to look up for WAF node details. Possible keys:

- `uuid` - (Optional) WAF node UUID.
- `hostname` - (Optional) WAF node name.
- `type` - (Optional) WAF node type. Can be: `cloud_node` for cloud WAF nodes, `node` for regular WAF nodes.
- `enabled` - (Optional) Indicator of the WAF node status. Can be: `true` for enabled WAF nodes and `false` for disabled WAF nodes.

To get details on all created WAF nodes, specify an empty set of the filters (`filter {}`).

## Attributes Reference

`nodes` - WAF node attributes in the `key=value` format. Possible keys:

- `id` - Internal WAF node ID.
- `hostname` - WAF node name.
- `type` - WAF node type. Can be: `cloud_node` for cloud WAF nodes, `node` for regular WAF nodes.
- `uuid` - WAF node UUID.
- `enabled` - Indicator of the WAF node status. Can be: `true` for enabled WAF nodes and `false` for disabled WAF nodes.
- `client_id` - ID of the client installed the WAF node.
- `active` - Node syncronisation status
- `instance_count` - Number of instances with the installed WAF node. Only for the `cloud_node` type.
- `active_instance_count` - Number of active instances with the installed WAF node. Only for the `cloud_node` type.
- `token` - WAF node token. Only for the `cloud_node` type.
- `requests_amount` - Number of requests processed by the WAF node.
- `ip` - IP address of the WAF node at the time of the last synchronization.
- `proton` - Installed `proton.db` version. Can be an integer value or `null`.
- `lom` - Installed `lom` version. Can be an integer value or `null`.

[1]: https://docs.wallarm.com/user-guides/nodes/nodes/
