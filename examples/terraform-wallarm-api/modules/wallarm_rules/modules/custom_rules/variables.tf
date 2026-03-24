variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "rules" {
  type = list(object({
    name          = string
    resource_type = string
    comment       = optional(string, "")

    # Detection point (required for most rule types)
    point = optional(list(list(string)))

    # Action scope — path auto-expanded into action conditions, other fields optional
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

    # ── Expandable fields (one rule created per list entry) ──────────────
    attack_types = optional(list(string), []) # disable_attack_type, vpatch
    stamps       = optional(list(number), []) # disable_stamp

    # ── Rule-specific fields ─────────────────────────────────────────────
    attack_type  = optional(string, "")  # regex, single-value vpatch/disable_attack_type
    mode         = optional(string, "")  # mode, brute, bola, enum, graphql, file_upload, overlimit, forced_browsing
    regex        = optional(string, "")  # regex
    regex_id     = optional(number, 0)   # ignore_regex (explicit ID, mutually exclusive with regex_rule)
    regex_rule   = optional(string, "")  # ignore_regex (reference a wallarm_rule_regex by name)
    experimental = optional(bool, false) # regex
    parser       = optional(string, "")  # parser_state
    state        = optional(string, "")  # parser_state
    file_type    = optional(string, "")  # uploads

    # Rate limiting
    delay      = optional(number, 0)
    burst      = optional(number, 0)
    rate       = optional(number, 0)
    rsp_status = optional(number, 0)
    time_unit  = optional(string, "") # rps, rpm

    # File upload size limit
    size      = optional(number, 0)
    size_unit = optional(string, "") # b, kb, mb, gb, tb

    # Response header
    header_name   = optional(string, "")
    header_mode   = optional(string, "") # append, replace
    header_values = optional(list(string), [])

    # Overlimit
    overlimit_time = optional(number, 0)

    # GraphQL
    introspection     = optional(bool, false)
    debug_enabled     = optional(bool, false)
    max_depth         = optional(number, 0)
    max_value_size_kb = optional(number, 0)
    max_doc_size_kb   = optional(number, 0)
    max_alias_size_kb = optional(number, 0)
    max_doc_per_batch = optional(number, 0)

    # Credential stuffing
    login_point     = optional(list(list(string)), [])
    login_regex     = optional(string, "")
    case_sensitive  = optional(bool, false)
    cred_stuff_type = optional(string, "default")

    # Threshold-based rules (brute, bola, enum, rate_limit_enum, forced_browsing)
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
  description = "Custom rule definitions. Path is auto-expanded into action conditions."
}

variable "config_dir" {
  type        = string
  description = "Directory where rule config files are written"
}

variable "config_format" {
  type        = string
  default     = "yaml"
  description = "Config file format: 'yaml' for YAML files or 'hcl' for Terraform resource blocks"

  validation {
    condition     = contains(["yaml", "hcl"], var.config_format)
    error_message = "config_format must be 'yaml' or 'hcl'."
  }
}
