# ─── Fetch hits (only on first apply per request_id) ─────────────────────────
# The data source runs only when fetch_hits=true. After the first apply,
# terraform_data.hits_state persists the result with ignore_changes.
# Subsequent applies read from state — no API calls.

data "wallarm_hits" "this" {
  count        = var.fetch_hits ? 1 : 0
  client_id    = var.client_id
  request_id   = var.request_id
  mode         = var.mode
  attack_types = length(var.attack_types) > 0 ? var.attack_types : null
  time         = var.time
}

# ─── Aggregation ──────────────────────────────────────────────────────────────

locals {
  raw_hits = try(data.wallarm_hits.this[0].hits, [])

  aggregated = {
    action_hash       = try(data.wallarm_hits.this[0].action_hash, "")
    action_dir_name   = try(data.wallarm_hits.this[0].action_dir_name, "")
    action_conditions = try(data.wallarm_hits.this[0].action_conditions, [])
    domain            = try(local.raw_hits[0].domain, "")
    path              = try(local.raw_hits[0].path, "")
    poolid            = try(local.raw_hits[0].poolid, 0)

    points = {
      for ph in distinct([for h in local.raw_hits : h.point_hash]) :
      ph => {
        point_wrapped = [for h in local.raw_hits : h.point_wrapped if h.point_hash == ph][0]
        stamps        = sort(distinct(flatten([for h in local.raw_hits : try(h.stamps, []) if h.point_hash == ph])))
        attack_types  = distinct([for h in local.raw_hits : try(h.type, "") if h.point_hash == ph && try(h.type, "") != ""])
        attack_ids    = distinct([for h in local.raw_hits : try(h.attack_id, "") if h.point_hash == ph && try(h.attack_id, "") != ""])
        hit_ids       = distinct([for h in local.raw_hits : try(h.id, "") if h.point_hash == ph && try(h.id, "") != ""])
      }
    }
  }
}

# ─── Persist aggregated data in Terraform state ──────────────────────────────
# Write-once: ignore_changes keeps the original data from the first apply.

resource "terraform_data" "hits_state" {
  input = local.aggregated

  lifecycle {
    ignore_changes = [input]
  }
}

# ─── Effective values ─────────────────────────────────────────────────────────
# Always use local.aggregated. When fetch_hits=true, the data source runs and
# aggregated has data. When fetch_hits=false, aggregated is empty (data source
# count=0) → empty rules, which is correct because YAML files on disk drive
# the rules_engine via fileset. The terraform_data persists data for reference
# but is NOT used as input to rule generation (avoids unknown value issues).

locals {
  effective = local.aggregated
}

# ─── Generate rule objects ────────────────────────────────────────────────────
# Converts aggregated hits into universal rule objects for the rules_engine.
# One rule object per (point_hash, rule_type) combination.
# Files are placed in the action_dir_name subdirectory of config_dir.

locals {
  action_dir = "${var.config_dir}/${local.effective.action_dir_name}"

  rules = flatten([
    for ph, cfg in local.effective.points : [
      for rt in var.rule_types : {
        name          = "hits_${substr(ph, 0, 8)}_${rt}"
        resource_type = rt == "disable_stamp" ? "wallarm_rule_disable_stamp" : "wallarm_rule_disable_attack_type"
        comment       = "FP from request ${var.request_id}"

        # Scope — full path from hit
        path     = local.effective.path
        domain   = local.effective.domain
        instance = local.effective.poolid != 0 ? tostring(local.effective.poolid) : ""
        method   = ""
        scheme   = ""
        proto    = ""
        query    = []
        headers  = []

        # Detection point
        point = cfg.point_wrapped

        # Multi-value fields
        stamps       = rt == "disable_stamp" ? cfg.stamps : []
        attack_types = rt == "disable_attack_type" ? cfg.attack_types : []
        file_types   = []
        parsers      = []

        # Unused rule-specific fields (defaults)
        attack_type    = ""
        mode           = ""
        regex          = ""
        regex_id       = 0
        regex_rule     = ""
        experimental   = false
        parser         = ""
        file_type      = ""
        delay          = 0
        burst          = 0
        rate           = 0
        rsp_status     = 0
        time_unit      = ""
        size           = 0
        size_unit      = ""
        header_name    = ""
        header_mode    = ""
        header_values  = []
        overlimit_time = 0
        introspection     = false
        debug_enabled     = false
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

        # Metadata (informational, preserved in YAML)
        metadata = {
          origin       = "hit"
          request_id   = var.request_id
          point_hash   = ph
          hit_ids      = try(cfg.hit_ids, [])
          attack_types = cfg.attack_types
        }

        # Action data for .action.yaml generation
        _action_conditions = local.effective.action_conditions
        _action_hash       = local.effective.action_hash

        # Internal: where to write the YAML config
        _config_dir = local.action_dir
      }
    ]
  ])
}
