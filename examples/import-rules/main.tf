# Import Rules Helper
#
# Generates import blocks for all existing Wallarm rules,
# allowing you to bring them under Terraform management.
#
# Usage:
#
#   1. terraform init && terraform apply
#      -> Only creates the data source in state (reads rules from API)
#
#   2. terraform output -raw import_blocks > wallarm_rule_imports.tf
#      -> Writes import blocks to a file (nothing stored in state)
#
#   3. terraform plan -generate-config-out=imported_rules.tf
#      -> Terraform generates resource configs matching each import block
#
#   4. Post-process the generated configs to apply Terraform defaults:
#      eval "$(terraform output -raw post_process_command)"
#      -> Removes null values, client_id, mitigation, variativity_disabled lines
#         (schema defaults apply: variativity_disabled=true, comment="Managed by Terraform")
#      -> Replaces empty comments with "Managed by Terraform"
#
#   5. terraform apply
#      -> Imports all rules and updates variativity_disabled + comment in the API
#
#   6. Move the generated resource blocks to your main config,
#      remove this helper and wallarm_rule_imports.tf
#
# To import only specific rule types, set the type variable:
#   terraform apply -var='rule_types=["wallarm_mode", "rate_limit"]'

terraform {
  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
    }
  }
}

provider "wallarm" {
  api_token = var.api_token
  api_host  = var.api_host
}

variable "api_token" {
  type      = string
  sensitive = true
}

variable "api_host" {
  type    = string
  default = "https://us1.api.wallarm.com"
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

variable "resources_file" {
  type        = string
  default     = "imported_rules.tf"
  description = "Path to the generated resources file for post-processing."
}

data "wallarm_rules" "all" {
  client_id = var.client_id
  type      = var.rule_types
}

output "import_blocks" {
  value = join("\n", [
    for rule in data.wallarm_rules.all.rules :
    "import {\n  to = ${rule.terraform_resource}.rule_${rule.rule_id}\n  id = \"${rule.import_id}\"\n}"
  ])
  description = "Import blocks for all existing rules"
}

output "rule_count" {
  value       = length(data.wallarm_rules.all.rules)
  description = "Number of rules to import"
}

output "post_process_command" {
  value       = join(" && ", [
    "sed -E '/((=[[:space:]]*null)|((client_id|mitigation|variativity_disabled)[[:space:]]*=))/d' ${var.resources_file} > ${var.resources_file}.tmp",
    "mv ${var.resources_file}.tmp ${var.resources_file}",
  ])
  description = "Command to post-process generated configs (works on macOS and Linux)"
}
