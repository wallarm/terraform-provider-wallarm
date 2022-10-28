---
layout: "wallarm"
page_title: "Wallarm: wallarm_denylist"
subcategory: "Common"
description: |-
  Provides the resource to manage denylist in the account.
---

# wallarm_denylist

Provides the resource to manage [denylist][1] in the account providing functionality to block all requests originated from denylisted IP addresses for a desired time either via `Date` or `Minutes` formats.

## Example Usage

```hcl
# Creates two new denylist entries
# with the determined block time (60 minutes) and until the 2nd of January 2026

resource "wallarm_denylist" "denylist_minutes" {
  ip_range = ["1.1.1.1/32"]
  application = [1]
  reason = "TEST DENYLIST MINUTES"
  time_format = "Minutes"
  time = 60
}

resource "wallarm_denylist" "denylist_date" {
  ip_range = ["2.2.2.2/32"]
  application = [1]
  reason = "TEST DENYLIST DATE"
  time_format = "RFC3339"
  time = "2026-01-02T15:04:05+07:00"
}
```

## Argument Reference

* `ip_range` - (**required**) IP range to be blocked. Can be defined as an array of ranges. Accept:
  - distinct IP addresses (e.g. `1.1.1.1`, `2.2.2.2`)
  - subnets (e.g. `1.1.1.1/24`, `2.2.2.2/30`)
* `time_format` - (**required**) block time format.
  Can be:
  - `Minutes` - Time in minutes (e.g. `60` is to block for 60 minutes)
  - `RFC3339` - RFC3339 time (e.g. `2021-06-01T15:04:05+07:00`)
* `time` - (**required**) time for (or until) which the IP address should be blocked.
* `application` - (optional) list of application IDs. 
  Default: all applications
* `reason` - (optional) arbitrary reason of blocking these IP addresses.
* `client_id` - (optional) ID of the client (tenant). The value is required for [multi-tenant scenarios][2].

## Attributes Reference

`address_id` - addresses ID attributes in the `key=value` format. Possible keys:

- `ip_addr` - discrete IP address.
- `ip_id` - ID of the entry with the concrete IP address.

[1]: https://docs.wallarm.com/user-guides/ip-lists/denylist/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
