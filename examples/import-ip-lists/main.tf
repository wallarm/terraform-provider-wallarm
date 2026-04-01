# Import IP Lists
#
# Imports all existing Wallarm IP list entries into Terraform state.
# Handles all entry types: subnets (IPs), countries, datacenters, proxy types.
# Groups subnets by expiration + application scope, with chunking for >1000 IPs.
#
# Usage:
#   1. terraform apply
#      → Fetches all IP list entries, writes import blocks to ./imports/
#   2. Copy generated files to your target config directory
#   3. terraform plan -generate-config-out=generated.tf
#      → Terraform generates resource configs from import blocks
#   4. Review generated resources, then terraform apply
#
# To import a single list type:
#   terraform apply -var='list_types=["denylist"]'

terraform {
  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
    }
  }
}

variable "api_token" {
  type      = string
  sensitive = true
}

variable "api_host" {
  type    = string
  default = "https://us1.api.wallarm.com"
}

provider "wallarm" {
  api_token = var.api_token
  api_host  = var.api_host
}

variable "client_id" {
  type    = number
  default = null
}

variable "list_types" {
  type        = list(string)
  default     = ["denylist", "allowlist", "graylist"]
  description = "IP list types to import."
}

variable "output_dir" {
  type    = string
  default = "./imports"
}

variable "max_subnets_per_resource" {
  type    = number
  default = 1000
}

# ─── Fetch all IP list entries ───────────────────────────────────────────────

data "wallarm_ip_lists" "this" {
  for_each  = toset(var.list_types)
  client_id = var.client_id
  list_type = each.key
}

# ─── Build import blocks ────────────────────────────────────────────────────

locals {
  # Map list_type display names to resource names.
  resource_names = {
    denylist  = "wallarm_denylist"
    allowlist = "wallarm_allowlist"
    graylist  = "wallarm_graylist"
  }

  # Resolve client_id for import IDs.
  client_id = var.client_id != null ? var.client_id : try(
    [for k, v in data.wallarm_ip_lists.this : v.id][0], 0
  )

  # ── Grouped types (country/datacenter/proxy): one import per API group ──

  grouped_entries = flatten([
    for list_type in var.list_types : [
      for e in data.wallarm_ip_lists.this[list_type].entries : {
        list_type     = list_type
        resource_name = local.resource_names[list_type]
        rule_type     = e.rule_type
        id            = e.id
      }
      if contains(["location", "datacenter", "proxy_type"], e.rule_type)
    ]
  ])

  grouped_blocks = [
    for e in local.grouped_entries :
    <<-EOT
import {
  to = ${e.resource_name}.import_${e.rule_type}_${e.id}
  id = "${local.client_id}/${e.id}"
}
EOT
  ]

  # ── Subnets: group by (list_type, expired_at, app_ids), chunk if >1000 ──

  subnet_entries = flatten([
    for list_type in var.list_types : [
      for e in data.wallarm_ip_lists.this[list_type].entries : {
        list_type     = list_type
        resource_name = local.resource_names[list_type]
        expired_at    = e.expired_at
        app_key       = length(e.application_ids) == 0 ? "all" : join(",", sort([for id in e.application_ids : tostring(id)]))
        values        = e.values
      }
      if e.rule_type == "subnet"
    ]
  ])

  # Unique grouping keys: list_type/expired_at/app_key
  subnet_group_keys = distinct([
    for e in local.subnet_entries : "${e.list_type}/${e.expired_at}/${e.app_key}"
  ])

  # Group entries by composite key.
  subnets_by_scope = {
    for key in local.subnet_group_keys :
    key => [for e in local.subnet_entries if "${e.list_type}/${e.expired_at}/${e.app_key}" == key : e]
  }

  # Build chunks (max 1000 subnets per resource).
  subnet_chunks = flatten([
    for key, entries in local.subnets_by_scope : [
      for idx in range(ceil(length(entries) / var.max_subnets_per_resource)) : {
        list_type      = entries[0].list_type
        resource_name  = entries[0].resource_name
        expired_at     = entries[0].expired_at
        app_key        = entries[0].app_key
        idx            = idx
        needs_chunking = length(entries) > var.max_subnets_per_resource
      }
    ]
  ])

  subnet_blocks = [
    for s in local.subnet_chunks :
    <<-EOT
import {
  to = ${s.resource_name}.import_subnet_${s.expired_at}_${replace(s.app_key, ",", "_")}${s.needs_chunking ? "_${s.idx}" : ""}
  id = "${local.client_id}/subnet/${s.expired_at}/apps/${s.app_key}${s.needs_chunking ? "/${s.idx}" : ""}"
}
EOT
  ]

  # ── Combine all import blocks per list type ──

  all_blocks = concat(local.grouped_blocks, local.subnet_blocks)
}

# ─── Write import files ─────────────────────────────────────────────────────

resource "local_file" "imports" {
  for_each = toset(var.list_types)
  filename = "${var.output_dir}/${each.key}_imports.tf"
  content = join("\n", [
    for b in local.all_blocks : b
    if length(regexall("${local.resource_names[each.key]}\\.", b)) > 0
  ])
}

# ─── Outputs ────────────────────────────────────────────────────────────────

output "summary" {
  value = {
    for list_type in var.list_types : list_type => {
      total_entries = length(data.wallarm_ip_lists.this[list_type].entries)
      grouped       = length([for e in local.grouped_entries if e.list_type == list_type])
      subnet_chunks = length([for s in local.subnet_chunks if s.list_type == list_type])
      import_file   = "${var.output_dir}/${list_type}_imports.tf"
    }
  }
}

output "import_blocks_count" {
  value = length(local.all_blocks)
}
