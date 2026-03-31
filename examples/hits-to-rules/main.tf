# Hits to Rules
#
# Creates false positive suppression rules from Wallarm hit data.
# Single terraform apply — no filesystem dependency, state-only.
#
# Usage:
#   1. Add request_ids to terraform.tfvars
#   2. terraform apply
#
# With attack mode (expand to related hits by attack_id):
#   terraform apply -var='mode=attack'
#
# Generate HCL config files (optional, for reference or migration):
#   terraform apply -var='generate_configs=true'
#
# Example request_ids in terraform.tfvars:
#   request_ids = {
#     "abc123def456" = []                   # all rule types
#     "789ghi012jkl" = ["disable_stamp"]    # stamp rules only
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
  type        = map(list(string))
  description = "Map of request_id → rule_types filter. Empty list = all types (disable_stamp + disable_attack_type)."
}

variable "mode" {
  type        = string
  default     = "request"
  description = "Fetch mode: 'request' (direct hits) or 'attack' (expand to related hits by attack_id)."
}

variable "generate_configs" {
  type        = bool
  default     = false
  description = "Generate HCL config files on disk for reference or migration."
}

variable "output_dir" {
  type        = string
  default     = "./generated_rules"
  description = "Output directory for generated HCL configs."
}

# ─── Fetch hits ─────────────────────────────────────────────────────────────

data "wallarm_hits" "this" {
  for_each   = var.request_ids
  client_id  = var.client_id
  request_id = each.key
  mode       = var.mode
}

# ─── Build rule maps ────────────────────────────────────────────────────────

locals {
  _rules_per_request = {
    for req_id, filter in var.request_ids :
    req_id => [
      for rule in data.wallarm_hits.this[req_id].rules : {
        key           = "${req_id}_${rule.key}"
        resource_type = rule.resource_type
        stamp         = rule.stamp
        attack_type   = rule.attack_type
        point         = rule.point
        action        = rule.action
      }
      if length(filter) == 0 || contains(filter, replace(rule.resource_type, "wallarm_rule_", ""))
    ]
  }

  all_rules = flatten(values(local._rules_per_request))

  stamp_rules = {
    for r in local.all_rules : r.key => r
    if r.resource_type == "wallarm_rule_disable_stamp"
  }

  attack_type_rules = {
    for r in local.all_rules : r.key => r
    if r.resource_type == "wallarm_rule_disable_attack_type"
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
  count  = var.generate_configs ? 1 : 0
  source = "hits"
  requests_json = jsonencode({
    for req_id, filter in var.request_ids : req_id => {
      hits              = jsonencode(data.wallarm_hits.this[req_id].hits)
      action_conditions = jsonencode(data.wallarm_hits.this[req_id].action_conditions)
      rule_types        = filter
    }
  })
  output_dir = var.output_dir
}

# ─── Outputs ────────────────────────────────────────────────────────────────

output "rules_created" {
  value = {
    disable_stamp       = length(local.stamp_rules)
    disable_attack_type = length(local.attack_type_rules)
    total               = length(local.all_rules)
  }
}

output "rule_ids" {
  value = merge(
    { for k, v in wallarm_rule_disable_stamp.this : k => v.rule_id },
    { for k, v in wallarm_rule_disable_attack_type.this : k => v.rule_id },
  )
}
