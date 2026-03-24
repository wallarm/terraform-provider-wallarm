# ═══════════════════════════════════════════════════════════════════════════════
# RULE RESOURCES — one block per resource type
# Path-to-action expansion is handled by the provider via action_* fields.
# All resources depend on generated config files to ensure proper ordering.
# ═══════════════════════════════════════════════════════════════════════════════

# ─── binary_data ──────────────────────────────────────────────────────────────

resource "wallarm_rule_binary_data" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_binary_data" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  dynamic "action_query" {
    for_each = local.rule_configs[each.key].query
    content {
      key   = action_query.value.key
      value = action_query.value.value
      type  = try(action_query.value.type, "equal")
    }
  }

  dynamic "action_header" {
    for_each = local.rule_configs[each.key].headers
    content {
      name  = action_header.value.name
      value = action_header.value.value
      type  = try(action_header.value.type, "equal")
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── masking (sensitive_data) ─────────────────────────────────────────────────

resource "wallarm_rule_masking" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_masking" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  dynamic "action_query" {
    for_each = local.rule_configs[each.key].query
    content {
      key   = action_query.value.key
      value = action_query.value.value
      type  = try(action_query.value.type, "equal")
    }
  }

  dynamic "action_header" {
    for_each = local.rule_configs[each.key].headers
    content {
      name  = action_header.value.name
      value = action_header.value.value
      type  = try(action_header.value.type, "equal")
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── disable_attack_type (expanded by attack_types) ──────────────────────────

resource "wallarm_rule_disable_attack_type" "this" {
  for_each        = { for k, v in local.attack_type_rules : k => v if v.resource_type == "wallarm_rule_disable_attack_type" }
  client_id       = var.client_id
  attack_type     = each.value.attack_type
  comment              = local.rule_configs[each.value.config_name].comment
  variativity_disabled = local.rule_configs[each.value.config_name].variativity_disabled
  point           = local.rule_configs[each.value.config_name].point
  action_path     = local.rule_configs[each.value.config_name].path
  action_domain   = local.rule_configs[each.value.config_name].domain
  action_instance = local.rule_configs[each.value.config_name].instance
  action_method   = local.rule_configs[each.value.config_name].method
  action_scheme   = local.rule_configs[each.value.config_name].scheme
  action_proto    = local.rule_configs[each.value.config_name].proto

  dynamic "action_query" {
    for_each = local.rule_configs[each.value.config_name].query
    content {
      key   = action_query.value.key
      value = action_query.value.value
      type  = try(action_query.value.type, "equal")
    }
  }

  dynamic "action_header" {
    for_each = local.rule_configs[each.value.config_name].headers
    content {
      name  = action_header.value.name
      value = action_header.value.value
      type  = try(action_header.value.type, "equal")
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── disable_stamp (expanded by stamps) ──────────────────────────────────────

resource "wallarm_rule_disable_stamp" "this" {
  for_each        = local.stamp_rules
  client_id       = var.client_id
  stamp           = each.value.stamp
  comment              = local.rule_configs[each.value.config_name].comment
  variativity_disabled = local.rule_configs[each.value.config_name].variativity_disabled
  point           = local.rule_configs[each.value.config_name].point
  action_path     = local.rule_configs[each.value.config_name].path
  action_domain   = local.rule_configs[each.value.config_name].domain
  action_instance = local.rule_configs[each.value.config_name].instance
  action_method   = local.rule_configs[each.value.config_name].method
  action_scheme   = local.rule_configs[each.value.config_name].scheme
  action_proto    = local.rule_configs[each.value.config_name].proto

  dynamic "action_query" {
    for_each = local.rule_configs[each.value.config_name].query
    content {
      key   = action_query.value.key
      value = action_query.value.value
      type  = try(action_query.value.type, "equal")
    }
  }

  dynamic "action_header" {
    for_each = local.rule_configs[each.value.config_name].headers
    content {
      name  = action_header.value.name
      value = action_header.value.value
      type  = try(action_header.value.type, "equal")
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── vpatch (expanded by attack_types) ───────────────────────────────────────

resource "wallarm_rule_vpatch" "this" {
  for_each        = { for k, v in local.attack_type_rules : k => v if v.resource_type == "wallarm_rule_vpatch" }
  client_id       = var.client_id
  attack_type     = each.value.attack_type
  comment              = local.rule_configs[each.value.config_name].comment
  variativity_disabled = local.rule_configs[each.value.config_name].variativity_disabled
  point           = local.rule_configs[each.value.config_name].point
  action_path     = local.rule_configs[each.value.config_name].path
  action_domain   = local.rule_configs[each.value.config_name].domain
  action_instance = local.rule_configs[each.value.config_name].instance
  action_method   = local.rule_configs[each.value.config_name].method
  action_scheme   = local.rule_configs[each.value.config_name].scheme
  action_proto    = local.rule_configs[each.value.config_name].proto

  dynamic "action_query" {
    for_each = local.rule_configs[each.value.config_name].query
    content {
      key   = action_query.value.key
      value = action_query.value.value
      type  = try(action_query.value.type, "equal")
    }
  }

  dynamic "action_header" {
    for_each = local.rule_configs[each.value.config_name].headers
    content {
      name  = action_header.value.name
      value = action_header.value.value
      type  = try(action_header.value.type, "equal")
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── uploads (expanded by file_types OR single file_type) ────────────────────

resource "wallarm_rule_uploads" "expanded" {
  for_each        = local.file_type_rules
  client_id       = var.client_id
  comment              = local.rule_configs[each.value.config_name].comment
  variativity_disabled = local.rule_configs[each.value.config_name].variativity_disabled
  file_type       = each.value.file_type
  point           = local.rule_configs[each.value.config_name].point
  action_path     = local.rule_configs[each.value.config_name].path
  action_domain   = local.rule_configs[each.value.config_name].domain
  action_instance = local.rule_configs[each.value.config_name].instance
  action_method   = local.rule_configs[each.value.config_name].method
  action_scheme   = local.rule_configs[each.value.config_name].scheme
  action_proto    = local.rule_configs[each.value.config_name].proto
  depends_on      = [terraform_data.write_configs]
}

resource "wallarm_rule_uploads" "single" {
  for_each        = local.uploads_single
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  file_type       = local.rule_configs[each.key].file_type
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── parser_state (expanded by parsers OR single parser) ─────────────────────

resource "wallarm_rule_parser_state" "expanded" {
  for_each        = local.parser_rules
  client_id       = var.client_id
  comment              = local.rule_configs[each.value.config_name].comment
  variativity_disabled = local.rule_configs[each.value.config_name].variativity_disabled
  parser          = each.value.parser
  state           = "disabled"
  point           = local.rule_configs[each.value.config_name].point
  action_path     = local.rule_configs[each.value.config_name].path
  action_domain   = local.rule_configs[each.value.config_name].domain
  action_instance = local.rule_configs[each.value.config_name].instance
  action_method   = local.rule_configs[each.value.config_name].method
  action_scheme   = local.rule_configs[each.value.config_name].scheme
  action_proto    = local.rule_configs[each.value.config_name].proto
  depends_on      = [terraform_data.write_configs]
}

resource "wallarm_rule_parser_state" "single" {
  for_each        = local.parser_state_single
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  parser          = local.rule_configs[each.key].parser
  state           = try(local.rule_configs[each.key].state, "disabled")
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── ignore_regex (disable_regex) ─────────────────────────────────────────────

resource "wallarm_rule_ignore_regex" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_ignore_regex" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  regex_id = (
    try(local.rule_configs[each.key].regex_rule, "") != ""
    ? wallarm_rule_regex.this[local.rule_configs[each.key].regex_rule].regex_id
    : try(local.rule_configs[each.key].regex_id, 0)
  )

  depends_on = [wallarm_rule_regex.this, terraform_data.write_configs]
}

# ─── regex ────────────────────────────────────────────────────────────────────

resource "wallarm_rule_regex" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_regex" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  attack_type     = local.rule_configs[each.key].attack_type
  regex           = local.rule_configs[each.key].regex
  experimental    = local.rule_configs[each.key].experimental
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── file_upload_size_limit ──────────────────────────────────────────────────

resource "wallarm_rule_file_upload_size_limit" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_file_upload_size_limit" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].mode
  size            = local.rule_configs[each.key].size
  size_unit       = local.rule_configs[each.key].size_unit
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── rate_limit ──────────────────────────────────────────────────────────────

resource "wallarm_rule_rate_limit" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_rate_limit" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  delay           = local.rule_configs[each.key].delay
  burst           = local.rule_configs[each.key].burst
  rate            = local.rule_configs[each.key].rate
  rsp_status      = local.rule_configs[each.key].rsp_status
  time_unit       = local.rule_configs[each.key].time_unit
  point           = local.rule_configs[each.key].point
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── credential_stuffing_point ───────────────────────────────────────────────

resource "wallarm_rule_credential_stuffing_point" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_credential_stuffing_point" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  point           = local.rule_configs[each.key].point
  login_point     = local.rule_configs[each.key].login_point
  cred_stuff_type = local.rule_configs[each.key].cred_stuff_type
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── credential_stuffing_regex ───────────────────────────────────────────────

resource "wallarm_rule_credential_stuffing_regex" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_credential_stuffing_regex" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  regex           = local.rule_configs[each.key].regex
  login_regex     = local.rule_configs[each.key].login_regex
  case_sensitive  = local.rule_configs[each.key].case_sensitive
  cred_stuff_type = local.rule_configs[each.key].cred_stuff_type
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── mode ────────────────────────────────────────────────────────────────────

resource "wallarm_rule_mode" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_mode" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].mode
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  dynamic "action_query" {
    for_each = local.rule_configs[each.key].query
    content {
      key   = action_query.value.key
      value = action_query.value.value
      type  = try(action_query.value.type, "equal")
    }
  }

  dynamic "action_header" {
    for_each = local.rule_configs[each.key].headers
    content {
      name  = action_header.value.name
      value = action_header.value.value
      type  = try(action_header.value.type, "equal")
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── set_response_header ────────────────────────────────────────────────────

resource "wallarm_rule_set_response_header" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_set_response_header" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].header_mode
  name            = local.rule_configs[each.key].header_name
  values          = try(toset(local.rule_configs[each.key].header_values), [])
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── overlimit_res_settings ─────────────────────────────────────────────────

resource "wallarm_rule_overlimit_res_settings" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_overlimit_res_settings" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  overlimit_time  = local.rule_configs[each.key].overlimit_time
  mode            = local.rule_configs[each.key].mode
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── graphql_detection ───────────────────────────────────────────────────────

resource "wallarm_rule_graphql_detection" "this" {
  for_each          = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_graphql_detection" }
  client_id         = var.client_id
  comment           = local.rule_configs[each.key].comment
  mode              = local.rule_configs[each.key].mode
  max_depth         = local.rule_configs[each.key].max_depth
  max_value_size_kb = local.rule_configs[each.key].max_value_size_kb
  max_doc_size_kb   = local.rule_configs[each.key].max_doc_size_kb
  max_alias_size_kb = local.rule_configs[each.key].max_alias_size_kb
  max_doc_per_batch = local.rule_configs[each.key].max_doc_per_batch
  introspection     = local.rule_configs[each.key].introspection
  debug_enabled     = local.rule_configs[each.key].debug_enabled
  action_path       = local.rule_configs[each.key].path
  action_domain     = local.rule_configs[each.key].domain
  action_instance   = local.rule_configs[each.key].instance
  action_method     = local.rule_configs[each.key].method
  action_scheme     = local.rule_configs[each.key].scheme
  action_proto      = local.rule_configs[each.key].proto
  depends_on        = [terraform_data.write_configs]
}

# ─── bruteforce_counter ─────────────────────────────────────────────────────

resource "wallarm_rule_bruteforce_counter" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_bruteforce_counter" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── dirbust_counter ────────────────────────────────────────────────────────

resource "wallarm_rule_dirbust_counter" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_dirbust_counter" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── bola_counter ───────────────────────────────────────────────────────────

resource "wallarm_rule_bola_counter" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_bola_counter" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto
  depends_on      = [terraform_data.write_configs]
}

# ─── brute ───────────────────────────────────────────────────────────────────

resource "wallarm_rule_brute" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_brute" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].mode
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
  }

  dynamic "enumerated_parameters" {
    for_each = try(local.rule_configs[each.key].enumerated_parameters.mode, "") == "regexp" ? [1] : []
    content {
      mode                  = local.normalized_enum_params_regexp[each.key].mode
      name_regexps          = local.normalized_enum_params_regexp[each.key].name_regexps
      value_regexps         = local.normalized_enum_params_regexp[each.key].value_regexps
      additional_parameters = local.normalized_enum_params_regexp[each.key].additional_parameters
      plain_parameters      = local.normalized_enum_params_regexp[each.key].plain_parameters
    }
  }
  dynamic "enumerated_parameters" {
    for_each = try(local.rule_configs[each.key].enumerated_parameters.mode, "") == "exact" ? [1] : []
    content {
      mode = local.normalized_enum_params_exact[each.key].mode
      dynamic "points" {
        for_each = local.normalized_enum_params_exact[each.key].points
        content {
          point     = points.value.point
          sensitive = try(points.value.sensitive, false)
        }
      }
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── bola ────────────────────────────────────────────────────────────────────

resource "wallarm_rule_bola" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_bola" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].mode
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
  }

  dynamic "enumerated_parameters" {
    for_each = try(local.rule_configs[each.key].enumerated_parameters.mode, "") == "regexp" ? [1] : []
    content {
      mode                  = local.normalized_enum_params_regexp[each.key].mode
      name_regexps          = local.normalized_enum_params_regexp[each.key].name_regexps
      value_regexps         = local.normalized_enum_params_regexp[each.key].value_regexps
      additional_parameters = local.normalized_enum_params_regexp[each.key].additional_parameters
      plain_parameters      = local.normalized_enum_params_regexp[each.key].plain_parameters
    }
  }
  dynamic "enumerated_parameters" {
    for_each = try(local.rule_configs[each.key].enumerated_parameters.mode, "") == "exact" ? [1] : []
    content {
      mode = local.normalized_enum_params_exact[each.key].mode
      dynamic "points" {
        for_each = local.normalized_enum_params_exact[each.key].points
        content {
          point     = points.value.point
          sensitive = try(points.value.sensitive, false)
        }
      }
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── enum ────────────────────────────────────────────────────────────────────

resource "wallarm_rule_enum" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_enum" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].mode
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
  }

  dynamic "enumerated_parameters" {
    for_each = try(local.rule_configs[each.key].enumerated_parameters.mode, "") == "regexp" ? [1] : []
    content {
      mode                  = local.normalized_enum_params_regexp[each.key].mode
      name_regexps          = local.normalized_enum_params_regexp[each.key].name_regexps
      value_regexps         = local.normalized_enum_params_regexp[each.key].value_regexps
      additional_parameters = local.normalized_enum_params_regexp[each.key].additional_parameters
      plain_parameters      = local.normalized_enum_params_regexp[each.key].plain_parameters
    }
  }
  dynamic "enumerated_parameters" {
    for_each = try(local.rule_configs[each.key].enumerated_parameters.mode, "") == "exact" ? [1] : []
    content {
      mode = local.normalized_enum_params_exact[each.key].mode
      dynamic "points" {
        for_each = local.normalized_enum_params_exact[each.key].points
        content {
          point     = points.value.point
          sensitive = try(points.value.sensitive, false)
        }
      }
    }
  }

  depends_on = [terraform_data.write_configs]
}

# ─── rate_limit_enum ─────────────────────────────────────────────────────────

resource "wallarm_rule_rate_limit_enum" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_rate_limit_enum" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].mode
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
  }

  depends_on = [terraform_data.write_configs]
}

# ─── forced_browsing ─────────────────────────────────────────────────────────

resource "wallarm_rule_forced_browsing" "this" {
  for_each        = { for n, rt in local.managed_rules : n => n if rt == "wallarm_rule_forced_browsing" }
  client_id       = var.client_id
  comment              = local.rule_configs[each.key].comment
  variativity_disabled = local.rule_configs[each.key].variativity_disabled
  mode            = local.rule_configs[each.key].mode
  action_path     = local.rule_configs[each.key].path
  action_domain   = local.rule_configs[each.key].domain
  action_instance = local.rule_configs[each.key].instance
  action_method   = local.rule_configs[each.key].method
  action_scheme   = local.rule_configs[each.key].scheme
  action_proto    = local.rule_configs[each.key].proto

  threshold {
    period = try(local.rule_configs[each.key].threshold.period, 0)
    count  = try(local.rule_configs[each.key].threshold.count, 0)
  }

  reaction {
    block_by_session = try(local.rule_configs[each.key].reaction.block_by_session, 0)
    block_by_ip      = try(local.rule_configs[each.key].reaction.block_by_ip, 0)
  }

  depends_on = [terraform_data.write_configs]
}
