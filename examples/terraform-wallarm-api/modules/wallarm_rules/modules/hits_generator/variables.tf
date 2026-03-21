variable "request_id" {
  type        = string
  description = "Original request ID from hits"
}

variable "rule_types" {
  type        = list(string)
  description = "Rule types to generate. Valid values: disable_stamp, disable_attack_type"

  validation {
    condition     = alltrue([for rt in var.rule_types : contains(["disable_stamp", "disable_attack_type"], rt)])
    error_message = "Valid rule_types: disable_stamp, disable_attack_type."
  }
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
  description = "Map of point_hash => { point_wrapped, stamps, attack_types, hit_ids }"
}

variable "action" {
  description = "Built action conditions from hits_fetcher (for reference HCL)"
}

variable "config_dir" {
  type        = string
  description = "Base directory for generated configs (request_id subdir is appended)"
}
