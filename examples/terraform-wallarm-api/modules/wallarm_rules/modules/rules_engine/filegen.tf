# ─── Config file generation ───────────────────────────────────────────────────
# YAML configs for generated rules (hits/imports). Written on first apply.
# File names use source prefix: hits_, imported_ (manual rules have no prefix).

locals {
  yaml_template_vars = {
    for name, r in local.rule_configs :
    name => {
      name          = name
      resource_type = r.resource_type
      comment       = r.comment
      path          = try(r.path, "")
      domain        = try(r.domain, "")
      instance      = try(r.instance, "")
      method        = try(r.method, "")
      scheme        = try(r.scheme, "")
      proto         = try(r.proto, "")
      query         = try(r.query, [])
      headers       = try(r.headers, [])
      point         = r.point
      attack_types  = r.attack_types
      stamps        = r.stamps
      file_types    = try(r.file_types, [])
      parsers       = try(r.parsers, [])
      attack_type   = r.attack_type
      mode          = r.mode
      regex         = r.regex
      regex_id      = r.regex_id
      regex_rule    = r.regex_rule
      experimental  = r.experimental
      parser        = r.parser
      file_type     = r.file_type
      delay         = r.delay
      burst         = r.burst
      rate          = r.rate
      rsp_status    = r.rsp_status
      time_unit     = r.time_unit
      size          = r.size
      size_unit     = r.size_unit
      header_name   = r.header_name
      header_mode   = r.header_mode
      header_values = r.header_values
      overlimit_time        = r.overlimit_time
      introspection         = r.introspection
      debug_enabled         = r.debug_enabled
      max_depth             = r.max_depth
      max_value_size_kb     = r.max_value_size_kb
      max_doc_size_kb       = r.max_doc_size_kb
      max_alias_size_kb     = r.max_alias_size_kb
      max_doc_per_batch     = r.max_doc_per_batch
      login_point           = r.login_point
      login_regex           = r.login_regex
      case_sensitive        = r.case_sensitive
      cred_stuff_type       = r.cred_stuff_type
      threshold             = r.threshold
      reaction              = r.reaction
      enumerated_parameters = r.enumerated_parameters
      metadata              = try(local.all_rules[name].metadata, null)
    }
  }
}

# ─── Write YAML for generated rules (hits/imports) ──────────────────────────
# wallarm_config_file: state-only delete → file persists on disk after for_each
# key disappears. fileset() discovers it → rules_engine continues to manage the rule.

resource "wallarm_config_file" "rule_configs" {
  for_each = { for r in var.generated_rules : r.name => trimprefix(r._config_dir, "./") }

  path    = "${each.value}/${each.key}.yaml"
  content = templatefile("${path.module}/templates/rule_config.yaml.tftpl", local.yaml_template_vars[each.key])

  lifecycle {
    ignore_changes = [content]
  }
}

# ─── Generate .action.yaml for each unique action directory ──────────────────

locals {
  generated_action_dirs = {
    for r in var.generated_rules :
    trimprefix(r._config_dir, "./") => {
      conditions      = try(r._action_conditions, [])
      conditions_hash = try(r._action_hash, "")
      path            = try(r.path, "")
      domain          = try(r.domain, "")
      instance        = try(r.instance, "")
    }...
  }
}

# ─── Default action .action.yaml ─────────────────────────────────────────────

resource "local_file" "default_action" {
  filename        = "${var.config_dirs[0]}/_default/.action.yaml"
  file_permission = "0644"

  content = yamlencode({
    conditions      = []
    conditions_hash = "5b8b61bd5ed79de9b3d130436a1e5a63ec663e224557ccb981bbb491a891b4dc"
    action_path     = "/**/*.*"
    action_domain   = ""
    action_instance = ""
  })
}
