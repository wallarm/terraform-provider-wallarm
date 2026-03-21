---
layout: "wallarm"
page_title: "Wallarm: wallarm_denylist"
subcategory: "Common"
description: |-
  Provides the resource to manage denylist in the account.
---

# wallarm_denylist

Provides the resource to manage [denylist][1] in the account. Denylisted entries block all requests from the specified sources for a desired time.

Supports four mutually exclusive entry types: IP addresses/subnets, countries, datacenters, and proxy types. Only one type can be used per resource.

## Example Usage

### IP addresses

```hcl
resource "wallarm_denylist" "ips_minutes" {
  ip_range    = ["1.1.1.1/32", "2.2.2.0/24"]
  application = [1]
  reason      = "Blocked IPs"
  time_format = "Minutes"
  time        = 60
}

resource "wallarm_denylist" "ips_date" {
  ip_range    = ["3.3.3.3"]
  reason      = "Blocked until date"
  time_format = "RFC3339"
  time        = "2026-06-01T00:00:00+00:00"
}

resource "wallarm_denylist" "ips_forever" {
  ip_range    = ["4.4.4.4"]
  reason      = "Permanently blocked"
  time_format = "Forever"
}
```

### Countries

```hcl
resource "wallarm_denylist" "countries" {
  country     = ["CN", "RU"]
  reason      = "Block by country"
  time_format = "Forever"
}
```

### Datacenters

```hcl
resource "wallarm_denylist" "datacenters" {
  datacenter  = ["aws", "gce"]
  reason      = "Block cloud providers"
  time_format = "Months"
  time        = 12
}
```

### Proxy types

```hcl
resource "wallarm_denylist" "proxies" {
  proxy_type  = ["TOR", "VPN"]
  reason      = "Block anonymous proxies"
  time_format = "Forever"
}
```

## Argument Reference

One of `ip_range`, `country`, `datacenter`, or `proxy_type` is required. They are mutually exclusive.

* `ip_range` - (optional) List of IP addresses or subnets to deny. Maximum **1000** entries per resource.
  - Distinct IPs: `"1.1.1.1"`, `"2.2.2.2"`
  - Subnets: `"1.1.1.1/24"`, `"2.2.2.2/30"`
* `country` - (optional) List of country codes (ISO 3166-1 alpha-2). Example: `["US", "DE", "CN"]`
* `datacenter` - (optional) List of datacenter providers.
  Valid values: `alibaba`, `aws`, `azure`, `docean`, `gce`, `hetzner`, `huawei`, `ibm`, `linode`, `oracle`, `ovh`, `plusserver`, `rackspace`, `tencent`
* `proxy_type` - (optional) List of proxy types.
  Valid values: `MIP`, `PUB`, `WEB`, `SES`, `TOR`, `VPN`
* `time_format` - (**required**) Time format for the entry duration.
  - `Minutes` - Time in minutes (e.g. `60`)
  - `Hours` - Time in hours (e.g. `5`)
  - `Days` - Time in days (e.g. `7`)
  - `Weeks` - Time in weeks (e.g. `4`)
  - `Months` - Time in months (e.g. `12`)
  - `RFC3339` - Absolute date/time (e.g. `"2026-06-01T00:00:00+00:00"`)
  - `Forever` - No expiration (`time` is not required)
* `time` - (optional) Duration or expiration time. Required for all `time_format` values except `Forever`.
* `application` - (optional) List of application IDs. Default: all applications.
* `reason` - (optional) Reason for denylisting. Default: `"Terraform managed IP list"`.
* `client_id` - (optional) ID of the client (tenant). Required for [multi-tenant scenarios][2].

## Attributes Reference

* `entry_count` - Number of config values successfully found in the API.
* `untracked_count` - Number of config values not found in the API.
* `untracked_ips` - List of config values not found in the API.
* `address_id` - List of tracked entries, each containing:
  - `rule_type` - Entry type (`subnet`, `location`, `datacenter`, `proxy_type`).
  - `value` - The entry value (IP, country code, etc.).
  - `ip_id` - API group ID.

## Import

Denylist resources can be imported:

```bash
# Grouped types (country/datacenter/proxy): import by group ID
terraform import wallarm_denylist.countries 8649/52000393

# Subnets: import all IPs with same expiration as one resource
terraform import wallarm_denylist.ips 8649/subnet/1804809600
```

[1]: https://docs.wallarm.com/user-guides/ip-lists/denylist/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
