# ─── Hints Cache ──────────────────────────────────────────────────────────────
# Persistent index of all rules from the Wallarm API.
# Stores references only — no rule-specific data fields.
# Full rule data is fetched ephemerally during refresh for import_rules output.
#
# is_managed is NOT computed here (would create circular dependency with rules_engine).
# The parent module computes it by comparing hints_index against rule_ids.

data "wallarm_rules" "all" {
  count     = var.refresh ? 1 : 0
  client_id = var.client_id
  type      = var.rule_types
}

# ─── Build index from API response ──────────────────────────────────────────

locals {
  _exported = var.refresh ? try(data.wallarm_rules.all[0].rules_export, []) : []

  # Index entries: references + expansion suffix
  _fetched_index = {
    for r in local._exported :
    tostring(r.rule_id) => {
      rule_id            = r.rule_id
      action_id          = r.action_id
      import_id          = r.import_id
      terraform_resource = r.terraform_resource
      conditions_hash    = r.conditions_hash
      action_dir_name    = r.action_dir_name
      suffix             = coalesce(
        r.stamp != 0 ? tostring(r.stamp) : "",
        r.attack_type != "" ? r.attack_type : "",
        r.file_type != "" ? r.file_type : "",
        r.parser != "" ? r.parser : "",
        ""
      )
    }
  }
}

# ─── Persist index in state ──────────────────────────────────────────────────

resource "terraform_data" "hints_index" {
  input = var.refresh ? local._fetched_index : null

  lifecycle {
    ignore_changes = [input]
  }
}

# ─── Effective index ─────────────────────────────────────────────────────────

locals {
  hints_index = var.refresh ? local._fetched_index : coalesce(try(terraform_data.hints_index.output, null), {})
}

# ─── Import rules (ephemeral — full data, only during refresh) ───────────────

locals {
  import_rules = [
    for r in local._exported : {
      name                 = "imported_${r.terraform_resource}_${r.rule_id}"
      resource_type        = r.terraform_resource
      comment              = coalesce(r.comment, "Managed by Terraform")
      variativity_disabled = true

      # No scope fields — imported rules use action blocks from Read
      path     = ""
      domain   = ""
      instance = ""
      method   = ""
      scheme   = ""
      proto    = ""
      query    = []
      headers  = []
      point    = try(jsondecode(r.point_json), [])

      # Expandable fields — single values
      stamps       = r.stamp != 0 ? [r.stamp] : []
      attack_types = r.attack_type != "" ? [r.attack_type] : []
      file_types   = r.file_type != "" ? [r.file_type] : []
      parsers      = r.parser != "" ? [r.parser] : []

      # Rule-specific
      attack_type    = r.attack_type
      mode           = r.mode
      regex          = r.regex
      regex_id       = r.regex_id
      regex_rule     = ""
      experimental   = r.experimental
      parser         = r.parser
      file_type      = r.file_type
      delay          = r.delay
      burst          = r.burst
      rate           = r.rate
      rsp_status     = r.rsp_status
      time_unit      = r.time_unit
      size           = r.size
      size_unit      = r.size_unit
      header_name    = r.header_name
      header_mode    = ""
      header_values  = try(jsondecode(r.header_values_json), [])
      overlimit_time = r.overlimit_time

      # GraphQL
      introspection     = r.introspection
      debug_enabled     = r.debug_enabled
      max_depth         = r.max_depth
      max_value_size_kb = r.max_value_size_kb
      max_doc_size_kb   = r.max_doc_size_kb
      max_alias_size_kb = r.max_aliases_size_kb
      max_doc_per_batch = r.max_doc_per_batch

      # Credential stuffing
      login_point     = try(jsondecode(r.login_point_json), [])
      login_regex     = r.login_regex
      case_sensitive  = r.case_sensitive
      cred_stuff_type = r.cred_stuff_type

      # Complex nested
      threshold             = try(jsondecode(r.threshold_json), null)
      reaction              = try(jsondecode(r.reaction_json), null)
      enumerated_parameters = try(jsondecode(r.enumerated_parameters_json), null)

      # Metadata
      metadata = {
        origin  = "import"
        rule_id = r.rule_id
      }

      # Action data
      _action_conditions = []
      _action_hash       = r.conditions_hash
      _config_dir        = "${var.configs_dir}/${r.action_dir_name}"
    }
  ]

  # ─── Import blocks ──────────────────────────────────────────────────────
  _addr_prefix = var.rules_engine_address != "" ? "${var.rules_engine_address}." : ""

  import_blocks_content = join("\n", [
    for r in local._exported :
    "import {\n  to = ${local._addr_prefix}${
      r.terraform_resource == "wallarm_rule_uploads" ? "wallarm_rule_uploads.expanded" :
      r.terraform_resource == "wallarm_rule_parser_state" ? "wallarm_rule_parser_state.expanded" :
      "${r.terraform_resource}.this"
    }[\"imported_${r.terraform_resource}_${r.rule_id}${
      coalesce(
        r.stamp != 0 ? "_${r.stamp}" : "",
        r.attack_type != "" ? "_${r.attack_type}" : "",
        r.file_type != "" ? "_${r.file_type}" : "",
        r.parser != "" ? "_${r.parser}" : "",
        ""
      )
    }\"]\n  id = \"${r.import_id}\"\n}"
  ])

  # ─── Unique action directories ─────────────────────────────────────────
  _action_dirs_grouped = {
    for id, r in local.hints_index :
    r.conditions_hash => {
      action_id       = r.action_id
      conditions_hash = r.conditions_hash
      dir_name        = r.action_dir_name
    }...
  }

  action_dirs = {
    for hash, entries in local._action_dirs_grouped :
    hash => entries[0]
  }
}
