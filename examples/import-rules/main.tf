# Import Wallarm Rules and Sync Status Check
#
# Imports existing Wallarm rules into Terraform. Generated resource configs
# and state live in this directory.
#
# Import all rules:
#   1. terraform apply -var='import_rules=true'
#   2. terraform output -raw import_blocks > import_rule_blocks.tf
#   3. terraform plan -generate-config-out=import_rule_configs.tf
#   4. terraform apply
#
# Import via local_file (alternative — uncomment Method 1, comment Method 2):
#   1. terraform apply -var='import_rules=true'
#      -> Writes import blocks via local_file resource
#   2. terraform plan -generate-config-out=import_rule_configs.tf
#   3. terraform apply
#
# Import specific rule types:
#   terraform apply -var='import_rules=true' -var='rule_types=["wallarm_mode","disable_stamp"]'
#
# Import all except specific types:
#   terraform apply -var='import_rules=true' -var='exclude_rule_types=["disable_stamp"]'
#
# Check sync status (API vs state, no changes made):
#   terraform apply -var='sync_status=true'
#
# Generate HCL configs for rules with complex points (fallback):
#   terraform apply -var='generate_configs=true'
#   -> Fetches rules from API and generates correct HCL with properly escaped
#      point values. Use when `terraform plan -generate-config-out` fails to
#      generate the point field (e.g. rules with XML namespace URIs).
#   terraform apply -var='generate_configs=true' -var='generate_rule_types=["disable_stamp","disable_attack_type"]'
# Re-importing existing resources is safe — Terraform generates configs
# only for resources not already in state.
#
# Check the Makefile for shorthand commands for all rules import operations.

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

variable "import_rules" {
  type        = bool
  default     = false
  description = "Must be true to activate rules import functionality."
}

variable "rule_types" {
  type        = list(string)
  default     = []
  description = "Include only these API rule type(s). Empty list means all types."
}

variable "exclude_rule_types" {
  type        = list(string)
  default     = []
  description = "Exclude these API rule type(s) from import blocks and sync status."
}

variable "sync_status" {
  type        = bool
  default     = false
  description = "Must be true to activate rules sync status (API <-> TF state)"
}


# ─── Sync status: compare API rules vs Terraform state ──────────────────────

data "external" "sync_check" {
  count = var.sync_status ? 1 : 0
  program = ["bash", "-c", <<-EOF
    ids=$(terraform show -json 2>/dev/null \
      | jq -r '.values.root_module.resources[]
        | select(.type | startswith("wallarm_rule_"))
        | .values.rule_id' \
      | sort -un \
      | paste -sd, - || echo "")
    printf '{"ids":"%s"}' "$ids"
  EOF
  ]
}

locals {
  state_rule_ids = var.sync_status ? toset(compact(split(",",
    try(data.external.sync_check[0].result.ids, "")
  ))) : toset([])

  api_rule_ids = var.sync_status ? toset([
    for r in data.wallarm_rules.all[0].rules : tostring(r.rule_id)
    if !contains(var.exclude_rule_types, r.type)
  ]) : toset([])

  unmanaged_ids = setsubtract(local.api_rule_ids, local.state_rule_ids)
}


# ─── Fetch rules from API ───────────────────────────────────────────────────

data "wallarm_rules" "all" {
  type  = var.rule_types
  count = var.import_rules || var.sync_status ? 1 : 0
}

# ─── Generate import blocks ─────────────────────────────────────────────────
#
# Method 1: via local_file (uncomment to use instead of Method 2)
#
# locals {
#   import_blocks = var.import_rules ? join("\n", [
#     for rule in data.wallarm_rules.all[0].rules :
#     "import {\n  to = ${rule.terraform_resource}.rule_${rule.rule_id}\n  id = \"${rule.import_id}\"\n}"
#     if !contains(var.exclude_rule_types, rule.type)
#   ]) : null
# }
#
# resource "local_file" "imports" {
#   filename        = "${path.root}/import_rule_blocks.tf"
#   content         = local.import_blocks
#   file_permission = "0644"
#   count           = var.import_rules ? 1 : 0
# }

# Method 2: via output (active)

output "import_blocks" {
  value = var.import_rules ? join("\n", [
    for rule in data.wallarm_rules.all[0].rules :
    "import {\n  to = ${rule.terraform_resource}.rule_${rule.rule_id}\n  id = \"${rule.import_id}\"\n}"
    if !contains(var.exclude_rule_types, rule.type)
  ]) : null
  description = "Import blocks for all existing rules"
}

# ─── Generate HCL configs via rule_generator (fallback) ───────────────────────
# Uses wallarm_rule_generator with source="api" to produce correct HCL for
# disable_stamp/disable_attack_type rules. Unlike -generate-config-out, this
# handles complex point values with special characters (XML namespaces, etc.).

variable "generate_configs" {
  type        = bool
  default     = false
  description = "Generate HCL configs via wallarm_rule_generator. Use as fallback when -generate-config-out fails on complex point values."
}

variable "generate_rule_types" {
  type        = list(string)
  default     = ["disable_stamp"]
  description = "Rule types to generate configs for. Supports: disable_stamp, disable_attack_type."
}

resource "wallarm_rule_generator" "from_api" {
  count      = var.generate_configs ? 1 : 0
  source     = "api"
  output_dir = "./"
  rule_types = var.generate_rule_types
  split      = false
}

# ─── Output Sync Status ────────────────────────────────────────────────────────────────

output "sync_status" {
  value = var.sync_status ? {
    total_in_api = length(local.api_rule_ids)
    in_state     = length(local.state_rule_ids)
    unmanaged    = length(local.unmanaged_ids)
  } : null
}
