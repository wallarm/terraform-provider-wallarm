locals {
  # Cartesian product of point_hash x rule_type — one config file per combination.
  # Filename uses first 8 chars of point_hash for brevity.
  point_rule_keys = merge([
    for ph, cfg in var.points : {
      for rt in var.rule_types :
      "${substr(ph, 0, 8)}_${rt}" => {
        point_hash    = ph
        rule_type     = rt
        point_wrapped = cfg.point_wrapped
        stamps        = cfg.stamps
        attack_types  = cfg.attack_types
        hit_ids       = try(cfg.hit_ids, [])
      }
    }
  ]...)

  # Read only editable fields from each config file.
  # Metadata is excluded — written fresh on every apply, never read back.
  rule_configs = {
    for key, cfg in local.point_rule_keys :
    key => try(fileexists("${var.config_dir}/${var.request_id}/${key}.yaml"), false) ? {
      stamps = try(yamldecode(file("${var.config_dir}/${var.request_id}/${key}.yaml")).stamps, [])
      # For disable_stamp, attack_types is metadata derived from the hit (persisted in
      # terraform_data via effective_hits). Never read it back from the YAML file —
      # it isn't stored there. For other rule types, read from YAML as normal.
      attack_types = can(regex("_disable_stamp$", key)) ? cfg.attack_types : try(yamldecode(file("${var.config_dir}/${var.request_id}/${key}.yaml")).attack_types, [])
      point        = yamldecode(file("${var.config_dir}/${var.request_id}/${key}.yaml")).point
      action       = yamldecode(file("${var.config_dir}/${var.request_id}/${key}.yaml")).action
    } : jsondecode(jsonencode({
      stamps       = cfg.stamps
      attack_types = cfg.attack_types
      point        = cfg.point_wrapped
      action       = var.action
    }))
  }

  # disable_stamp: one rule per (point_hash, stamp)
  stamp_rules = contains(var.rule_types, "disable_stamp") ? merge([
    for key, cfg in local.rule_configs :
    can(regex("_disable_stamp$", key)) ? {
      for s in cfg.stamps :
      "${key}_${s}" => { config_key = key, stamp = s }
    } : {}
  ]...) : {}

  # disable_attack_type: one rule per (point_hash, attack_type)
  attack_type_rules = contains(var.rule_types, "disable_attack_type") ? merge([
    for key, cfg in local.rule_configs :
    can(regex("_disable_attack_type$", key)) ? {
      for at in cfg.attack_types :
      "${key}_${at}" => { config_key = key, attack_type = at }
    } : {}
  ]...) : {}
}

# ─── Generate config files ───────────────────────────────────────────────────

resource "local_file" "rule_config" {
  for_each = local.point_rule_keys

  filename        = "${var.config_dir}/${var.request_id}/${each.key}.yaml"
  file_permission = "0644"

  content = templatefile("${path.module}/templates/rule_config.yaml.tftpl", {
    # Editable — preserved from file on subsequent applies
    stamps       = local.rule_configs[each.key].stamps
    attack_types = local.rule_configs[each.key].attack_types
    point        = local.rule_configs[each.key].point
    action       = local.rule_configs[each.key].action
    # Metadata — always current, never read back by Terraform
    rule_type  = each.value.rule_type
    point_hash = each.value.point_hash
    request_id = var.request_id
    domain     = var.domain
    path       = var.path
    poolid     = var.poolid
    hit_ids    = each.value.hit_ids
  })
}

# ─── Create rules ─────────────────────────────────────────────────────────────

resource "wallarm_rule_disable_stamp" "this" {
  for_each = local.stamp_rules

  client_id = var.client_id
  stamp     = each.value.stamp
  point     = local.rule_configs[each.value.config_key].point

  dynamic "action" {
    for_each = local.rule_configs[each.value.config_key].action
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = action.value.point
    }
  }

  depends_on = [local_file.rule_config]
}

resource "wallarm_rule_disable_attack_type" "this" {
  for_each = local.attack_type_rules

  client_id   = var.client_id
  attack_type = each.value.attack_type
  point       = local.rule_configs[each.value.config_key].point

  dynamic "action" {
    for_each = local.rule_configs[each.value.config_key].action
    content {
      type  = action.value.type == "" ? null : action.value.type
      value = try(action.value.value, "")
      point = action.value.point
    }
  }

  depends_on = [local_file.rule_config]
}
