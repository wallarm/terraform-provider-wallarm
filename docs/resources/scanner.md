---
layout: "wallarm"
page_title: "Wallarm: wallarm_scanner"
subcategory: "Common"
description: |-
  Provides the resource to manage Wallarm list of exposed assets.
---

# wallarm_scanner

Provides the resource to manage the list of exposed assets [scanned][1] by Wallarm for typical vulnerabilities.

## Example Usage

```hcl
# Adds "1.1.1.1", "example.com", "2.2.2.2/31" to the list of exposed assets
# and enables these elements to be subjected for scanning

resource "wallarm_scanner" "scan" {
    element = ["1.1.1.1", "example.com", "2.2.2.2/31"]
    disabled = false
}
```

## Argument Reference

* `element` - (**required**) array of IP addresses, subnets, domains to add to the list of exposed assets.
* `disabled` - (**required**) indicator of a need to scan specified elements. Can be: `true` to add the elements to the list of exposed assets but not scan, `false` to add the elements to the list and scan them.
* `client_id` - (optional) exposed assets are added to list of client with this ID. The value is required for [multi-tenant scenarios][2].

## Attributes Reference

* `resource_id` - ID of the added element. The value is used to control the state of an elements since all the API requests use the unique ID.

[1]: https://docs.wallarm.com/user-guides/scanner/intro/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
