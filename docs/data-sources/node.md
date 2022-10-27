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

In the given example, it is assumed that there are three Wallarm nodes with the following names:

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
# Looks up for the Wallarm node details by its UUID which is unique over an account

data "wallarm_node" "example" {
  filter {
    uuid = "b161e6f9-33d2-491e-a584-513522d312db"
  }
}
```

## Argument Reference

`filter` - (**required**) filters set in the `key=value` format used to look up for Wallarm node details. Possible keys:

- `uuid` - (optional) Wallarm node UUID.
- `hostname` - (optional) Wallarm node name.
- `enabled` - (optional) indicator of the Wallarm node status. Can be: `true` for enabled Wallarm nodes and `false` for disabled Wallarm nodes.

To get details on all created Wallarm nodes, specify an empty set of the filters (`filter {}`).

## Attributes Reference

`nodes` - Wallarm node attributes in the `key=value` format. Possible keys:

- `id` - internal Wallarm node ID.
- `hostname` - Wallarm node name.
- `type` - Wallarm node type. Can be: `cloud_node` for cloud Wallarm nodes, `node` for regular Wallarm nodes. See [node types description](https://docs.wallarm.com/3.6/user-guides/nodes/nodes/#filtering-the-nodes). Note that `CDN` node is not supported.
- `uuid` - Wallarm node UUID.
- `enabled` - indicator of the Wallarm node status. Can be: `true` for enabled Wallarm nodes and `false` for disabled Wallarm nodes.
- `client_id` - ID of the client installed the Wallarm node.
- `active` - node synchronization status.
- `instance_count` - number of instances with the installed Wallarm node. Only for the `cloud_node` type.
- `active_instance_count` - number of active instances with the installed Wallarm node. Only for the `cloud_node` type.
- `token` - Wallarm node token. Only for the `cloud_node` type.
- `requests_amount` - number of requests processed by the Wallarm node.
- `ip` - IP address of the Wallarm node at the time of the last synchronization.
- `proton` - installed `proton.db` version. Can be an integer value or `null`.
- `lom` - installed `lom` (currently named "[custom ruleset](https://docs.wallarm.com/user-guides/rules/compiling/)") version. Can be an integer value or `null`.

[1]: https://docs.wallarm.com/user-guides/nodes/nodes/
