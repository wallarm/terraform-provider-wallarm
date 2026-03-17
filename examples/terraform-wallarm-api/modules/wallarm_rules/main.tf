# ─── Fetch hits per request_id ─────────────────────────────────────────────────

module "hits" {
  for_each = var.requests
  source   = "./modules/hits_fetcher"

  client_id  = var.client_id
  request_id = each.key
  mode       = var.hits_mode

  # "request" (default, existing behavior) or "attack"
  # terraform apply -var='hits_mode=attack'

  # Gate: skip API fetch when fp_rules config files already exist for this request_id.
  # First apply (no configs yet) → true  → data source fetches from API.
  # Subsequent  (configs exist)  → false → reads persisted data from terraform state.
  fetch_hits = length(try(fileset("${var.fp_config_dir}/${each.key}", "*.yaml"), toset([]))) == 0
}

# ─── Build rules map ─────────────────────────────────────────────────────────

locals {
  rules_by_request = {
    for req_id, rule_types in var.requests : req_id => {
      request_id = req_id
      rule_types = rule_types
      action     = module.hits[req_id].action
      domain     = module.hits[req_id].domain
      path       = module.hits[req_id].path
      poolid     = module.hits[req_id].poolid
      points     = module.hits[req_id].points
    }
  }
}

# ─── Create rules from fetched hits (false positive rules) ──────────────────

module "fp_rules" {
  for_each = local.rules_by_request
  source   = "./modules/fp_rules"

  client_id  = var.client_id
  request_id = each.value.request_id
  rule_types = each.value.rule_types
  action     = each.value.action
  domain     = each.value.domain
  path       = each.value.path
  poolid     = each.value.poolid
  points     = each.value.points
  config_dir = var.fp_config_dir
}

# ─── Custom rules from variables ─────────────────────────────────────────────

module "custom_rules" {
  source = "./modules/custom_rules"

  client_id     = var.client_id
  rules         = var.custom_rules
  config_dir    = var.config_dir
  config_format = var.config_format
}
