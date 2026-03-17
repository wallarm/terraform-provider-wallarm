variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "requests" {
  type        = map(list(string))
  default     = {}
  description = "Static map of request_id => list of rule types."
}

variable "hits_mode" {
  type        = string
  default     = "request"
  description = "Fetch mode for hits: 'request' (direct hits only) or 'attack' (expand to all related hits by attack_id)"
}

variable "custom_rules" {
  type        = any
  default     = []
  description = "Custom rule definitions. Passed through to custom_rules child module."
}

variable "config_dir" {
  type        = string
  description = "Directory where rule config files are written"
}

variable "fp_config_dir" {
  type        = string
  description = "Directory where false-positive rule config files are written"
}

variable "config_format" {
  type        = string
  default     = "yaml"
  description = "Config file format: 'yaml' for YAML files or 'hcl' for Terraform resource blocks"
}
