# Import Wallarm Rules into Terraform. See docs/guides/rules_import.md and
# the Makefile in this directory for shorthand commands.

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
  description = "Activate rules import."
}

variable "rule_types" {
  type        = list(string)
  default     = []
  description = "Include only these API rule types. Empty = all."
}

variable "exclude_rule_types" {
  type        = list(string)
  default     = []
  description = "Exclude these API rule types."
}

variable "sync_status" {
  type        = bool
  default     = false
  description = "Output sync status (API vs state)."
}

variable "filter_rules_in_state" {
  type        = bool
  default     = true
  description = "Skip rules whose rule_id is already in state. Set false to rebuild from scratch."
}

variable "generate_configs" {
  type        = bool
  default     = false
  description = "Generate stamp configs via wallarm_rule_generator (fallback for complex points)."
}


# State lookup — shared by sync_status output and import-block filtering.

locals {
  needs_state_lookup = var.sync_status || (var.import_rules && var.filter_rules_in_state)
}

data "external" "sync_check" {
  count = local.needs_state_lookup ? 1 : 0
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
  state_rule_ids = local.needs_state_lookup ? toset(compact(split(",",
    try(data.external.sync_check[0].result.ids, "")
  ))) : toset([])

  api_rule_ids = var.sync_status ? toset([
    for r in data.wallarm_rules.all[0].rules : tostring(r.rule_id)
    if !contains(var.exclude_rule_types, r.type)
  ]) : toset([])

  unmanaged_ids = setsubtract(local.api_rule_ids, local.state_rule_ids)
}


# Fetch rules from API.

data "wallarm_rules" "all" {
  type  = var.rule_types
  count = var.import_rules || var.sync_status ? 1 : 0
}


# Generate import blocks. Filters out rule_ids already in state when
# filter_rules_in_state = true (default).

locals {
  import_blocks = var.import_rules ? join("\n", [
    for rule in data.wallarm_rules.all[0].rules :
    "import {\n  to = ${rule.terraform_resource}.rule_${rule.rule_id}\n  id = \"${rule.import_id}\"\n}"
    if !contains(var.exclude_rule_types, rule.type)
    && !(var.filter_rules_in_state && contains(local.state_rule_ids, tostring(rule.rule_id)))
  ]) : null
}

resource "local_file" "imports" {
  filename        = "${path.root}/import_rule_blocks.tf"
  content         = local.import_blocks
  file_permission = "0644"
  count           = var.import_rules ? 1 : 0
}

output "import_blocks" {
  value       = local.import_blocks
  description = "Import blocks for existing rules."
}


# Generator fallback for stamps with complex point values.

resource "wallarm_rule_generator" "from_api" {
  count           = var.generate_configs ? 1 : 0
  source          = "api"
  output_dir      = "./"
  output_filename = "import_rule_stamp_configs.tf"
  rule_types      = ["disable_stamp"]
  split           = false
}


output "sync_status" {
  value = var.sync_status ? {
    total_in_api = length(local.api_rule_ids)
    in_state     = length(local.state_rule_ids)
    unmanaged    = length(local.unmanaged_ids)
  } : null
}
