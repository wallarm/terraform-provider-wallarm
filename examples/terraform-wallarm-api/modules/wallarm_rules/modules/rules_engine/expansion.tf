# ─── Multi-value expansion ────────────────────────────────────────────────────
# Rules with list fields (stamps, attack_types, file_types, parsers) are expanded
# into individual resources — one per list entry. All share the same action + point.

locals {
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

  # uploads: one rule per (name, file_type)
  file_type_rules = merge([
    for name, rt in local.managed_rules :
    rt == "wallarm_rule_uploads" ? {
      for ft in try(local.rule_configs[name].file_types, []) :
      "${name}_${ft}" => { config_name = name, file_type = ft }
    } : {}
  ]...)

  # parser_state: one rule per (name, parser) — state always "disabled"
  parser_rules = merge([
    for name, rt in local.managed_rules :
    rt == "wallarm_rule_parser_state" ? {
      for p in try(local.rule_configs[name].parsers, []) :
      "${name}_${p}" => { config_name = name, parser = p }
    } : {}
  ]...)

  # ─── Determine if a rule uses expansion or single-value ─────────────────
  # uploads: expanded if file_types list is non-empty, single if file_type is set
  uploads_single = {
    for name, rt in local.managed_rules : name => name
    if rt == "wallarm_rule_uploads" && length(try(local.rule_configs[name].file_types, [])) == 0
  }

  # parser_state: expanded if parsers list is non-empty, single if parser is set
  parser_state_single = {
    for name, rt in local.managed_rules : name => name
    if rt == "wallarm_rule_parser_state" && length(try(local.rule_configs[name].parsers, [])) == 0
  }
}
