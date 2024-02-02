---
layout: "wallarm"
page_title: "Wallarm: wallarm_tenant"
subcategory: "Common"
description: |-
  Provides the resource to manage tenants of the company.
---

# wallarm_tenant

Provides the resource to manage [tenants][1] of the company. To use this resource, your token has to have the 'Global Administrator' role.

## Example Usage

```hcl
# Creates a new tenant "Tenant 1"

resource "wallarm_tenant" "tenant1" {
  name = "Tenant 1"
  client_id = 123
}
```

## Argument Reference

* `name` - (**required**) tenant name.
* `client_id` - (optional) ID of the client which is a partner for the created tenant. By default, this argument has the value of the current client ID.


## Attributes Reference

* `tenant_id` - client ID of the created tenant.

[1]: https://docs.wallarm.com/installation/multi-tenant/overview/
