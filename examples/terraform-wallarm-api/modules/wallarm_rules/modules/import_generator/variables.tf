variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Set to true to fetch rules from API and generate import blocks."
}

variable "convert_imports" {
  type        = bool
  default     = false
  description = "Set to true to generate YAML configs + moved blocks for migrating imported rules to rules_engine."
}

variable "rule_types" {
  type        = list(string)
  default     = []
  description = "Optional filter by API rule type(s). Empty = all types."
}

variable "import_address_prefix" {
  type        = string
  default     = ""
  description = "Terraform address prefix for imported resources. Used in import blocks."
}

variable "rules_engine_address" {
  type        = string
  default     = "module.wallarm_rules.module.rules"
  description = "Terraform address of the rules_engine module. Used in moved blocks TO address."
}

variable "import_config_dir" {
  type        = string
  default     = ""
  description = "Directory for YAML configs generated during convert_imports."
}
