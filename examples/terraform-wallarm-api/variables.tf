variable "api_host" {
  type = string
}

variable "api_token" {
  type = string
}

variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "requests" {
  type        = map(list(string))
  description = "Static map of request_id => list of rule types. Example: { \"abc123\" = [\"disable_stamp\", \"disable_attack_type\"] }"
}

variable "custom_rules" {
  type = list(object({
    name          = string
    resource_type = string
    comment       = optional(string, "Managed by Terraform")
    point         = optional(list(list(string)))
    # Action scope
    path     = optional(string, "")
    domain   = optional(string, "")
    instance = optional(string, "")
    method   = optional(string, "")
    scheme   = optional(string, "")
    proto    = optional(string, "")
    # Query parameter conditions: each entry matches key=value in the query string
    query = optional(list(object({
      key   = string                     # query parameter name
      value = string                     # match value
      type  = optional(string, "equal") # match type: equal, iequal, regex, absent
    })), [])
    # Arbitrary header conditions (in addition to domain → HOST)
    headers = optional(list(object({
      name  = string                     # header name, e.g. "X-Forwarded-For", "Content-Type"
      value = string                     # match value
      type  = optional(string, "equal") # match type: equal, iequal, regex, absent
    })), [])
    # Expandable fields
    attack_types = optional(list(string), [])
    stamps       = optional(list(number), [])
    # Rule-specific fields
    attack_type    = optional(string, "")
    mode           = optional(string, "")
    regex          = optional(string, "")
    regex_id       = optional(number, 0)
    regex_rule     = optional(string, "")  # ignore_regex (reference a wallarm_rule_regex by name)
    experimental   = optional(bool, false)
    parser         = optional(string, "")
    state          = optional(string, "")
    file_type      = optional(string, "")
    delay          = optional(number, 0)
    burst          = optional(number, 0)
    rate           = optional(number, 0)
    rsp_status     = optional(number, 0)
    time_unit      = optional(string, "")
    size           = optional(number, 0)
    size_unit      = optional(string, "")
    header_name    = optional(string, "")
    header_mode    = optional(string, "")
    header_values  = optional(list(string), [])
    overlimit_time = optional(number, 0)
    introspection  = optional(bool, false)
    debug_enabled  = optional(bool, false)
    max_depth         = optional(number, 0)
    max_value_size_kb = optional(number, 0)
    max_doc_size_kb   = optional(number, 0)
    max_alias_size_kb = optional(number, 0)
    max_doc_per_batch = optional(number, 0)
    login_point     = optional(list(list(string)), [])
    login_regex     = optional(string, "")
    case_sensitive  = optional(bool, false)
    cred_stuff_type = optional(string, "default")
    threshold = optional(object({
      period = number
      count  = number
    }))
    reaction = optional(object({
      block_by_session = optional(number, 0)
      block_by_ip      = optional(number, 0)
      graylist_by_ip   = optional(number, 0)
    }))
    # Enumerated parameters (required for brute, bola, enum)
    enumerated_parameters = optional(object({
      mode = string # "regexp" or "exact"
      points = optional(list(object({
        point     = optional(list(string))
        sensitive = optional(bool, false)
      })), [])
      name_regexps          = optional(list(string), [""])
      value_regexps         = optional(list(string), [""])
      additional_parameters = optional(bool)
      plain_parameters      = optional(bool)
    }))
  }))
  default     = []
  description = "Custom rules defined directly in variables. Path is auto-expanded into action conditions."
}

variable "hits_mode" {
  type        = string
  default     = "request"
  description = "Fetch mode for hits data source: 'request' (direct hits only) or 'attack' (expand to all related hits by attack_id)"
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Must be true to activate rules import functionality."
}

variable "subnet_import_mode" {
  type        = string
  default     = "grouped"
  description = "How to import subnet entries: 'grouped' merges by expired_at into resources of max 1000 IPs; 'individual' creates one resource per IP."
}

variable "config_format" {
  type        = string
  default     = "yaml"
  description = "Config file format: 'yaml' for YAML files or 'hcl' for Terraform resource blocks"
}
