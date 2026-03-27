---
layout: "wallarm"
page_title: "Importing Wallarm IP Lists"
subcategory: "Guide"
description: |-
  How to import existing Wallarm IP list entries into Terraform.
---

# Importing Wallarm IP Lists

IP list entries created via the Console UI or API can be imported into Terraform. This guide covers the import ID formats, API grouping behavior, and automated import using the `wallarm_ip_lists` data source.

## Import ID Formats

IP list resources support three import ID formats depending on the entry type:

| Format | Use case | Example |
|--------|----------|---------|
| `{clientID}/{groupID}` | Country, datacenter, proxy type | `8649/52000393` |
| `{clientID}/subnet/{expiredAt}` | All subnets with same expiration (<=1000) | `8649/subnet/1804809600` |
| `{clientID}/subnet/{expiredAt}/{chunkIdx}` | Chunked import for large subnet sets (>1000) | `8649/subnet/1804809600/0` |

- **Grouped types** (country/datacenter/proxy) — the API stores all values of the same type as a single group with one ID. Importing by group ID brings in all values at once.
- **Subnets** — each IP is a separate API group. The `subnet/{expiredAt}` format merges all IPs with the same expiration timestamp into one resource.
- **Chunked subnets** — when a single expiration group contains more than 1000 IPs, use the `/{chunkIdx}` suffix (0-indexed) to import in batches of 1000. IPs are sorted lexicographically for deterministic chunk boundaries.

Refer to the individual resource documentation for import command examples:
[`wallarm_allowlist`](../resources/allowlist),
[`wallarm_denylist`](../resources/denylist),
[`wallarm_graylist`](../resources/graylist).

## API Grouping Behavior

Understanding how the API groups entries is important for planning your import:

| Rule type | API grouping | Terraform resource |
|-----------|-------------|-------------------|
| `location` (country) | All values in **1 group** | 1 resource per group |
| `datacenter` | All values in **1 group** | 1 resource per group |
| `proxy_type` | All values in **1 group** | 1 resource per group |
| `subnet` (IPs) | **1 group per IP** | Merged by expiration — all IPs with same `expired_at` become one resource |

Subnet resources are limited to **1000 IPs per resource**. If a non-chunked import finds more than 1000 subnets with the same expiration, it returns an error with instructions to use the chunked format.

## Automated Import with the IP Lists Data Source

The [`wallarm_ip_lists`](../data-sources/ip_lists) data source discovers all existing entries for a given list type. This enables generating import blocks automatically.

### Example: Import All Denylist Entries

**1. Create a discovery configuration:**

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

  # Subnets: merge by expired_at, chunk into max 1000 per resource
  subnet_expiries = distinct([
    for e in local.deny_entries : e.expired_at
    if e.rule_type == "subnet"
  ])

  subnets_by_expiry = {
    for exp in local.subnet_expiries :
    exp => [for e in local.deny_entries : e if e.rule_type == "subnet" && e.expired_at == exp]
  }

  subnet_chunks = flatten([
    for exp, entries in local.subnets_by_expiry : [
      for idx in range(ceil(length(entries) / local.max_subnets_per_resource)) : {
        # When total entries fit in one chunk, use the simple format (no chunk index).
        # When multiple chunks are needed, ALL chunks get an index — including chunk 0.
        needs_chunking = length(entries) > local.max_subnets_per_resource
        exp            = exp
        idx            = idx
      }
    ]
  ])

  subnet_import_blocks = [
    for s in local.subnet_chunks :
    <<-EOT
    import {
      to = wallarm_denylist.import_subnet_${s.exp}${s.needs_chunking ? "_${s.idx}" : ""}
      id = "${local.client_id}/subnet/${s.exp}${s.needs_chunking ? "/${s.idx}" : ""}"
    }
    EOT
  ]
}

resource "local_file" "deny_imports" {
  filename = "./wallarm_denylist_imports.tf"
  content  = join("\n", concat(local.grouped_blocks, local.subnet_import_blocks))
}
```

**2. Apply to generate the import file:**

```
$ terraform apply
```

**3. Copy the generated import file to your target configuration directory, then generate resource configs:**

```
$ terraform plan -generate-config-out=generated.tf
```

Terraform reads the import blocks and writes matching resource configurations into `generated.tf`. For example, the generated output might look like:

```hcl
# generated.tf

resource "wallarm_denylist" "import_subnet_1804809600" {
  ip_range    = ["1.1.1.1", "2.2.2.0/24", "3.3.3.3"]
  reason      = "Blocked IPs"
  time_format = "RFC3339"
  time        = "2027-03-15T00:00:00+00:00"
}

resource "wallarm_denylist" "import_location_52000393" {
  country     = ["CN", "RU"]
  reason      = "Block by country"
  time_format = "Forever"
}
```

**4. Review the generated resources** — adjust names, reasons, or grouping as needed, then apply:

```
$ terraform apply
```

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

When writing IP list resources (manually or after import), follow these grouping guidelines:

- **One entry type per resource** — `ip_range`, `country`, `datacenter`, and `proxy_type` are mutually exclusive
- **Group subnets by expiration** — IPs with the same lifetime naturally belong in one resource
- **Stay within the 1000 IP limit** — split large sets into multiple resources
- **Separate by application scope** — IPs scoped to different applications should be separate resources

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
