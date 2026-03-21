# ─── Config file generation ───────────────────────────────────────────────────
# YAML configs for generated rules (hits). Written on first apply.
# ignore_changes on content preserves user edits on subsequent applies.

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

# ─── Write YAML for generated rules (hits) ───────────────────────────────────

resource "local_file" "generated_config" {
  for_each = { for r in var.generated_rules : r.name => trimprefix(r._config_dir, "./") }

  filename        = "${each.value}/${each.key}.yaml"
  file_permission = "0644"

  content = templatefile("${path.module}/templates/rule_config.yaml.tftpl",
    local.yaml_template_vars[each.key]
  )

  lifecycle {
    ignore_changes = [content]
  }
}
