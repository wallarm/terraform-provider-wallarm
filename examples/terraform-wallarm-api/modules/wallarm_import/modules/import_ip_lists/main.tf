variable "client_id" {
  type    = number
  default = null
}

variable "is_importing" {
  type = bool
}

variable "subnet_import_mode" {
  type        = string
  default     = "grouped"
  description = "How to import subnet entries: 'grouped' merges by expired_at into resources of max 1000 IPs; 'individual' creates one resource per IP."

  validation {
    condition     = contains(["grouped", "individual"], var.subnet_import_mode)
    error_message = "subnet_import_mode must be 'grouped' or 'individual'."
  }
}

# ─── Fetch all IP list entries ────────────────────────────────────────────────

data "wallarm_ip_lists" "denylist" {
  count     = var.is_importing ? 1 : 0
  client_id = var.client_id
  list_type = "denylist"
}

data "wallarm_ip_lists" "allowlist" {
  count     = var.is_importing ? 1 : 0
  client_id = var.client_id
  list_type = "allowlist"
}

data "wallarm_ip_lists" "graylist" {
  count     = var.is_importing ? 1 : 0
  client_id = var.client_id
  list_type = "graylist"
}

# ─── Generate import blocks ──────────────────────────────────────────────────

locals {
  max_subnets_per_resource = 1000

  list_configs = {
    denylist  = { entries = try(data.wallarm_ip_lists.denylist[0].entries, []),  resource = "wallarm_denylist" }
    allowlist = { entries = try(data.wallarm_ip_lists.allowlist[0].entries, []), resource = "wallarm_allowlist" }
    graylist  = { entries = try(data.wallarm_ip_lists.graylist[0].entries, []),  resource = "wallarm_graylist" }
  }

  # ── Individual mode: one import block per subnet group ─────────────────────
  individual_subnet_blocks = {
    for list_name, cfg in local.list_configs : list_name => [
      for e in cfg.entries :
      "import {\n  to = ${cfg.resource}.import_subnet_${e.id}\n  id = \"${var.client_id}/${e.id}\"\n}"
      if e.rule_type == "subnet"
    ]
  }

  # ── Grouped mode: merge subnets by expired_at, chunk into max 1000 ─────────
  # Step 1: collect all subnet entries per list, keyed by expired_at.
  subnets_by_expiry = {
    for list_name, cfg in local.list_configs : list_name => {
      for exp_at in distinct([for e in cfg.entries : e.expired_at if e.rule_type == "subnet"]) :
      exp_at => [for e in cfg.entries : e if e.rule_type == "subnet" && e.expired_at == exp_at]
    }
  }

  # Step 2: chunk each expired_at group into batches of max_subnets_per_resource.
  # Each chunk becomes one import block with id = "{clientID}/subnet/{expiredAt}".
  # For chunks beyond the first, we append a chunk index to the resource name.
  grouped_subnet_blocks = {
    for list_name, cfg in local.list_configs : list_name => flatten([
      for exp_at, entries in try(local.subnets_by_expiry[list_name], {}) : [
        for chunk_idx in range(ceil(length(entries) / local.max_subnets_per_resource)) :
        "import {\n  to = ${cfg.resource}.import_subnet_${exp_at}${chunk_idx > 0 ? "_${chunk_idx}" : ""}\n  id = \"${var.client_id}/subnet/${exp_at}${chunk_idx > 0 ? "/${chunk_idx}" : ""}\"\n}"
      ]
    ])
  }

  # ── Select mode ────────────────────────────────────────────────────────────
  subnet_blocks = var.subnet_import_mode == "individual" ? local.individual_subnet_blocks : local.grouped_subnet_blocks

  import_blocks = {
    for list_name, cfg in local.list_configs : list_name => join("\n", concat(
      # Grouped types: one import per API group.
      [
        for e in cfg.entries :
        "import {\n  to = ${cfg.resource}.import_${e.rule_type}_${e.id}\n  id = \"${var.client_id}/${e.id}\"\n}"
        if contains(["location", "datacenter", "proxy_type"], e.rule_type)
      ],

      # Subnets: mode-dependent (individual or grouped).
      try(local.subnet_blocks[list_name], []),
    ))
    if length(cfg.entries) > 0
  }
}

# ─── Write import blocks to files ─────────────────────────────────────────────

resource "local_file" "ip_list_imports" {
  for_each        = var.is_importing ? local.import_blocks : {}
  filename        = "${path.root}/wallarm_${each.key}_imports.tf"
  content         = each.value
  file_permission = "0644"
}

# ─── Outputs ──────────────────────────────────────────────────────────────────

output "denylist_entries" {
  value = try(data.wallarm_ip_lists.denylist[0].entries, [])
}

output "allowlist_entries" {
  value = try(data.wallarm_ip_lists.allowlist[0].entries, [])
}

output "graylist_entries" {
  value = try(data.wallarm_ip_lists.graylist[0].entries, [])
}

output "entry_counts" {
  value = {
    denylist  = length(try(data.wallarm_ip_lists.denylist[0].entries, []))
    allowlist = length(try(data.wallarm_ip_lists.allowlist[0].entries, []))
    graylist  = length(try(data.wallarm_ip_lists.graylist[0].entries, []))
  }
}
