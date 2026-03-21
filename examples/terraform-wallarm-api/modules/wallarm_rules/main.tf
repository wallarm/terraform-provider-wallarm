# ─── Fetch hits per request_id ─────────────────────────────────────────────────

module "hits" {
  for_each = var.requests
  source   = "./modules/hits_fetcher"

  client_id  = var.client_id
  request_id = each.key
  mode       = var.hits_mode

  # Gate: skip API fetch when fp_rules config files already exist for this request_id.
  # First apply (no configs yet) → true  → data source fetches from API.
  # Subsequent  (configs exist)  → false → reads persisted data from terraform state.
  fetch_hits = length(try(fileset("${var.fp_config_dir}/${each.key}", "*.yaml"), toset([]))) == 0
}

# ─── Convert hits to universal rule objects ───────────────────────────────────

module "hits_generator" {
  for_each = var.requests
  source   = "./modules/hits_generator"

  request_id = each.key
  rule_types = each.value
  domain     = module.hits[each.key].domain
  path       = module.hits[each.key].path
  poolid     = module.hits[each.key].poolid
  points     = module.hits[each.key].points
  action     = module.hits[each.key].action
  config_dir = var.fp_config_dir
}

# ─── Import / Convert existing rules ─────────────────────────────────────────
#
# Standard import (is_importing=true):
#   1. terraform apply -var='is_importing=true'    → wallarm_rule_imports.tf
#   2. terraform plan -generate-config-out=imported.tf
#   3. terraform apply                             → imports into state
#
# Convert to YAML (convert_imports=true, after standard import):
#   4. terraform apply -var='convert_imports=true'  → YAML + moved blocks
#   5. Remove imported.tf and wallarm_rule_imports.tf
#   6. terraform apply                              → state migrates to rules_engine

module "import_generator" {
  source = "./modules/import_generator"

  client_id             = var.client_id
  is_importing          = var.is_importing
  convert_imports       = var.convert_imports
  rule_types            = var.import_rule_types
  import_address_prefix = var.import_address_prefix
  rules_engine_address  = var.rules_engine_address
  import_config_dir     = var.import_config_dir
}

resource "local_file" "import_blocks" {
  count           = var.is_importing && module.import_generator.import_blocks != "" ? 1 : 0
  filename        = "${path.root}/wallarm_rule_imports.tf"
  file_permission = "0644"
  content         = module.import_generator.import_blocks
}

resource "local_file" "moved_blocks" {
  count           = var.convert_imports && module.import_generator.moved_blocks != "" ? 1 : 0
  filename        = "${path.root}/wallarm_moved_blocks.tf"
  file_permission = "0644"
  content         = module.import_generator.moved_blocks
}

# ─── Unified rules engine ────────────────────────────────────────────────────
# Reads YAML configs from custom_rules/, fp_rules/, and import_rules/ directories.

module "rules" {
  source = "./modules/rules_engine"

  client_id = var.client_id

  config_dirs = [
    var.config_dir,           # custom_rules/
    var.fp_config_dir,        # fp_rules/
    var.import_config_dir,    # import_rules/ (YAML from converted imports)
  ]

  generated_rules = flatten([
    for req_id in keys(var.requests) : module.hits_generator[req_id].rules
  ])
}
