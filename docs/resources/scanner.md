---
layout: "wallarm"
page_title: "Wallarm: wallarm_scanner"
subcategory: "Common"
description: |-
  Provides the resource to manage Wallarm Scanner scope.
---

# wallarm_scanner

Provides the resource to manage the Wallarm Scanner scope. The scope defines company resources that must be scanned by the Wallarm dynamic application security testing (DAST) tool.

## Example Usage

```hcl
# Adds "1.1.1.1", "example.com", "2.2.2.2/31" to the Scanner scope
# and enables these elements to be subjected for scanning

resource "wallarm_scanner" "scan" {
    element = ["1.1.1.1", "example.com", "2.2.2.2/31"]
    disabled = false
}
```

## Argument Reference

* `element` - (Required) Array of IP addresses, subnets, domains to add to the Scanner scope.
* `disabled` - (Required) Indicator of a need to scan specified elements. Can be: `true` to add the elements to the scope but not scan, `false` to ass the elements to the scope and scan them.
* `client_id` - (Optional) ID of the client to add the elements to the scope of. The value is required for multi-tenant scenarios.

## Attributes Reference

* `resource_id` - ID of the added element. The value is used to control the state of an elements since all the API requests use the unique ID.