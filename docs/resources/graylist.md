---
layout: "wallarm"
page_title: "Wallarm: wallarm_graylist"
subcategory: "IP Lists"
description: |-
  Provides the resource to manage graylist in the account.
---

# wallarm_graylist

Provides the resource to manage [graylist][1] in the account. Graylisted entries are only blocked when they send malicious requests — legitimate traffic is allowed through. The graylist is processed by the node only in the safe blocking [filtration mode](https://docs.wallarm.com/admin-en/configure-wallarm-mode/).

Supports four mutually exclusive entry types: IP addresses/subnets, countries, datacenters, and proxy types. Only one type can be used per resource.

## Example Usage

### IP addresses

```hcl
resource "wallarm_graylist" "ips_minutes" {
  ip_range    = ["1.1.1.1/32", "2.2.2.0/24"]
  application = [1]
  reason      = "Suspicious IPs"
  time_format = "Minutes"
  time        = 60
}

resource "wallarm_graylist" "ips_forever" {
  ip_range    = ["5.5.5.5"]
  reason      = "Permanently graylisted"
  time_format = "Forever"
}
```

### Countries

```hcl
resource "wallarm_graylist" "countries" {
  country     = ["CN", "RU"]
  reason      = "Graylist by country"
  time_format = "Forever"
}
```

### Datacenters

```hcl
resource "wallarm_graylist" "datacenters" {
  datacenter  = ["aws", "gce"]
  reason      = "Graylist cloud providers"
  time_format = "Months"
  time        = 6
}
```

### Proxy types

```hcl
resource "wallarm_graylist" "proxies" {
  proxy_type  = ["TOR", "VPN"]
  reason      = "Graylist anonymous proxies"
  time_format = "Forever"
}
```

## Argument Reference

One of `ip_range`, `country`, `datacenter`, or `proxy_type` is required. They are mutually exclusive.

* `ip_range` - (optional) List of IP addresses or subnets to graylist. Maximum **1000** entries per resource.
  - Distinct IPs: `"1.1.1.1"`, `"2.2.2.2"`
  - Subnets: `"1.1.1.1/24"`, `"2.2.2.2/30"`
* `country` - (optional) List of country codes (ISO 3166-1 alpha-2). Example: `["US", "DE", "CN"]`
* `datacenter` - (optional) List of datacenter providers.
  Valid values: `alibaba`, `aws`, `azure`, `docean`, `gce`, `hetzner`, `huawei`, `ibm`, `linode`, `oracle`, `ovh`, `plusserver`, `rackspace`, `tencent`
* `proxy_type` - (optional) List of proxy types.
  Valid values: `DCH`, `MIP`, `PUB`, `WEB`, `SES`, `TOR`, `VPN`
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
* `reason` - (optional) Reason for graylisting. Default: `"Terraform managed IP list"`.
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

```bash
# Grouped types (country/datacenter/proxy): import by group ID
$ terraform import wallarm_graylist.countries 8649/52000393

# Subnets: import all IPs with same expiration
$ terraform import wallarm_graylist.ips 8649/subnet/1804809600

# Subnets scoped to specific applications
$ terraform import wallarm_graylist.ips_app 8649/subnet/1804809600/apps/1,3

# Subnets with no application filter
$ terraform import wallarm_graylist.ips_all 8649/subnet/1804809600/apps/all

# Subnets (chunked): when >1000 IPs share the same expiration + app scope
$ terraform import wallarm_graylist.ips_0 8649/subnet/1804809600/apps/all/0
```

| Format | Use case |
|--------|----------|
| `{clientID}/{groupID}` | Country, datacenter, proxy type |
| `{clientID}/subnet/{expiredAt}` | Subnets with same expiration and same app scope |
| `{clientID}/subnet/{expiredAt}/apps/{appIDs}` | Subnets filtered by application scope (`1,3` or `all`) |
| `{clientID}/subnet/{expiredAt}/apps/{appIDs}/{chunkIdx}` | Chunked import for >1000 subnets (0-indexed) |

If the simple format (`subnet/{expiredAt}`) is used but entries have mixed application scopes, the import will error with guidance to use the `/apps/{appIDs}` format.

For automated bulk import using the `wallarm_ip_lists` data source, see the [IP List Import Guide](../guides/ip_list_import).

[1]: https://docs.wallarm.com/user-guides/ip-lists/graylist/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
