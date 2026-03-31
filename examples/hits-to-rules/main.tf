# Hits to Rules
#
# Creates false positive suppression rules from Wallarm hit data.
# Single terraform apply — state-only, no filesystem dependency.
#
# How it works:
#   - wallarm_hits_index tracks which request_ids are cached
#   - data.wallarm_hits fetches ONLY for new (uncached) request_ids
#   - terraform_data.rules_cache persists rules per request_id (ignore_changes)
#   - After first apply, cached_request_ids matches request_ids → no more fetches
#   - Rules survive in state even after hits expire from the API
#
# Usage:
#   1. Add request_ids to terraform.tfvars
#   2. terraform apply
#
# Add more request_ids later — just add to tfvars and apply.
# Only new entries trigger API calls.
#
# Generate HCL config files (optional, for reference or migration):
#   terraform apply -var='generate_configs=true'
#
# Per-request config via JSON:
#   request_ids = {
#     "abc123" = "{}"                                       # defaults
#     "def456" = "{\"mode\":\"attack\"}"                    # attack mode
#     "ghi789" = "{\"rule_types\":[\"disable_stamp\"]}"     # stamp only
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
  description = "Map of request_id → config JSON. Empty object {} = defaults."
}

variable "default_mode" {
  type    = string
  default = "request"
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

# ─── Hits index (tracks cached request_ids) ─────────────────────────────────

resource "wallarm_hits_index" "this" {
  client_id   = var.client_id
  request_ids = keys(var.request_ids)
}

# ─── Detect new request_ids ─────────────────────────────────────────────────

locals {
  _cached = toset(compact(split(",", wallarm_hits_index.this.cached_request_ids)))

  _new_request_ids = toset([
    for id in keys(var.request_ids) : id
    if !contains(local._cached, id)
  ])

  _request_configs = {
    for id, cfg_json in var.request_ids :
    id => jsondecode(cfg_json)
  }
}

# ─── Fetch hits ONLY for new request_ids ────────────────────────────────────

data "wallarm_hits" "new" {
  for_each   = local._new_request_ids
  client_id  = var.client_id
  request_id = each.key
  mode       = try(local._request_configs[each.key].mode, var.default_mode)
}

# ─── Persist rules in state (hits are ephemeral) ──────────────────────────

resource "terraform_data" "rules_cache" {
  for_each = var.request_ids
  input = try(jsonencode([
    for rule in data.wallarm_hits.new[each.key].rules : merge(rule, {
      key = "${substr(each.key, 0, 8)}_${rule.key}"
    })
  ]), null)
  lifecycle {
    ignore_changes = [input]
  }
}

# ─── Build rule maps ────────────────────────────────────────────────────────

locals {
  _all_rules = flatten([
    for req_id in keys(var.request_ids) :
    try(jsondecode(terraform_data.rules_cache[req_id].input), [])
  ])

  stamp_rules = {
    for rule in local._all_rules : rule.key => rule
    if rule.resource_type == "wallarm_rule_disable_stamp"
  }

  attack_type_rules = {
    for rule in local._all_rules : rule.key => rule
    if rule.resource_type == "wallarm_rule_disable_attack_type"
  }
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
  count     = var.generate_configs && length(local._all_rules) > 0 ? 1 : 0
  source    = "hits"
  client_id = var.client_id
  requests_json = jsonencode({
    for req_id in keys(var.request_ids) : req_id => {
      hits              = try(jsonencode(data.wallarm_hits.new[req_id].hits), "[]")
      action_conditions = try(jsonencode(data.wallarm_hits.new[req_id].action_conditions), "[]")
    }
  })
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
