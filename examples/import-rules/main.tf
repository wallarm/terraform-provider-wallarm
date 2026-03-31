# Import Wallarm Rules
#
# Imports existing Wallarm rules into Terraform. Two import cases are supported:
#
# ─── Case 1: Native import (all rules) ─────────────────────────────────────
#
# Uses Terraform's built-in -generate-config-out to create resource configs.
# Works for most rule types. Steps:
#
#   1. terraform apply -var='import_rules=true'
#      -> Fetches rules from API, writes import blocks (local_file + output)
#   2. terraform plan -generate-config-out=import_rule_configs.tf
#      -> Terraform generates resource HCL from import blocks
#   3. Fix generated configs (set defaults, remove nulls):
#      sed -E \
#        -e 's/(variativity_disabled[[:space:]]*)=[[:space:]]*false/\1= true/' \
#        -e 's/(comment[[:space:]]*)=[[:space:]]*null/\1= "Managed by Terraform"/' \
#        -e '/=[[:space:]]*null/d' \
#        import_rule_configs.tf > import_rule_configs.tf.tmp && \
#        mv import_rule_configs.tf.tmp import_rule_configs.tf
#   4. terraform apply
#      -> Imports all rules into state
#
# ─── Case 2: Native import + custom generator for stamps ───────────────────
#
# For disable_stamp rules with complex point values (e.g. XML namespace URIs),
# -generate-config-out may fail to generate the point field. Use the custom
# wallarm_rule_generator as a fallback for stamps, native import for the rest.
#
#   1. Import all rules EXCEPT stamps:
#      terraform apply -var='import_rules=true' -var='exclude_rule_types=["disable_stamp"]'
#      terraform plan -generate-config-out=import_rule_configs.tf
#      (fix configs with sed as in Case 1)
#      terraform apply
#
#   2. Generate stamp configs via wallarm_rule_generator:
#      terraform apply -var='generate_configs=true' -var='import_rules=true' -var='rule_types=["disable_stamp"]'
#      -> Writes import_rule_stamp_configs.tf (via rule_generator, handles complex points)
#      -> Writes import_rule_blocks.tf (import blocks for stamps only)
#      terraform apply
#      -> Imports stamp rules using the generated configs
#
# ─── Other operations ──────────────────────────────────────────────────────
#
# Import specific rule types only:
#   terraform apply -var='import_rules=true' -var='rule_types=["wallarm_mode"]'
#
# Check sync status (API vs state, read-only):
#   terraform plan -refresh=false -var='sync_status=true'
#
# Re-importing existing resources is safe — Terraform generates configs
# only for resources not already in state.
#
# See the Makefile for shorthand commands for all operations.

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
  default = "https://api.wallarm.com"
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
# Method 1: via local_file
#
locals {
  import_blocks = var.import_rules ? join("\n", [
    for rule in data.wallarm_rules.all[0].rules :
    "import {\n  to = ${rule.terraform_resource}.rule_${rule.rule_id}\n  id = \"${rule.import_id}\"\n}"
    if !contains(var.exclude_rule_types, rule.type)
  ]) : null
}

resource "local_file" "imports" {
  filename        = "${path.root}/import_rule_blocks.tf"
  content         = local.import_blocks
  file_permission = "0644"
  count           = var.import_rules ? 1 : 0
}

# Method 2: via output (active)

output "import_blocks" {
  value       = local.import_blocks
  description = "Import blocks for existing rules"
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

resource "wallarm_rule_generator" "from_api" {
  count           = var.generate_configs ? 1 : 0
  source          = "api"
  output_dir      = "./"
  output_filename = "import_rule_stamp_configs.tf"
  rule_types      = ["disable_stamp"]
  split           = false
}

# ─── Output Sync Status ────────────────────────────────────────────────────────────────

output "sync_status" {
  value = var.sync_status ? {
    total_in_api = length(local.api_rule_ids)
    in_state     = length(local.state_rule_ids)
    unmanaged    = length(local.unmanaged_ids)
  } : null
}
