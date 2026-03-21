# ─── Hits Generator ───────────────────────────────────────────────────────────
# Converts hits_fetcher output into universal rule objects for the rules_engine.
# One rule object per (point_hash, rule_type) combination.

locals {
  req_prefix = substr(var.request_id, 0, 8)

  # Cartesian product: point_hash x rule_type → one rule object each
  rules = flatten([
    for ph, cfg in var.points : [
      for rt in var.rule_types : {
        name          = "${local.req_prefix}_${substr(ph, 0, 8)}_${rt}"
        resource_type = rt == "disable_stamp" ? "wallarm_rule_disable_stamp" : "wallarm_rule_disable_attack_type"
        comment       = "FP from request ${var.request_id}"

        # Scope — full path from hit (no wildcards), user can add wildcards later
        path     = var.path
        domain   = var.domain
        instance = var.poolid != 0 ? tostring(var.poolid) : ""
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
          source       = "hits"
          request_id   = var.request_id
          point_hash   = ph
          hit_ids      = try(cfg.hit_ids, [])
          attack_types = cfg.attack_types
        }

        # Built action conditions (for reference HCL)
        _action = var.action

        # Internal: where to write the YAML config
        _config_dir = "${var.config_dir}/${var.request_id}"
      }
    ]
  ])
}
