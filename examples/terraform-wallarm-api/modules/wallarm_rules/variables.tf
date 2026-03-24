variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "requests" {
  type        = map(list(string))
  default     = {}
  description = "Static map of request_id => list of rule types (for FP rules from hits)."
}

variable "hits_mode" {
  type        = string
  default     = "request"
  description = "Fetch mode for hits: 'request' (direct hits only) or 'attack' (expand to all related hits by attack_id)"
}

variable "configs_dir" {
  type        = string
  description = "Top-level directory for all rule configs. Action subdirectories are created automatically."
}

variable "fetch_hits" {
  type        = bool
  default     = false
  description = "Force fetch hits from API even if YAML configs exist. Auto-fetches on first apply (no configs yet)."
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Fetch all rules from API, generate import blocks, and create resources. Single apply imports + updates defaults."
}

variable "import_rule_types" {
  type        = list(string)
  default     = []
  description = "Optional filter by API rule type(s) for import. Empty = all types."
}

variable "discover_actions" {
  type        = bool
  default     = false
  description = "Discover new action scopes from API and generate .action.yaml files."
}

variable "rules_engine_address" {
  type        = string
  default     = "module.wallarm_rules.module.rules"
  description = "Full Terraform address of the rules_engine module. Used in import blocks."
}
