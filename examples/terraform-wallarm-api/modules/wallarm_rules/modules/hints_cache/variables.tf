variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "refresh" {
  type        = bool
  default     = false
  description = "When true, refetch all rules from API and rebuild the index."
}

variable "rule_types" {
  type        = list(string)
  default     = []
  description = "Optional filter by API rule type(s). Empty = all types."
}

variable "configs_dir" {
  type        = string
  description = "Top-level configs directory for action subdirectories."
}

variable "rules_engine_address" {
  type        = string
  default     = ""
  description = "Terraform address prefix for rule resources in import blocks."
}
