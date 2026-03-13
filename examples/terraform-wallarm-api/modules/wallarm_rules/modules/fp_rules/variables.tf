variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "request_id" {
  type        = string
  description = "Original request ID from hits — stored in metadata"
}

variable "rule_types" {
  type        = list(string)
  description = "Rule types to create. Valid values: disable_stamp, disable_attack_type"

  validation {
    condition     = alltrue([for rt in var.rule_types : contains(["disable_stamp", "disable_attack_type"], rt)])
    error_message = "Valid rule_types: disable_stamp, disable_attack_type."
  }
}


variable "action" {
  description = "Rule action conditions (TypeSet from wallarm_hits data source)"
}

variable "domain" {
  type        = string
  description = "Request domain from hit"
}

variable "path" {
  type        = string
  description = "Request path from hit"
}

variable "poolid" {
  type        = number
  description = "Application pool ID from hit"
}

variable "points" {
  type        = any
  description = "Map of point_hash => { point_wrapped, stamps, attack_types }"
}

variable "config_dir" {
  type        = string
  description = "Directory where rule config YAML files are written"
}
