# ─── Path-to-action expansion ──────────────────────────────────────────────────
# Mirrors buildActionFromHit() + locationToConditions() + actionNameExtConditions()
# from terraform-provider-wallarm/wallarm/provider/data_source_hits.go

locals {
  max_path_depth = 10

  rules_by_name = { for r in var.rules : r.name => r }

  # Discover existing YAML configs for reading config values.
  # Filenames follow the pattern: resource_type_name.yaml
  # Since both resource_type and name contain underscores, we read the name from inside the YAML.
  # NOTE: YAML files are used only for config overrides, NOT for lifecycle management.
  # Rule lifecycle is driven exclusively by var.rules.
  yaml_configs = { for f in fileset(var.config_dir, "*.yaml") :
    yamldecode(file("${var.config_dir}/${f}")).name => {
      filename      = f
      resource_type = yamldecode(file("${var.config_dir}/${f}")).resource_type
    }
    if can(yamldecode(file("${var.config_dir}/${f}")).resource_type) && can(yamldecode(file("${var.config_dir}/${f}")).name)
  }

  yaml_config_names = toset(keys(local.yaml_configs))

  # Rule lifecycle is driven by var.rules only.
  # Removing a rule from var.rules destroys the resource AND its config file.
  managed_rules = { for name, r in local.rules_by_name : name => r.resource_type }

  # ─── Split path into components per rule ──────────────────────────────────
  # Supports wildcards: * (any single value) and ** (any depth).
  path_parts = {
    for name, r in local.rules_by_name : name => {
      raw_parts = r.path != "" && r.path != "/" ? split("/", trimprefix(r.path, "/")) : []
      is_root   = r.path == "/" || r.path == ""
      too_deep  = (
        r.path != "" && r.path != "/"
        && !contains(r.path != "" && r.path != "/" ? split("/", trimprefix(r.path, "/")) : [], "**")
        && length(split("/", r.path)) > local.max_path_depth
      )
    }
  }

  path_details = {
    for name, r in local.rules_by_name : name => {
      # Directory segments (all but last element of raw_parts)
      dir_segments = (
        local.path_parts[name].is_root || local.path_parts[name].too_deep
        ? []
        : length(local.path_parts[name].raw_parts) > 1
        ? slice(local.path_parts[name].raw_parts, 0, length(local.path_parts[name].raw_parts) - 1)
        : []
      )

      # Final segment — the action component (action_name[.action_ext])
      last_segment = (
        local.path_parts[name].is_root || local.path_parts[name].too_deep
        ? ""
        : length(local.path_parts[name].raw_parts) > 0
        ? local.path_parts[name].raw_parts[length(local.path_parts[name].raw_parts) - 1]
        : ""
      )

      has_dot = (
        !local.path_parts[name].is_root && !local.path_parts[name].too_deep
        && length(local.path_parts[name].raw_parts) > 0
        ? length(split(".", local.path_parts[name].raw_parts[length(local.path_parts[name].raw_parts) - 1])) > 1
        : false
      )

      # ** as the last directory segment → indefinite depth (no limiter)
      has_globstar = (
        !local.path_parts[name].is_root && !local.path_parts[name].too_deep
        && length(local.path_parts[name].raw_parts) > 1
        && slice(local.path_parts[name].raw_parts, 0, length(local.path_parts[name].raw_parts) - 1)[
             length(local.path_parts[name].raw_parts) - 2
           ] == "**"
      )

      # Indexed segments: dirs with trailing ** stripped.
      # These produce path_N conditions; * segments are filtered in expanded_action.
      indexed_segments = (
        local.path_parts[name].is_root || local.path_parts[name].too_deep
        ? []
        : length(local.path_parts[name].raw_parts) > 1
        ? (
          slice(local.path_parts[name].raw_parts, 0, length(local.path_parts[name].raw_parts) - 1)[
            length(local.path_parts[name].raw_parts) - 2
          ] == "**"
          ? slice(local.path_parts[name].raw_parts, 0, length(local.path_parts[name].raw_parts) - 2)
          : slice(local.path_parts[name].raw_parts, 0, length(local.path_parts[name].raw_parts) - 1)
        )
        : []
      )

      # Limiter: path_N absent after last indexed segment.
      # Suppressed when ** is present (allows any depth).
      has_limiter = (
        !local.path_parts[name].is_root && !local.path_parts[name].too_deep
        && r.path != ""
        && !(
          length(local.path_parts[name].raw_parts) > 1
          && slice(local.path_parts[name].raw_parts, 0, length(local.path_parts[name].raw_parts) - 1)[
               length(local.path_parts[name].raw_parts) - 2
             ] == "**"
        )
      )
    }
  }

  action_name_ext = {
    for name, r in local.rules_by_name : name => {
      action_name_raw = (
        local.path_parts[name].is_root ? "" :
        local.path_parts[name].too_deep ? "" :
        local.path_details[name].last_segment == "" ? "" :
        local.path_details[name].has_dot
        ? join(".", slice(
          split(".", local.path_details[name].last_segment),
          0,
          length(split(".", local.path_details[name].last_segment)) - 1
        ))
        : local.path_details[name].last_segment
      )
      action_ext_raw = (
        local.path_parts[name].is_root ? "" :
        local.path_parts[name].too_deep ? "" :
        local.path_details[name].last_segment == "" ? "" :
        local.path_details[name].has_dot
        ? split(".", local.path_details[name].last_segment)[length(split(".", local.path_details[name].last_segment)) - 1]
        : ""
      )
      # * as action_name → skip condition (match any)
      action_name_is_wildcard = (
        !local.path_parts[name].is_root && !local.path_parts[name].too_deep
        && local.path_details[name].last_segment != ""
        && (
          local.path_details[name].has_dot
          ? join(".", slice(
              split(".", local.path_details[name].last_segment),
              0,
              length(split(".", local.path_details[name].last_segment)) - 1
            )) == "*"
          : local.path_details[name].last_segment == "*"
        )
      )
      # * as action_ext → skip condition (match any extension)
      action_ext_is_wildcard = (
        !local.path_parts[name].is_root && !local.path_parts[name].too_deep
        && local.path_details[name].last_segment != ""
        && local.path_details[name].has_dot
        && split(".", local.path_details[name].last_segment)[length(split(".", local.path_details[name].last_segment)) - 1] == "*"
      )
      # No dot in final segment → action_ext is absent
      ext_absent = (
        !local.path_parts[name].is_root
        && !local.path_parts[name].too_deep
        && local.path_details[name].last_segment != ""
        && !local.path_details[name].has_dot
      )
    }
  }

  # ─── Validation: reject invalid ** patterns ─────────────────────────────
  path_validation = {
    for name, r in local.rules_by_name : name => (
      # final segment must not be "**" (action_name must be defined)
      !(
        length(local.path_parts[name].raw_parts) > 0
        && local.path_parts[name].raw_parts[length(local.path_parts[name].raw_parts) - 1] == "**"
      )
      # "**" must not appear in indexed_segments (only allowed as last dir, which is stripped)
      && !contains(local.path_details[name].indexed_segments, "**")
    )
  }

  # ─── Build expanded action per rule ──────────────────────────────────────
  expanded_action = {
    for name, r in local.rules_by_name : name => concat(
      # Instance
      r.instance != "" ? [{ type = "", value = "", point = { instance = r.instance } }] : [],
      # Domain (HOST header) — skip when domain is "*" (match any)
      r.domain != "" && r.domain != "*" ? [{ type = "iequal", value = r.domain, point = { header = "HOST" } }] : [],
      # Custom headers
      [for h in r.headers : { type = h.type, value = h.value, point = { header = h.name } }],
      # too_deep fallback → single URI match
      local.path_parts[name].too_deep ? [
        { type = "equal", value = r.path, point = { uri = r.path } }
      ] : [],
      # Root path "/"
      local.path_parts[name].is_root && r.path == "/" ? [
        { type = "equal", value = "", point = { action_name = "" } },
        { type = "absent", value = "", point = { action_ext = "" } },
        { type = "absent", value = "", point = { path = "0" } },
      ] : [],
      # action_name — skip when wildcard *
      (
        !local.path_parts[name].is_root && !local.path_parts[name].too_deep && r.path != ""
        && !local.action_name_ext[name].action_name_is_wildcard
      ) ? [
        { type = "equal", value = "", point = { action_name = local.action_name_ext[name].action_name_raw } }
      ] : [],
      # action_ext: absent (no dot) | equal (specific ext) | skip (wildcard *)
      (
        !local.path_parts[name].is_root && !local.path_parts[name].too_deep && r.path != ""
      ) ? (
        local.action_name_ext[name].ext_absent
        ? [{ type = "absent", value = "", point = { action_ext = "" } }]
        : (
          !local.action_name_ext[name].action_ext_is_wildcard
          ? [{ type = "equal", value = "", point = { action_ext = local.action_name_ext[name].action_ext_raw } }]
          : []
        )
      ) : [],
      # Path segments — skip * wildcards (match any value at that position)
      !local.path_parts[name].is_root && !local.path_parts[name].too_deep ? [
        for i, seg in local.path_details[name].indexed_segments :
        { type = "equal", value = seg, point = { path = tostring(i) } }
        if seg != "*"
      ] : [],
      # Path limiter — suppressed when ** is present
      local.path_details[name].has_limiter ? [
        { type = "absent", value = "", point = { path = tostring(length(local.path_details[name].indexed_segments)) } }
      ] : [],
      # Method, scheme, proto, query
      r.method != "" ? [{ type = "equal", value = "", point = { method = r.method } }] : [],
      r.scheme != "" ? [{ type = "equal", value = "", point = { scheme = r.scheme } }] : [],
      r.proto != "" ? [{ type = "equal", value = "", point = { proto = r.proto } }] : [],
      [for q in r.query : { type = try(q.type, "equal"), value = q.value, point = { query = q.key } }],
    )
  }

  # ─── Rule configs: variables-first pattern ───────────────────────────────
  # Variables are the authoritative source. YAML provides base values only —
  # variable values always override YAML. Action is always computed.
  rule_configs = {
    for name, rt in local.managed_rules :
    name => merge(
      try(jsondecode(jsonencode(yamldecode(file("${var.config_dir}/${local.yaml_configs[name].filename}")))), {}),
      try(jsondecode(jsonencode({
        resource_type         = local.rules_by_name[name].resource_type
        comment               = local.rules_by_name[name].comment
        point                 = local.rules_by_name[name].point
        action                = local.expanded_action[name]
        attack_types          = local.rules_by_name[name].attack_types
        stamps                = local.rules_by_name[name].stamps
        attack_type           = local.rules_by_name[name].attack_type
        mode                  = local.rules_by_name[name].mode
        regex                 = local.rules_by_name[name].regex
        regex_id              = local.rules_by_name[name].regex_id
        regex_rule            = local.rules_by_name[name].regex_rule
        experimental          = local.rules_by_name[name].experimental
        parser                = local.rules_by_name[name].parser
        state                 = local.rules_by_name[name].state
        file_type             = local.rules_by_name[name].file_type
        delay                 = local.rules_by_name[name].delay
        burst                 = local.rules_by_name[name].burst
        rate                  = local.rules_by_name[name].rate
        rsp_status            = local.rules_by_name[name].rsp_status
        time_unit             = local.rules_by_name[name].time_unit
        size                  = local.rules_by_name[name].size
        size_unit             = local.rules_by_name[name].size_unit
        header_name           = local.rules_by_name[name].header_name
        header_mode           = local.rules_by_name[name].header_mode
        header_values         = local.rules_by_name[name].header_values
        overlimit_time        = local.rules_by_name[name].overlimit_time
        introspection         = local.rules_by_name[name].introspection
        debug_enabled         = local.rules_by_name[name].debug_enabled
        max_depth             = local.rules_by_name[name].max_depth
        max_value_size_kb     = local.rules_by_name[name].max_value_size_kb
        max_doc_size_kb       = local.rules_by_name[name].max_doc_size_kb
        max_alias_size_kb     = local.rules_by_name[name].max_alias_size_kb
        max_doc_per_batch     = local.rules_by_name[name].max_doc_per_batch
        login_point           = local.rules_by_name[name].login_point
        login_regex           = local.rules_by_name[name].login_regex
        case_sensitive        = local.rules_by_name[name].case_sensitive
        cred_stuff_type       = local.rules_by_name[name].cred_stuff_type
        threshold             = local.rules_by_name[name].threshold
        reaction              = local.rules_by_name[name].reaction
        enumerated_parameters = local.rules_by_name[name].enumerated_parameters
      })), {}),
      # action is always computed from expanded_action (path/domain/headers/etc.)
      { action = local.expanded_action[name] }
    )
  }

  # ─── Normalized enumerated_parameters for brute/bola/enum ─────────────────
  # Pre-compute with explicit types to avoid null/type coercion issues
  # in the resource blocks' ternary expressions.

  normalized_enum_params_regexp = {
    for name, rt in local.managed_rules :
    name => {
      mode                  = try(local.rule_configs[name].enumerated_parameters.mode, "regexp")
      name_regexps          = tolist([for v in try(local.rule_configs[name].enumerated_parameters.name_regexps, [""]) : try(tostring(v), "")])
      value_regexps         = tolist([for v in try(local.rule_configs[name].enumerated_parameters.value_regexps, [""]) : try(tostring(v), "")])
      additional_parameters = try(local.rule_configs[name].enumerated_parameters.additional_parameters, false)
      plain_parameters      = try(local.rule_configs[name].enumerated_parameters.plain_parameters, false)
    }
    if(contains(["wallarm_rule_enum", "wallarm_rule_brute", "wallarm_rule_bola"], rt) && local.rule_configs[name].enumerated_parameters.mode == "regexp")
  }

  normalized_enum_params_exact = {
    for name, rt in local.managed_rules :
    name => {
      mode   = try(local.rule_configs[name].enumerated_parameters.mode, "exact")
      points = try(local.rule_configs[name].enumerated_parameters.points, [])
    }
    if(contains(["wallarm_rule_enum", "wallarm_rule_brute", "wallarm_rule_bola"], rt) && local.rule_configs[name].enumerated_parameters.mode == "exact")
  }

  # ─── Expansion maps for multi-value rules ────────────────────────────────

  # disable_attack_type + vpatch: one rule per (name, attack_type)
  attack_type_rules = merge([
    for name, rt in local.managed_rules :
    contains(["wallarm_rule_disable_attack_type", "wallarm_rule_vpatch"], rt) ? {
      for at in try(local.rule_configs[name].attack_types, []) :
      "${name}_${at}" => { config_name = name, attack_type = at, resource_type = rt }
    } : {}
  ]...)

  # disable_stamp: one rule per (name, stamp)
  stamp_rules = merge([
    for name, rt in local.managed_rules :
    rt == "wallarm_rule_disable_stamp" ? {
      for s in try(local.rule_configs[name].stamps, []) :
      "${name}_${s}" => { config_name = name, stamp = s }
    } : {}
  ]...)
}

# ─── Generate YAML config files ──────────────────────────────────────────────
# ─── Path validation: reject invalid ** patterns ──────────────────────────────

resource "terraform_data" "path_validation" {
  for_each = { for name, valid in local.path_validation : name => local.rules_by_name[name].path if !valid }

  lifecycle {
    precondition {
      condition     = each.value == null
      error_message = "Invalid wildcard path '${each.value}' for rule '${each.key}': '**' must be the last directory segment (before the action component) and cannot be the final path component."
    }
  }
}

# Config files are created alongside resources. Removing a rule from var.rules
# destroys both the resource and its config file — no orphaned YAMLs.

locals {
  config_template_vars = {
    for name, rt in local.managed_rules :
    name => {
      name                  = name
      resource_type         = rt
      client_id             = var.client_id
      comment               = try(local.rule_configs[name].comment, "Managed by Terraform")
      attack_types          = try(local.rule_configs[name].attack_types, [])
      stamps                = try(local.rule_configs[name].stamps, [])
      attack_type           = try(local.rule_configs[name].attack_type, "")
      mode                  = try(local.rule_configs[name].mode, "")
      regex                 = try(local.rule_configs[name].regex, "")
      regex_id              = try(local.rule_configs[name].regex_id, 0)
      regex_rule            = try(local.rule_configs[name].regex_rule, "")
      experimental          = try(local.rule_configs[name].experimental, false)
      parser                = try(local.rule_configs[name].parser, "")
      state                 = try(local.rule_configs[name].state, "")
      file_type             = try(local.rule_configs[name].file_type, "")
      delay                 = try(local.rule_configs[name].delay, 0)
      burst                 = try(local.rule_configs[name].burst, 0)
      rate                  = try(local.rule_configs[name].rate, 0)
      rsp_status            = try(local.rule_configs[name].rsp_status, 0)
      time_unit             = try(local.rule_configs[name].time_unit, "")
      size                  = try(local.rule_configs[name].size, 0)
      size_unit             = try(local.rule_configs[name].size_unit, "")
      header_name           = try(local.rule_configs[name].header_name, "")
      header_mode           = try(local.rule_configs[name].header_mode, "")
      header_values         = try(local.rule_configs[name].header_values, [])
      overlimit_time        = try(local.rule_configs[name].overlimit_time, 0)
      introspection         = try(local.rule_configs[name].introspection, false)
      debug_enabled         = try(local.rule_configs[name].debug_enabled, false)
      max_depth             = try(local.rule_configs[name].max_depth, 0)
      max_value_size_kb     = try(local.rule_configs[name].max_value_size_kb, 0)
      max_doc_size_kb       = try(local.rule_configs[name].max_doc_size_kb, 0)
      max_alias_size_kb     = try(local.rule_configs[name].max_alias_size_kb, 0)
      max_doc_per_batch     = try(local.rule_configs[name].max_doc_per_batch, 0)
      login_point           = try(local.rule_configs[name].login_point, [])
      login_regex           = try(local.rule_configs[name].login_regex, "")
      case_sensitive        = try(local.rule_configs[name].case_sensitive, false)
      cred_stuff_type       = try(local.rule_configs[name].cred_stuff_type, "default")
      threshold             = try(local.rule_configs[name].threshold, null)
      reaction              = try(local.rule_configs[name].reaction, null)
      enumerated_parameters = try(local.rule_configs[name].enumerated_parameters, null)
      point                 = try(local.rule_configs[name].point, [])
      action                = try(local.rule_configs[name].action, [])
    }
  }

  config_ext      = var.config_format == "hcl" ? "tf" : "yaml"
  config_template = var.config_format == "hcl" ? "rule_config.tf.tftpl" : "rule_config.yaml.tftpl"
}

resource "local_file" "rule_config" {
  for_each = local.managed_rules

  filename        = "${var.config_dir}/${each.value}_${each.key}.${local.config_ext}"
  file_permission = "0644"

  content = templatefile("${path.module}/templates/${local.config_template}", local.config_template_vars[each.key])
}

# ═══════════════════════════════════════════════════════════════════════════════
# RULE RESOURCES — one block per resource type
# ═══════════════════════════════════════════════════════════════════════════════

# ─── binary_data ──────────────────────────────────────────────────────────────

resource "wallarm_rule_binary_data" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_binary_data" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  point     = try(local.rule_configs[each.key].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── masking (sensitive_data) ─────────────────────────────────────────────────

resource "wallarm_rule_masking" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_masking" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  point     = try(local.rule_configs[each.key].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── disable_attack_type (expanded by attack_types) ──────────────────────────

resource "wallarm_rule_disable_attack_type" "this" {
  for_each    = { for k, v in local.attack_type_rules : k => v if v.resource_type == "wallarm_rule_disable_attack_type" }
  client_id   = var.client_id
  attack_type = each.value.attack_type
  comment     = try(local.rule_configs[each.value.config_name].comment, "Managed by Terraform")
  point       = try(local.rule_configs[each.value.config_name].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.value.config_name].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── disable_stamp (expanded by stamps) ──────────────────────────────────────

resource "wallarm_rule_disable_stamp" "this" {
  for_each  = local.stamp_rules
  client_id = var.client_id
  stamp     = each.value.stamp
  comment   = try(local.rule_configs[each.value.config_name].comment, "Managed by Terraform")
  point     = try(local.rule_configs[each.value.config_name].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.value.config_name].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── vpatch (expanded by attack_types) ───────────────────────────────────────

resource "wallarm_rule_vpatch" "this" {
  for_each    = { for k, v in local.attack_type_rules : k => v if v.resource_type == "wallarm_rule_vpatch" }
  client_id   = var.client_id
  attack_type = each.value.attack_type
  comment     = try(local.rule_configs[each.value.config_name].comment, "Managed by Terraform")
  point       = try(local.rule_configs[each.value.config_name].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.value.config_name].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── uploads ──────────────────────────────────────────────────────────────────

resource "wallarm_rule_uploads" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_uploads" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  file_type = try(local.rule_configs[each.key].file_type, "")
  point     = try(local.rule_configs[each.key].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── ignore_regex (disable_regex) ─────────────────────────────────────────────

resource "wallarm_rule_ignore_regex" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_ignore_regex" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  point     = try(local.rule_configs[each.key].point, [])

  # Resolve regex_id: by name reference (regex_rule) or explicit numeric ID (regex_id)
  regex_id = (
    try(local.rule_configs[each.key].regex_rule, "") != ""
    ? wallarm_rule_regex.this[local.rule_configs[each.key].regex_rule].regex_id
    : try(local.rule_configs[each.key].regex_id, 0)
  )

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
  depends_on = [wallarm_rule_regex.this]
}

# ─── parser_state ─────────────────────────────────────────────────────────────

resource "wallarm_rule_parser_state" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_parser_state" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  parser    = try(local.rule_configs[each.key].parser, "")
  state     = try(local.rule_configs[each.key].state, "")
  point     = try(local.rule_configs[each.key].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── regex ────────────────────────────────────────────────────────────────────

resource "wallarm_rule_regex" "this" {
  for_each     = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_regex" }
  client_id    = var.client_id
  comment      = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  attack_type  = try(local.rule_configs[each.key].attack_type, "")
  regex        = try(local.rule_configs[each.key].regex, "")
  experimental = try(local.rule_configs[each.key].experimental, false)
  point        = try(local.rule_configs[each.key].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── file_upload_size_limit ───────────────────────────────────────────────────

resource "wallarm_rule_file_upload_size_limit" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_file_upload_size_limit" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].mode, "")
  size      = try(local.rule_configs[each.key].size, 0)
  size_unit = try(local.rule_configs[each.key].size_unit, "")
  point     = try(local.rule_configs[each.key].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── rate_limit ───────────────────────────────────────────────────────────────

resource "wallarm_rule_rate_limit" "this" {
  for_each   = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_rate_limit" }
  client_id  = var.client_id
  comment    = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  delay      = try(local.rule_configs[each.key].delay, 0)
  burst      = try(local.rule_configs[each.key].burst, 0)
  rate       = try(local.rule_configs[each.key].rate, 0)
  rsp_status = try(local.rule_configs[each.key].rsp_status, 0)
  time_unit  = try(local.rule_configs[each.key].time_unit, "")
  point      = try(local.rule_configs[each.key].point, [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── credential_stuffing_point ────────────────────────────────────────────────

resource "wallarm_rule_credential_stuffing_point" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_credential_stuffing_point" }
  client_id       = var.client_id
  comment         = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  point           = try(local.rule_configs[each.key].point, [])
  login_point     = try(local.rule_configs[each.key].login_point, [])
  cred_stuff_type = try(local.rule_configs[each.key].cred_stuff_type, "default")

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── credential_stuffing_regex ────────────────────────────────────────────────

resource "wallarm_rule_credential_stuffing_regex" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_credential_stuffing_regex" }
  client_id       = var.client_id
  comment         = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  regex           = try(local.rule_configs[each.key].regex, "")
  login_regex     = try(local.rule_configs[each.key].login_regex, "")
  case_sensitive  = try(local.rule_configs[each.key].case_sensitive, false)
  cred_stuff_type = try(local.rule_configs[each.key].cred_stuff_type, "default")

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── mode ─────────────────────────────────────────────────────────────────────

resource "wallarm_rule_mode" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_mode" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].mode, "")

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── set_response_header ─────────────────────────────────────────────────────

resource "wallarm_rule_set_response_header" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_set_response_header" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].header_mode, "")
  name      = try(local.rule_configs[each.key].header_name, "")
  values    = try(toset(local.rule_configs[each.key].header_values), [])

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── overlimit_res_settings ──────────────────────────────────────────────────

resource "wallarm_rule_overlimit_res_settings" "this" {
  for_each       = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_overlimit_res_settings" }
  client_id      = var.client_id
  comment        = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  overlimit_time = try(local.rule_configs[each.key].overlimit_time, 0)
  mode           = try(local.rule_configs[each.key].mode, "")

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── graphql_detection ────────────────────────────────────────────────────────

resource "wallarm_rule_graphql_detection" "this" {
  for_each          = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_graphql_detection" }
  client_id         = var.client_id
  comment           = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode              = try(local.rule_configs[each.key].mode, "")
  max_depth         = try(local.rule_configs[each.key].max_depth, 0)
  max_value_size_kb = try(local.rule_configs[each.key].max_value_size_kb, 0)
  max_doc_size_kb   = try(local.rule_configs[each.key].max_doc_size_kb, 0)
  max_alias_size_kb = try(local.rule_configs[each.key].max_alias_size_kb, 0)
  max_doc_per_batch = try(local.rule_configs[each.key].max_doc_per_batch, 0)
  introspection     = try(local.rule_configs[each.key].introspection, false)
  debug_enabled     = try(local.rule_configs[each.key].debug_enabled, false)

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── bruteforce_counter ──────────────────────────────────────────────────────

resource "wallarm_rule_bruteforce_counter" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_bruteforce_counter" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── dirbust_counter ─────────────────────────────────────────────────────────

resource "wallarm_rule_dirbust_counter" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_dirbust_counter" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── bola_counter ─────────────────────────────────────────────────────────────

resource "wallarm_rule_bola_counter" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_bola_counter" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── enum ─────────────────────────────────────────────────────────────────────

resource "wallarm_rule_enum" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_enum" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].mode, "")

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
    graylist_by_ip   = try(local.rule_configs[each.key].reaction.graylist_by_ip, 0)
  }

  dynamic "enumerated_parameters" {
    for_each = local.rule_configs[each.key].enumerated_parameters.mode == "regexp" ? [1] : []
    content {
      mode                  = local.normalized_enum_params_regexp[each.key].mode
      name_regexps          = local.normalized_enum_params_regexp[each.key].name_regexps
      value_regexps         = local.normalized_enum_params_regexp[each.key].value_regexps
      additional_parameters = local.normalized_enum_params_regexp[each.key].additional_parameters
      plain_parameters      = local.normalized_enum_params_regexp[each.key].plain_parameters
    }
  }
  dynamic "enumerated_parameters" {
    for_each = local.rule_configs[each.key].enumerated_parameters.mode == "exact" ? [1] : []
    content {
      mode = local.normalized_enum_params_exact[each.key].mode

      dynamic "points" {
        for_each = local.rule_configs[each.key].enumerated_parameters.mode == "exact" ? local.normalized_enum_params_exact[each.key].points : []
        content {
          point     = points.value.point
          sensitive = try(points.value.sensitive, false)
        }
      }
    }
  }

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── brute_enum ────────────────────────────────────────────────────────────────────

resource "wallarm_rule_brute" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_brute" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].mode, "")

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
    graylist_by_ip   = try(local.rule_configs[each.key].reaction.graylist_by_ip, 0)
  }

  dynamic "enumerated_parameters" {
    for_each = local.rule_configs[each.key].enumerated_parameters.mode == "regexp" ? [1] : []
    content {
      mode                  = local.normalized_enum_params_regexp[each.key].mode
      name_regexps          = local.normalized_enum_params_regexp[each.key].name_regexps
      value_regexps         = local.normalized_enum_params_regexp[each.key].value_regexps
      additional_parameters = local.normalized_enum_params_regexp[each.key].additional_parameters
      plain_parameters      = local.normalized_enum_params_regexp[each.key].plain_parameters
    }
  }
  dynamic "enumerated_parameters" {
    for_each = local.rule_configs[each.key].enumerated_parameters.mode == "exact" ? [1] : []
    content {
      mode = local.normalized_enum_params_exact[each.key].mode

      dynamic "points" {
        for_each = local.rule_configs[each.key].enumerated_parameters.mode == "exact" ? local.normalized_enum_params_exact[each.key].points : []
        content {
          point     = points.value.point
          sensitive = try(points.value.sensitive, false)
        }
      }
    }
  }

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── bola_enum ─────────────────────────────────────────────────────────────────────

resource "wallarm_rule_bola" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_bola" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].mode, "")

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
    graylist_by_ip   = try(local.rule_configs[each.key].reaction.graylist_by_ip, 0)
  }

  dynamic "enumerated_parameters" {
    for_each = local.rule_configs[each.key].enumerated_parameters.mode == "regexp" ? [1] : []
    content {
      mode                  = local.normalized_enum_params_regexp[each.key].mode
      name_regexps          = local.normalized_enum_params_regexp[each.key].name_regexps
      value_regexps         = local.normalized_enum_params_regexp[each.key].value_regexps
      additional_parameters = local.normalized_enum_params_regexp[each.key].additional_parameters
      plain_parameters      = local.normalized_enum_params_regexp[each.key].plain_parameters
    }
  }
  dynamic "enumerated_parameters" {
    for_each = local.rule_configs[each.key].enumerated_parameters.mode == "exact" ? [1] : []
    content {
      mode = local.normalized_enum_params_exact[each.key].mode

      dynamic "points" {
        for_each = local.rule_configs[each.key].enumerated_parameters.mode == "exact" ? local.normalized_enum_params_exact[each.key].points : []
        content {
          point     = points.value.point
          sensitive = try(points.value.sensitive, false)
        }
      }
    }
  }

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── rate_limit_enum ──────────────────────────────────────────────────────────

resource "wallarm_rule_rate_limit_enum" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_rate_limit_enum" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].mode, "")

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
    graylist_by_ip   = try(local.rule_configs[each.key].reaction.graylist_by_ip, 0)
  }

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}

# ─── forced_browsing_enum ─────────────────────────────────────────────────────────

resource "wallarm_rule_forced_browsing" "this" {
  for_each  = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_forced_browsing" }
  client_id = var.client_id
  comment   = try(local.rule_configs[each.key].comment, "Managed by Terraform")
  mode      = try(local.rule_configs[each.key].mode, "")

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
    graylist_by_ip   = try(local.rule_configs[each.key].reaction.graylist_by_ip, 0)
  }

  dynamic "action" {
    for_each = try(local.rule_configs[each.key].action, [])
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = { for k, v in action.value.point : k => k == "header" ? upper(v) : v }
    }
  }
}
