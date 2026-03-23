variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "request_id" {
  type        = string
  description = "Request ID to fetch hits for"
}

variable "rule_types" {
  type        = list(string)
  description = "Rule types to generate. Valid values: disable_stamp, disable_attack_type"

  validation {
    condition     = alltrue([for rt in var.rule_types : contains(["disable_stamp", "disable_attack_type"], rt)])
    error_message = "Valid rule_types: disable_stamp, disable_attack_type."
  }
}

variable "config_dir" {
  type        = string
  description = "Top-level configs directory. Rules are placed in action_dir_name subdirectory."
}

variable "time" {
  type        = list(number)
  default     = []
  description = "Optional [from, to] unix timestamps. Empty list = provider defaults (6 months ago -> now)."
}

variable "fetch_hits" {
  type        = bool
  default     = false
  description = "When true, fetch hits from the Wallarm API. When false (default), read persisted data from Terraform state."
}

variable "mode" {
  type        = string
  default     = "request"
  description = "Fetch mode: 'request' fetches hits for the request_id only; 'attack' expands to all related hits by attack_id."
}

variable "attack_types" {
  type        = list(string)
  default     = []
  description = "Override allowed attack types for filtering in attack mode. Empty list uses provider defaults."
}
