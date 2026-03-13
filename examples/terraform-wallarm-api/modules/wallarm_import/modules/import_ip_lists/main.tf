variable "client_id" {
  type    = number
  default = null
}

variable "is_importing" {
  type = bool
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
# Grouped types (country/datacenter/proxy): import by group ID
# Subnets: import by expired_at (merges all IPs with same expiry into one resource)

locals {
  list_configs = {
    denylist  = { entries = try(data.wallarm_ip_lists.denylist[0].entries, []),  resource = "wallarm_denylist" }
    allowlist = { entries = try(data.wallarm_ip_lists.allowlist[0].entries, []), resource = "wallarm_allowlist" }
    graylist  = { entries = try(data.wallarm_ip_lists.graylist[0].entries, []),  resource = "wallarm_graylist" }
  }

  import_blocks = {
    for list_name, cfg in local.list_configs : list_name => join("\n", concat(
      # Grouped types: one import per API group.
      [
        for e in cfg.entries :
        "import {\n  to = ${cfg.resource}.import_${e.rule_type}_${e.id}\n  id = \"${var.client_id}/${e.id}\"\n}"
        if contains(["location", "datacenter", "proxy_type"], e.rule_type)
      ],

      # Subnets: one import per unique expired_at (merges all IPs with same expiry).
      [
        for exp_at in distinct([for e in cfg.entries : e.expired_at if e.rule_type == "subnet"]) :
        "import {\n  to = ${cfg.resource}.import_subnet_${exp_at}\n  id = \"${var.client_id}/subnet/${exp_at}\"\n}"
      ],
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
