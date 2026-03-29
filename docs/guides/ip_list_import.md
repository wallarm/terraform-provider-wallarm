---
layout: "wallarm"
page_title: "Importing Wallarm IP Lists"
subcategory: "Guide"
description: |-
  How to import existing Wallarm IP list entries into Terraform.
---

# Importing Wallarm IP Lists

IP list entries created via the Console UI or API can be imported into Terraform.

## Quick Start

### Import blocked countries

Countries are stored as a single API group. Find the group ID in the Wallarm Console or via the `wallarm_ip_lists` data source, then import by group ID:

```bash
terraform import wallarm_denylist.countries 8649/52000393
```

Write the matching resource:

```hcl
resource "wallarm_denylist" "countries" {
  country     = ["CN", "RU"]
  reason      = "Block by country"
  time_format = "Forever"
}
```

Run `terraform plan` to verify — no changes expected.

### Import blocked IPs

Subnets are imported by expiration timestamp. All IPs with the same `expired_at` are merged into one resource:

```bash
terraform import wallarm_denylist.ips 8649/subnet/1804809600
```

Write the matching resource:

```hcl
resource "wallarm_denylist" "ips" {
  ip_range    = ["1.1.1.1", "2.2.2.0/24", "3.3.3.3"]
  reason      = "Blocked IPs"
  time_format = "RFC3339"
  time        = "2027-03-15T00:00:00+00:00"
}
```

Or let Terraform generate the config automatically:

```bash
terraform plan -generate-config-out=generated.tf
```

### Import IPs scoped to specific applications

If IPs with the same expiration are assigned to different applications, import each scope separately using the `/apps/{appIDs}` format:

```bash
# IPs scoped to applications 1 and 3
terraform import wallarm_denylist.ips_app1 8649/subnet/1804809600/apps/1,3

# IPs with no application filter
terraform import wallarm_denylist.ips_all 8649/subnet/1804809600/apps/all
```

## Import ID Reference

| Format | Use case | Example |
|--------|----------|---------|
| `{clientID}/{groupID}` | Country, datacenter, proxy type | `8649/52000393` |
| `{clientID}/subnet/{expiredAt}` | Subnets with same expiration and same app scope | `8649/subnet/1804809600` |
| `{clientID}/subnet/{expiredAt}/apps/{appIDs}` | Subnets filtered by application scope | `8649/subnet/1804809600/apps/1,3` |
| `{clientID}/subnet/{expiredAt}/apps/{appIDs}/{chunkIdx}` | Chunked import for >1000 subnets | `8649/subnet/1804809600/apps/all/0` |

Notes:
- **Grouped types** (country/datacenter/proxy) — the API stores all values of the same type as a single group with one ID.
- **Subnets** — each IP is a separate API group. The importer merges all IPs with the same expiration into one resource.
- **Application scopes** — `{appIDs}` is sorted comma-separated IDs (e.g. `1,3`) or `all`. If the simple format is used but entries have mixed app scopes, the importer errors with guidance.
- **Chunking** — when a group exceeds 1000 IPs, use `/{chunkIdx}` (0-indexed, 1000 per chunk). IPs are sorted lexicographically for deterministic boundaries.

Resource documentation:
[`wallarm_allowlist`](../resources/allowlist),
[`wallarm_denylist`](../resources/denylist),
[`wallarm_graylist`](../resources/graylist).

## Automated Bulk Import

For large IP lists, use the [`wallarm_ip_lists`](../data-sources/ip_lists) data source to discover all entries and generate import blocks automatically.

The example below handles all entry types — countries, datacenters, proxy types, and subnets — with proper grouping by expiration, application scope, and chunking for large sets.

```hcl
data "wallarm_ip_lists" "deny" {
  list_type = "denylist"
}

locals {
  max_subnets_per_resource = 1000

  deny_entries = data.wallarm_ip_lists.deny.entries
  client_id    = 8649

  # Grouped types (country/datacenter/proxy): one import per API group
  grouped_blocks = [
    for e in local.deny_entries :
    <<-EOT
    import {
      to = wallarm_denylist.import_${e.rule_type}_${e.id}
      id = "${local.client_id}/${e.id}"
    }
    EOT
    if contains(["location", "datacenter", "proxy_type"], e.rule_type)
  ]

  # Subnets: group by (expired_at, application_ids), chunk into max 1000 per resource.
  # Entries with the same expiration but different application scopes become separate resources.

  # Build a canonical app_ids key for grouping: sorted IDs joined by comma, or "all".
  subnet_entries_with_key = [
    for e in local.deny_entries : merge(e, {
      app_key = length(e.application_ids) == 0 ? "all" : join(",", sort(e.application_ids))
    })
    if e.rule_type == "subnet"
  ]

  # Unique (expired_at, app_key) pairs.
  subnet_group_keys = distinct([
    for e in local.subnet_entries_with_key : "${e.expired_at}/${e.app_key}"
  ])

  # Group entries by the composite key.
  subnets_by_scope = {
    for key in local.subnet_group_keys :
    key => [for e in local.subnet_entries_with_key : e if "${e.expired_at}/${e.app_key}" == key]
  }

  subnet_chunks = flatten([
    for key, entries in local.subnets_by_scope : [
      for idx in range(ceil(length(entries) / local.max_subnets_per_resource)) : {
        # When total entries fit in one chunk, use the simple format (no chunk index).
        # When multiple chunks are needed, ALL chunks get an index — including chunk 0.
        needs_chunking = length(entries) > local.max_subnets_per_resource
        exp            = entries[0].expired_at
        app_key        = entries[0].app_key
        idx            = idx
      }
    ]
  ])

  subnet_import_blocks = [
    for s in local.subnet_chunks :
    <<-EOT
    import {
      to = wallarm_denylist.import_subnet_${s.exp}_${replace(s.app_key, ",", "_")}${s.needs_chunking ? "_${s.idx}" : ""}
      id = "${local.client_id}/subnet/${s.exp}/apps/${s.app_key}${s.needs_chunking ? "/${s.idx}" : ""}"
    }
    EOT
  ]
}

resource "local_file" "deny_imports" {
  filename = "./wallarm_denylist_imports.tf"
  content  = join("\n", concat(local.grouped_blocks, local.subnet_import_blocks))
}
```

**Steps:**

1. `terraform apply` — generates `wallarm_denylist_imports.tf` with import blocks
2. Copy the file to your target configuration directory
3. `terraform plan -generate-config-out=generated.tf` — generates resource configs
4. Review and adjust the generated resources, then `terraform apply`

### Import All Three List Types

Query all lists in a single configuration:

```hcl
data "wallarm_ip_lists" "deny" {
  list_type = "denylist"
}

data "wallarm_ip_lists" "allow" {
  list_type = "allowlist"
}

data "wallarm_ip_lists" "gray" {
  list_type = "graylist"
}
```

Apply the same grouping logic from the example above to each list type.

## Grouping Recommendations

- **One entry type per resource** — `ip_range`, `country`, `datacenter`, and `proxy_type` are mutually exclusive
- **Group subnets by expiration** — IPs with the same lifetime naturally belong in one resource
- **Separate by application scope** — IPs scoped to different applications should be separate resources
- **Stay within the 1000 IP limit** — split large sets into multiple resources

### Example: Well-structured IP list config

```hcl
# Permanent blocks — no expiration
resource "wallarm_denylist" "permanent_ips" {
  ip_range    = ["1.1.1.1", "2.2.2.0/24"]
  reason      = "Permanently blocked"
  time_format = "Forever"
}

# Temporary blocks — same expiration
resource "wallarm_denylist" "temp_ips" {
  ip_range    = ["3.3.3.3", "4.4.4.4"]
  reason      = "Blocked until end of month"
  time_format = "RFC3339"
  time        = "2026-04-01T00:00:00+00:00"
}

# Countries — all values in one resource
resource "wallarm_denylist" "countries" {
  country     = ["CN", "RU"]
  reason      = "Block by country"
  time_format = "Forever"
}

# Datacenters — all values in one resource
resource "wallarm_denylist" "datacenters" {
  datacenter  = ["aws", "gce"]
  reason      = "Block cloud providers"
  time_format = "Forever"
}
```

## Important Notes

- **Plan after import:** Always run `terraform plan` after importing to verify the configuration matches actual state.
- **Bare IPs are normalized:** The API normalizes `1.2.3.4` to `1.2.3.4/32`. The provider handles this transparently.
- **Multi-tenant:** Specify `client_id` in the data source and resources when working with multiple tenants.
