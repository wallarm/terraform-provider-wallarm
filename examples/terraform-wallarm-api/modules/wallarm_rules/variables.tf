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

variable "config_dir" {
  type        = string
  description = "Directory for custom rule YAML configs"
}

variable "fp_config_dir" {
  type        = string
  description = "Directory for false-positive rule YAML configs (generated from hits)"
}

variable "import_config_dir" {
  type        = string
  description = "Directory for YAML configs converted from imported rules"
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Set to true to fetch rules from API and generate import blocks"
}

variable "convert_imports" {
  type        = bool
  default     = false
  description = "Set to true to generate YAML configs + moved blocks for migrating imported rules to rules_engine"
}

variable "import_rule_types" {
  type        = list(string)
  default     = []
  description = "Optional filter by API rule type(s) for import. Empty = all types."
}

variable "import_address_prefix" {
  type        = string
  default     = ""
  description = "Terraform address prefix for imported resources in import blocks."
}

variable "rules_engine_address" {
  type        = string
  default     = "module.wallarm_rules.module.rules"
  description = "Terraform address of the rules_engine module. Used in moved blocks."
}
