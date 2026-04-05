# Hits to Rules
#
# Creates false positive suppression rules from Wallarm hit data.
# Per-request data is cached in terraform_data with ignore_changes.
# Deduplication by action_hash happens in locals (Map 2).
#
# How it works:
#   - wallarm_hits_index tracks request_ids (gating)
#   - data.wallarm_hits fetches ONLY for new (uncached) request_ids
#   - terraform_data.cache stores aggregated data per request_id (Map 1)
#   - Locals build a deduplicated map by action_hash (Map 2) and feed rules
#
# First-time setup:
#   1. Add request_ids to terraform.tfvars
#   2. terraform apply                    — fetches hits, creates rules
#
# Add more request_ids later — just add to tfvars and apply.
# Only new entries trigger API calls.
#
# Per-request config via JSON:
#   request_ids = {
#     "abc123" = "{}"                                          # defaults
#     "def456" = "{\"mode\":\"attack\"}"                       # attack mode
#     "ghi789" = "{\"rule_types\":[\"disable_stamp\"]}"        # stamp rules only
#     "jkl012" = "{\"attack_types\":[\"sqli\"]}"               # only sqli hits
#     "mno345" = "{\"mode\":\"attack\", \"attack_types\":[\"xss\",\"rce\"]}"
#   }

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

variable "request_ids" {
  type        = map(string)
  default     = {}
  description = "Map of request_id → config JSON. Empty object {} = defaults. Leave empty on first apply to initialize state."
}

variable "default_mode" {
  type    = string
  default = "request"
}

variable "include_instance" {
  type        = bool
  default     = true
  description = "Include instance (pool ID) in action conditions. Set to false if your account excludes instance from actions."
}

variable "generate_configs" {
  type        = bool
  default     = false
  description = "Generate HCL config files on disk for reference or migration."
}

variable "output_dir" {
  type    = string
  default = "./generated_rules"
}

# ─── Hits index (gating) ───────────────────────────────────────────────────

resource "wallarm_hits_index" "this" {
  client_id   = var.client_id
  request_ids = keys(var.request_ids)
}

# ─── Detect new request_ids ─────────────────────────────────────────────────

locals {
  _request_ids_to_fetch = wallarm_hits_index.this.ready ? toset([
    for id in keys(var.request_ids) : id
    if !contains(wallarm_hits_index.this.cached_request_ids, id)
  ]) : toset(keys(var.request_ids))

  _request_configs = {
    for id, cfg_json in var.request_ids :
    id => jsondecode(cfg_json)
  }
}

# ─── Fetch hits ONLY for new request_ids ────────────────────────────────────

data "wallarm_hits" "new" {
  for_each         = local._request_ids_to_fetch
  client_id        = var.client_id
  request_id       = each.key
  mode             = try(local._request_configs[each.key].mode, var.default_mode)
  attack_types     = try(local._request_configs[each.key].attack_types, [])
  rule_types       = try(local._request_configs[each.key].rule_types, [])
  include_instance = var.include_instance
}

# ─── Map 1: Per-request cache (terraform_data, ignore_changes) ─────────────

resource "terraform_data" "cache" {
  for_each = var.request_ids
  input    = try(data.wallarm_hits.new[each.key].aggregated, null)
  lifecycle { ignore_changes = [input] }
}

# ─── Map 2: Deduplicated by action_hash (built in locals) ──────────────────

locals {
  # Decode each cached entry. Skip entries with no data (null input).
  _cached_data = {
    for req_id, td in terraform_data.cache :
    req_id => try(jsondecode(td.input), null)
    if td.input != null && try(jsondecode(td.input).action_hash, "") != ""
  }

  # Action map: action_hash → action conditions (deduplicated naturally).
  _actions = {
    for req_id, data in local._cached_data :
    data.action_hash => data.action...
  }
  actions = { for ah, v in local._actions : ah => v[0] }

  # Merge groups across request_ids by action_hash.
  # Groups from different request_ids with the same action are combined.
  _groups_by_action = {
    for req_id, data in local._cached_data :
    data.action_hash => data.groups...
  }

  # Flatten and deduplicate groups within each action_hash.
  # Stamp groups: merge stamps for same point key.
  # Attack type groups: deduplicate by key (point + attack_type).
  _merged_groups = {
    for ah, group_lists in local._groups_by_action :
    ah => {
      # Flatten all group lists into one.
      groups = { for g in flatten(group_lists) : g.key => g... }
    }
  }

  # Build final deduplicated group map: action_hash_group_key → group data.
  _groups = merge([
    for ah, mg in local._merged_groups : {
      for gk, g_list in mg.groups :
      "${ah}_${gk}" => {
        action      = local.actions[ah]
        point       = g_list[0].point
        attack_type = try(g_list[0].attack_type, "")
        stamps      = distinct(flatten([for g in g_list : try(g.stamps, [])]))
      }
    }
  ]...)

  # Expand stamps: one rule per stamp group per stamp.
  stamp_rules = merge([
    for gk, g in local._groups : {
      for s in g.stamps :
      "${gk}_${s}" => {
        stamp  = s
        point  = g.point
        action = g.action
      }
    }
    if length(g.stamps) > 0
  ]...)

  # Attack type rules: groups that have an attack_type set.
  attack_type_rules = {
    for gk, g in local._groups :
    gk => {
      attack_type = g.attack_type
      point       = g.point
      action      = g.action
    }
    if g.attack_type != ""
  }

  # Flat list for generator and counts.
  _all_rules = concat(
    [for k, v in local.stamp_rules : merge(v, { key = k, resource_type = "wallarm_rule_disable_stamp", attack_type = "" })],
    [for k, v in local.attack_type_rules : merge(v, { key = k, resource_type = "wallarm_rule_disable_attack_type", stamp = 0 })],
  )
}

# ─── Create rules ──────────────────────────────────────────────────────────

resource "wallarm_rule_disable_stamp" "this" {
  for_each             = local.stamp_rules
  client_id            = var.client_id
  comment              = "Managed by Terraform"
  variativity_disabled = true
  stamp                = each.value.stamp
  point                = each.value.point

  dynamic "action" {
    for_each = each.value.action
    content {
      type  = action.value.type
      value = action.value.value
      point = action.value.point
    }
  }
}

resource "wallarm_rule_disable_attack_type" "this" {
  for_each             = local.attack_type_rules
  client_id            = var.client_id
  comment              = "Managed by Terraform"
  variativity_disabled = true
  attack_type          = each.value.attack_type
  point                = each.value.point

  dynamic "action" {
    for_each = each.value.action
    content {
      type  = action.value.type
      value = action.value.value
      point = action.value.point
    }
  }
}

# ─── Optional: generate HCL configs ────────────────────────────────────────

resource "wallarm_rule_generator" "configs" {
  count      = var.generate_configs && length(local._all_rules) > 0 ? 1 : 0
  source     = "rules"
  client_id  = var.client_id
  moved_from = "this"
  split      = true
  rules_json = jsonencode(local._all_rules)
  output_dir = var.output_dir
}

# ─── Outputs ────────────────────────────────────────────────────────────────

output "rules_created" {
  value = {
    disable_stamp       = length(local.stamp_rules)
    disable_attack_type = length(local.attack_type_rules)
    total               = length(local._all_rules)
  }
}

output "rule_ids" {
  value = merge(
    { for k, v in wallarm_rule_disable_stamp.this : k => v.rule_id },
    { for k, v in wallarm_rule_disable_attack_type.this : k => v.rule_id },
  )
}
