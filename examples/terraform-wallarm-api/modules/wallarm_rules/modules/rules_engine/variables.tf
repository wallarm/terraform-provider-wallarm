variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "config_dirs" {
  type        = list(string)
  description = "Directories to discover YAML config files from (recursive **/*.yaml)"
}

variable "generated_rules" {
  type = list(object({
    name          = string
    resource_type = string
    comment       = optional(string, "Managed by Terraform")
    path          = optional(string, "")
    domain        = optional(string, "")
    instance      = optional(string, "")
    method        = optional(string, "")
    scheme        = optional(string, "")
    proto         = optional(string, "")
    query         = optional(list(object({ key = string, value = string, type = optional(string, "equal") })), [])
    headers       = optional(list(object({ name = string, value = string, type = optional(string, "equal") })), [])
    point         = optional(list(list(string)), [])
    attack_types  = optional(list(string), [])
    stamps        = optional(list(number), [])
    file_types    = optional(list(string), [])
    parsers       = optional(list(string), [])
    attack_type   = optional(string, "")
    mode          = optional(string, "")
    regex         = optional(string, "")
    regex_id      = optional(number, 0)
    regex_rule    = optional(string, "")
    experimental  = optional(bool, false)
    parser        = optional(string, "")
    file_type     = optional(string, "")
    delay         = optional(number, 0)
    burst         = optional(number, 0)
    rate          = optional(number, 0)
    rsp_status    = optional(number, 0)
    time_unit     = optional(string, "")
    size          = optional(number, 0)
    size_unit     = optional(string, "")
    header_name   = optional(string, "")
    header_mode   = optional(string, "")
    header_values = optional(list(string), [])
    overlimit_time        = optional(number, 0)
    introspection         = optional(bool, false)
    debug_enabled         = optional(bool, false)
    max_depth             = optional(number, 0)
    max_value_size_kb     = optional(number, 0)
    max_doc_size_kb       = optional(number, 0)
    max_alias_size_kb     = optional(number, 0)
    max_doc_per_batch     = optional(number, 0)
    login_point           = optional(list(list(string)), [])
    login_regex           = optional(string, "")
    case_sensitive        = optional(bool, false)
    cred_stuff_type       = optional(string, "default")
    threshold             = optional(object({ count = number, period = number }))
    reaction              = optional(object({ block_by_session = optional(number), block_by_ip = optional(number), graylist_by_ip = optional(number) }))
    enumerated_parameters = optional(any)
    metadata              = optional(any)
    _config_dir           = string
    _action_conditions    = optional(list(object({ type = string, point = list(string), value = optional(string, "") })), [])
    _action_hash          = optional(string, "")
  }))
  default     = []
  description = "Rules from generators (hits/imports). Each object has the same fields as a YAML config plus _config_dir and action metadata."
}
