# ─── Import Generator ─────────────────────────────────────────────────────────
# Two modes:
#
# 1. Standard import (is_importing=true):
#    terraform apply -var='is_importing=true'   → wallarm_rule_imports.tf
#    terraform plan -generate-config-out=imported.tf
#    terraform apply
#
# 2. Convert to YAML (convert_imports=true, after standard import):
#    terraform apply -var='convert_imports=true' → YAML configs + moved blocks
#    - Remove imported.tf and wallarm_rule_imports.tf
#    terraform apply                             → state migrates to rules_engine

data "wallarm_rules" "all" {
  count     = var.is_importing || var.convert_imports ? 1 : 0
  client_id = var.client_id
  type      = var.rule_types
}

locals {
  exported = var.is_importing || var.convert_imports ? data.wallarm_rules.all[0].rules_export : []

  # ── Resource name mapping ───────────────────────────────────────────────
  _resource_label = {
    "wallarm_rule_binary_data"              = "wallarm_rule_binary_data"
    "wallarm_rule_masking"                  = "wallarm_rule_masking"
    "wallarm_rule_disable_attack_type"      = "wallarm_rule_disable_attack_type"
    "wallarm_rule_disable_stamp"            = "wallarm_rule_disable_stamp"
    "wallarm_rule_vpatch"                   = "wallarm_rule_vpatch"
    "wallarm_rule_uploads"                  = "wallarm_rule_uploads"
    "wallarm_rule_ignore_regex"             = "wallarm_rule_ignore_regex"
    "wallarm_rule_parser_state"             = "wallarm_rule_parser_state"
    "wallarm_rule_regex"                    = "wallarm_rule_regex"
    "wallarm_rule_file_upload_size_limit"   = "wallarm_rule_file_upload_size_limit"
    "wallarm_rule_rate_limit"               = "wallarm_rule_rate_limit"
    "wallarm_rule_credential_stuffing_point"  = "wallarm_rule_credential_stuffing_point"
    "wallarm_rule_credential_stuffing_regex"  = "wallarm_rule_credential_stuffing_regex"
    "wallarm_rule_mode"                     = "wallarm_rule_mode"
    "wallarm_rule_set_response_header"      = "wallarm_rule_set_response_header"
    "wallarm_rule_overlimit_res_settings"   = "wallarm_rule_overlimit_res_settings"
    "wallarm_rule_graphql_detection"        = "wallarm_rule_graphql_detection"
    "wallarm_rule_brute"                    = "wallarm_rule_brute"
    "wallarm_rule_bruteforce_counter"       = "wallarm_rule_bruteforce_counter"
    "wallarm_rule_dirbust_counter"          = "wallarm_rule_dirbust_counter"
    "wallarm_rule_bola"                     = "wallarm_rule_bola"
    "wallarm_rule_bola_counter"             = "wallarm_rule_bola_counter"
    "wallarm_rule_enum"                     = "wallarm_rule_enum"
    "wallarm_rule_rate_limit_enum"          = "wallarm_rule_rate_limit_enum"
    "wallarm_rule_forced_browsing"          = "wallarm_rule_forced_browsing"
  }

  # Rules engine resource labels for expanded types (different from standalone)
  _engine_label = {
    "wallarm_rule_uploads"      = "wallarm_rule_uploads.expanded"
    "wallarm_rule_parser_state" = "wallarm_rule_parser_state.expanded"
  }

  # ── Groupable types ─────────────────────────────────────────────────────
  stamp_by_action = { for r in local.exported :
    r.action_id => r...
    if r.terraform_resource == "wallarm_rule_disable_stamp"
  }

  disable_at_by_action = { for r in local.exported :
    r.action_id => r...
    if r.terraform_resource == "wallarm_rule_disable_attack_type"
  }

  vpatch_by_action = { for r in local.exported :
    r.action_id => r...
    if r.terraform_resource == "wallarm_rule_vpatch"
  }

  uploads_by_action = { for r in local.exported :
    r.action_id => r...
    if r.terraform_resource == "wallarm_rule_uploads"
  }

  parser_by_action = { for r in local.exported :
    r.action_id => r...
    if r.terraform_resource == "wallarm_rule_parser_state"
  }

  single_types = toset([
    "wallarm_rule_binary_data", "wallarm_rule_masking",
    "wallarm_rule_ignore_regex", "wallarm_rule_regex",
    "wallarm_rule_file_upload_size_limit", "wallarm_rule_rate_limit",
    "wallarm_rule_credential_stuffing_point", "wallarm_rule_credential_stuffing_regex",
    "wallarm_rule_mode", "wallarm_rule_set_response_header",
    "wallarm_rule_overlimit_res_settings", "wallarm_rule_graphql_detection",
    "wallarm_rule_brute", "wallarm_rule_bruteforce_counter",
    "wallarm_rule_dirbust_counter", "wallarm_rule_bola", "wallarm_rule_bola_counter",
    "wallarm_rule_enum", "wallarm_rule_rate_limit_enum", "wallarm_rule_forced_browsing",
  ])

  # ── Import blocks (is_importing mode) ──────────────────────────────────
  import_blocks = [for r in local.exported :
    {
      to = var.import_address_prefix != "" ? "${var.import_address_prefix}.${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}" : "${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}"
      id = r.import_id
    }
    if contains(keys(local._resource_label), r.terraform_resource)
  ]

  import_blocks_content = join("\n", [for b in local.import_blocks :
    "import {\n  to = ${b.to}\n  id = \"${b.id}\"\n}"
  ])

  # ── Moved blocks + YAML configs (convert_imports mode) ─────────────────

  # Single rules: one moved block per rule
  single_moved = [for r in local.exported :
    {
      from = "${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}"
      to   = "${var.rules_engine_address}.${local._resource_label[r.terraform_resource]}.this[\"imported_${r.terraform_resource}_${r.rule_id}\"]"
    }
    if contains(local.single_types, r.terraform_resource)
  ]

  # Stamp rules: one moved per individual stamp
  stamp_moved = [for r in local.exported :
    {
      from = "${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}"
      to   = "${var.rules_engine_address}.wallarm_rule_disable_stamp.this[\"imported_disable_stamp_${r.action_id}_${r.stamp}\"]"
    }
    if r.terraform_resource == "wallarm_rule_disable_stamp"
  ]

  # disable_attack_type: one moved per rule
  disable_at_moved = [for r in local.exported :
    {
      from = "${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}"
      to   = "${var.rules_engine_address}.wallarm_rule_disable_attack_type.this[\"imported_disable_attack_type_${r.action_id}_${r.attack_type}\"]"
    }
    if r.terraform_resource == "wallarm_rule_disable_attack_type"
  ]

  # vpatch: one moved per rule
  vpatch_moved = [for r in local.exported :
    {
      from = "${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}"
      to   = "${var.rules_engine_address}.wallarm_rule_vpatch.this[\"imported_vpatch_${r.action_id}_${r.attack_type}\"]"
    }
    if r.terraform_resource == "wallarm_rule_vpatch"
  ]

  # uploads: one moved per rule
  uploads_moved = [for r in local.exported :
    {
      from = "${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}"
      to   = "${var.rules_engine_address}.wallarm_rule_uploads.expanded[\"imported_uploads_${r.action_id}_${r.file_type}\"]"
    }
    if r.terraform_resource == "wallarm_rule_uploads"
  ]

  # parser_state: one moved per rule
  parser_moved = [for r in local.exported :
    {
      from = "${local._resource_label[r.terraform_resource]}.rule_${r.rule_id}"
      to   = "${var.rules_engine_address}.wallarm_rule_parser_state.expanded[\"imported_parser_state_${r.action_id}_${r.parser}\"]"
    }
    if r.terraform_resource == "wallarm_rule_parser_state"
  ]

  moved_blocks = concat(
    local.single_moved,
    local.stamp_moved,
    local.disable_at_moved,
    local.vpatch_moved,
    local.uploads_moved,
    local.parser_moved,
  )

  moved_blocks_content = join("\n", [for b in local.moved_blocks :
    "moved {\n  from = ${b.from}\n  to   = ${b.to}\n}"
  ])

  # ── YAML config data (for convert mode) ────────────────────────────────

  _defaults = {
    attack_type = ""
    mode        = ""
    regex       = ""
    regex_id    = 0
    regex_rule  = ""
    experimental = false
    parser      = ""
    file_type   = ""
    delay       = 0
    burst       = 0
    rate        = 0
    rsp_status  = 0
    time_unit   = ""
    size        = 0
    size_unit   = ""
    header_name = ""
    header_mode = ""
    header_values = []
    overlimit_time = 0
    introspection  = false
    debug_enabled  = false
    max_depth         = 0
    max_value_size_kb = 0
    max_doc_size_kb   = 0
    max_alias_size_kb = 0
    max_doc_per_batch = 0
    login_point     = []
    login_regex     = ""
    case_sensitive  = false
    cred_stuff_type = "default"
    threshold             = null
    reaction              = null
    enumerated_parameters = null
  }

  # YAML data for single rules
  yaml_single = { for r in local.exported :
    "imported_${r.terraform_resource}_${r.rule_id}" => {
      name          = "imported_${r.terraform_resource}_${r.rule_id}"
      resource_type = r.terraform_resource
      comment       = try(r.comment, "")
      path          = r.path
      domain        = r.domain
      instance      = r.instance
      method        = r.method
      scheme        = r.scheme
      proto         = r.proto
      query         = try(jsondecode(r.query_json), [])
      headers       = try(jsondecode(r.headers_json), [])
      point         = try(jsondecode(r.point_json), [])
      stamps        = []
      attack_types  = []
      file_types    = []
      parsers       = []
      attack_type   = try(r.attack_type, "")
      mode          = try(r.mode, "")
      stamp         = try(r.stamp, 0)
      metadata         = { origin = "import", rule_id = r.rule_id, import_id = r.import_id }
      _action_dir_name = r.action_dir_name
    }
    if contains(local.single_types, r.terraform_resource)
  }

  # YAML data for grouped stamp rules
  yaml_stamp_groups = { for action_id, rules in local.stamp_by_action :
    "imported_disable_stamp_${action_id}" => merge(local._defaults, {
      name          = "imported_disable_stamp_${action_id}"
      resource_type = "wallarm_rule_disable_stamp"
      comment       = try(rules[0].comment, "")
      path          = rules[0].path
      domain        = rules[0].domain
      instance      = rules[0].instance
      method        = rules[0].method
      scheme        = rules[0].scheme
      proto         = rules[0].proto
      query         = try(jsondecode(rules[0].query_json), [])
      headers       = try(jsondecode(rules[0].headers_json), [])
      point         = try(jsondecode(rules[0].point_json), [])
      stamps        = [for r in rules : r.stamp]
      attack_types  = []
      file_types    = []
      parsers       = []
      metadata         = { origin = "import", action_id = action_id, rule_ids = [for r in rules : r.rule_id] }
      _action_dir_name = rules[0].action_dir_name
    })
  }

  # YAML data for grouped disable_attack_type rules
  yaml_disable_at_groups = { for action_id, rules in local.disable_at_by_action :
    "imported_disable_attack_type_${action_id}" => merge(local._defaults, {
      name          = "imported_disable_attack_type_${action_id}"
      resource_type = "wallarm_rule_disable_attack_type"
      comment       = try(rules[0].comment, "")
      path          = rules[0].path
      domain        = rules[0].domain
      instance      = rules[0].instance
      method        = rules[0].method
      scheme        = rules[0].scheme
      proto         = rules[0].proto
      query         = try(jsondecode(rules[0].query_json), [])
      headers       = try(jsondecode(rules[0].headers_json), [])
      point         = try(jsondecode(rules[0].point_json), [])
      stamps        = []
      attack_types  = [for r in rules : r.attack_type]
      file_types    = []
      parsers       = []
      metadata         = { origin = "import", action_id = action_id, rule_ids = [for r in rules : r.rule_id] }
      _action_dir_name = rules[0].action_dir_name
    })
  }

  # YAML data for grouped vpatch rules
  yaml_vpatch_groups = { for action_id, rules in local.vpatch_by_action :
    "imported_vpatch_${action_id}" => merge(local._defaults, {
      name          = "imported_vpatch_${action_id}"
      resource_type = "wallarm_rule_vpatch"
      comment       = try(rules[0].comment, "")
      path          = rules[0].path
      domain        = rules[0].domain
      instance      = rules[0].instance
      method        = rules[0].method
      scheme        = rules[0].scheme
      proto         = rules[0].proto
      query         = try(jsondecode(rules[0].query_json), [])
      headers       = try(jsondecode(rules[0].headers_json), [])
      point         = try(jsondecode(rules[0].point_json), [])
      stamps        = []
      attack_types  = [for r in rules : r.attack_type]
      file_types    = []
      parsers       = []
      metadata         = { origin = "import", action_id = action_id, rule_ids = [for r in rules : r.rule_id] }
      _action_dir_name = rules[0].action_dir_name
    })
  }

  # YAML data for grouped uploads rules
  yaml_uploads_groups = { for action_id, rules in local.uploads_by_action :
    "imported_uploads_${action_id}" => merge(local._defaults, {
      name          = "imported_uploads_${action_id}"
      resource_type = "wallarm_rule_uploads"
      comment       = try(rules[0].comment, "")
      path          = rules[0].path
      domain        = rules[0].domain
      instance      = rules[0].instance
      method        = rules[0].method
      scheme        = rules[0].scheme
      proto         = rules[0].proto
      query         = try(jsondecode(rules[0].query_json), [])
      headers       = try(jsondecode(rules[0].headers_json), [])
      point         = try(jsondecode(rules[0].point_json), [])
      stamps        = []
      attack_types  = []
      file_types    = [for r in rules : r.file_type]
      parsers       = []
      metadata         = { origin = "import", action_id = action_id, rule_ids = [for r in rules : r.rule_id] }
      _action_dir_name = rules[0].action_dir_name
    })
  }

  # YAML data for grouped parser_state rules
  yaml_parser_groups = { for action_id, rules in local.parser_by_action :
    "imported_parser_state_${action_id}" => merge(local._defaults, {
      name          = "imported_parser_state_${action_id}"
      resource_type = "wallarm_rule_parser_state"
      comment       = try(rules[0].comment, "")
      path          = rules[0].path
      domain        = rules[0].domain
      instance      = rules[0].instance
      method        = rules[0].method
      scheme        = rules[0].scheme
      proto         = rules[0].proto
      query         = try(jsondecode(rules[0].query_json), [])
      headers       = try(jsondecode(rules[0].headers_json), [])
      point         = try(jsondecode(rules[0].point_json), [])
      stamps        = []
      attack_types  = []
      file_types    = []
      parsers       = [for r in rules : r.parser]
      metadata         = { origin = "import", action_id = action_id, rule_ids = [for r in rules : r.rule_id] }
      _action_dir_name = rules[0].action_dir_name
    })
  }

  all_yaml_configs = merge(
    local.yaml_single,
    local.yaml_stamp_groups,
    local.yaml_disable_at_groups,
    local.yaml_vpatch_groups,
    local.yaml_uploads_groups,
    local.yaml_parser_groups,
  )

  # ── Unique action directories for .action.yaml generation ──────────────
  _import_actions_grouped = {
    for r in local.exported :
    r.action_dir_name => {
      dir_name        = r.action_dir_name
      conditions_hash = r.conditions_hash
      action_id       = r.action_id
      conditions      = try(jsondecode(r.action_json), [])
      path            = r.path
      domain          = r.domain
      instance        = r.instance
    }...
  }

  _unique_import_actions = {
    for dir, entries in local._import_actions_grouped :
    dir => entries[0]
  }
}

# ─── Write YAML configs (convert_imports mode) ───────────────────────────────

resource "local_file" "yaml_config" {
  for_each = { for k, v in local.all_yaml_configs : k => v if var.convert_imports }

  filename        = "${var.import_config_dir}/${each.value._action_dir_name}/${each.key}.yaml"
  file_permission = "0644"

  content = yamlencode({
    for k, v in each.value : k => v
    if !startswith(k, "_")
  })
}

# ─── Generate .action.yaml for each unique action directory ──────────────────

resource "local_file" "action_config" {
  for_each = { for k, v in local._unique_import_actions : k => v if var.convert_imports }

  filename        = "${var.import_config_dir}/${each.value.dir_name}/.action.yaml"
  file_permission = "0644"

  content = yamlencode({
    conditions      = each.value.conditions
    conditions_hash = each.value.conditions_hash
    action_id       = each.value.action_id
    action_path     = each.value.path
    action_domain   = each.value.domain
    action_instance = each.value.instance
  })
}
