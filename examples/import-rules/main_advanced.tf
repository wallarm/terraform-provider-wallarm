# Import Rules Helper
#
# Generates import blocks for all existing Wallarm rules,
# allowing you to bring them under Terraform management.
#
# Usage:
#
# Wallarm Rules Import Advanced Workflow
#
#
#   1. terraform init && terraform apply -auto-approve -var='is_importing=true'
#      -> Creates the data source in state (reads rules from API) and writes import blocks to a file
#
#   2. terraform plan -var='is_importing=true' -generate-config-out=imported_rules.tf
#      -> Terraform generates resource configs matching each import block
#
#   3. terraform apply --auto-approve
#      -> Imports all rules into state. The provider Read function automatically
#         sets variativity_disabled=true and comment="Managed by Terraform" (if empty),
#         so the first apply will also update these values in the API. Import blocks file
#         is automatically removed, because the variable is_importing is not passed.
#
# Outcome - all resources are imported, configruation file is created, state is clean (no extra state entities or files),
# terraform plan gives clean output
#
# To import only specific rule types, set the type variable:
#   terraform apply -var='rule_types=["wallarm_mode", "rate_limit"]'
#

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

variable "rule_types" {
  type        = list(string)
  default     = []
  description = "Filter by API rule type(s). Empty list means all types."
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Must be true to activate rules import functionality."
}

# # --- Fetch all non-system rules ---

data "wallarm_rules" "all" {
  client_id = var.client_id
  count     = var.is_importing ? 1 : 0
}

# # --- Fetch only specific rule types ---

# data "wallarm_rules" "stamps_and_parsers" {
#   client_id = var.client_id
#   type      = ["disable_stamp", "disable_attack_type", "parser_state"]
#   count     = var.is_importing ? 1 : 0
# }

# # --- Useful locals ---

# locals {
#   # All rules grouped by API type
#   rules_by_type = {
#     for t in distinct(data.wallarm_rules.all.rules[*].api_type) :
#     t => [for r in data.wallarm_rules.all.rules : r if r.api_type == t]
#   }

#   # Count per type
#   rule_counts = {
#     for t, rules in local.rules_by_type : t => length(rules)
#   }

#   # Only rules that have a known Terraform resource mapping
#   importable_rules = [
#     for r in data.wallarm_rules.all.rules : r
#     if r.terraform_resource_type != ""
#   ]
# }

locals {
  # Build import blocks for all fetched/filtered rules
  import_blocks = var.is_importing ? join("\n\n", [
    for r in data.wallarm_rules.all[0].rules :
    <<-EOT
    import {
      to = ${r.terraform_resource}.imported["${r.rule_id}"]
      id = "${r.import_id}"
    }
    EOT
  ]) : null
}

# --- Generate import blocks ---
# Uncomment to write a file with import blocks for all importable rules.

resource "local_file" "imports" {
  depends_on      = [data.wallarm_rules.all]
  filename        = "${path.root}/wallarm_rule_imports.tf"
  content         = local.import_blocks
  file_permission = "0644"
  count           = var.is_importing ? 1 : 0
}

# --- Outputs ---

output "total_rules" {
  count = var.is_importing ? 1 : 0
  value = length(data.wallarm_rules.all[0].rules)
}

output "rule_counts_by_type" {
  count = var.is_importing ? 1 : 0
  value = local.rule_counts
}

# output "all_rules" {
#   value = data.wallarm_rules.all[0].rules
# }

# output "importable_rules" {
#   description = "Rules with known Terraform resource types, ready for import"
#   value = [
#     for r in local.importable_rules : {
#       import_id     = r.import_id
#       resource_type = r.terraform_resource_type
#       api_type      = r.api_type
#       comment       = r.comment
#     }
#   ]
# }
