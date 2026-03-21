# ─── Config discovery ─────────────────────────────────────────────────────────
# Reads YAML configs from config_dirs, merges with generated_rules.
# Path-to-action expansion is handled by the provider via action_* fields.

locals {
  # ─── Discover YAML configs from all directories (recursive) ─────────────
  all_yaml_files = merge([
    for dir in var.config_dirs : {
      for f in try(fileset(dir, "**/*.yaml"), toset([])) :
      yamldecode(file("${dir}/${f}")).name => merge(
        yamldecode(file("${dir}/${f}")),
        {
          _source_dir  = dir
          _source_file = f
          _ref_dir     = dirname(f) != "." ? "${dir}/${dirname(f)}/_reference" : "${dir}/_reference"
        }
      )
      if can(yamldecode(file("${dir}/${f}")).name) && can(yamldecode(file("${dir}/${f}")).resource_type)
    }
  ]...)

  # Generated rules not yet persisted as YAML files (first apply only)
  gen_rules = { for r in var.generated_rules : r.name => r
    if !contains(keys(local.all_yaml_files), r.name)
  }

  # All rules: file-based + generated (file wins if both exist)
  all_rules = merge(
    { for name, r in local.gen_rules : name => merge(r, {
      _ref_dir = "${r._config_dir}/_reference"
    }) },
    local.all_yaml_files,
  )

  # Map: name → resource_type
  managed_rules = { for name, r in local.all_rules : name => try(r.resource_type, "") }

  # ─── Normalize rule fields with defaults ────────────────────────────────
  rule_configs = {
    for name, r in local.all_rules : name => {
      name          = name
      resource_type = try(r.resource_type, "")
      comment       = try(r.comment, "Managed by Terraform")

      # Scope fields — passed directly to provider action_* fields
      path     = try(r.path, "")
      domain   = try(r.domain, "")
      instance = try(tostring(r.instance), "")
      method   = try(r.method, "")
      scheme   = try(r.scheme, "")
      proto    = try(r.proto, "")
      query    = try(r.query, [])
      headers  = try(r.headers, [])

      # Detection point
      point = try(r.point, [])

      # Expandable fields
      attack_types = try(r.attack_types, [])
      stamps       = try(r.stamps, [])
      file_types   = try(r.file_types, [])
      parsers      = try(r.parsers, [])

      # Rule-specific fields
      attack_type    = try(r.attack_type, "")
      mode           = try(r.mode, "")
      regex          = try(r.regex, "")
      regex_id       = try(r.regex_id, 0)
      regex_rule     = try(r.regex_rule, "")
      experimental   = try(r.experimental, false)
      parser         = try(r.parser, "")
      file_type      = try(r.file_type, "")
      delay          = try(r.delay, 0)
      burst          = try(r.burst, 0)
      rate           = try(r.rate, 0)
      rsp_status     = try(r.rsp_status, 0)
      time_unit      = try(r.time_unit, "")
      size           = try(r.size, 0)
      size_unit      = try(r.size_unit, "")
      header_name    = try(r.header_name, "")
      header_mode    = try(r.header_mode, "")
      header_values  = try(r.header_values, [])
      overlimit_time = try(r.overlimit_time, 0)

      # GraphQL
      introspection     = try(r.introspection, false)
      debug_enabled     = try(r.debug_enabled, false)
      max_depth         = try(r.max_depth, 0)
      max_value_size_kb = try(r.max_value_size_kb, 0)
      max_doc_size_kb   = try(r.max_doc_size_kb, 0)
      max_alias_size_kb = try(r.max_alias_size_kb, 0)
      max_doc_per_batch = try(r.max_doc_per_batch, 0)

      # Credential stuffing
      login_point     = try(r.login_point, [])
      login_regex     = try(r.login_regex, "")
      case_sensitive  = try(r.case_sensitive, false)
      cred_stuff_type = try(r.cred_stuff_type, "default")

      # Threshold-based
      threshold             = try(r.threshold, null)
      reaction              = try(r.reaction, null)
      enumerated_parameters = try(r.enumerated_parameters, null)

      # Internal fields
      _ref_dir = try(r._ref_dir, "")
    }
  }

  # ─── Normalized enumerated_parameters for brute/bola/enum ──────────────
  normalized_enum_params_regexp = {
    for name, rt in local.managed_rules :
    name => {
      mode                  = try(local.rule_configs[name].enumerated_parameters.mode, "regexp")
      name_regexps          = tolist([for v in try(local.rule_configs[name].enumerated_parameters.name_regexps, [""]) : try(tostring(v), "")])
      value_regexps         = tolist([for v in try(local.rule_configs[name].enumerated_parameters.value_regexps, [""]) : try(tostring(v), "")])
      additional_parameters = try(local.rule_configs[name].enumerated_parameters.additional_parameters, false)
      plain_parameters      = try(local.rule_configs[name].enumerated_parameters.plain_parameters, false)
    }
    if(contains(["wallarm_rule_enum", "wallarm_rule_brute", "wallarm_rule_bola"], rt)
      && try(local.rule_configs[name].enumerated_parameters.mode, "") == "regexp")
  }

  normalized_enum_params_exact = {
    for name, rt in local.managed_rules :
    name => {
      mode   = try(local.rule_configs[name].enumerated_parameters.mode, "exact")
      points = try(local.rule_configs[name].enumerated_parameters.points, [])
    }
    if(contains(["wallarm_rule_enum", "wallarm_rule_brute", "wallarm_rule_bola"], rt)
      && try(local.rule_configs[name].enumerated_parameters.mode, "") == "exact")
  }
}
