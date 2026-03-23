# ─── Fetch hits and generate rule objects ─────────────────────────────────────

module "hits" {
  for_each = var.requests
  source   = "./modules/hits_fetcher"

  client_id  = var.client_id
  request_id = each.key
  rule_types = each.value
  config_dir = var.configs_dir
  mode       = var.hits_mode

  # Gate: fetch from API when explicitly requested or when no hits YAML files exist.
  # After first apply, hits_*.yaml files exist → fetch skipped → reads from state.
  # Set var.fetch_hits=true to force re-fetch (e.g., when adding new request_ids).
  fetch_hits = var.fetch_hits || length(try(fileset(var.configs_dir, "**/hits_*.yaml"), toset([]))) == 0
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
  import_config_dir     = var.configs_dir
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
# Single configs_dir with action subdirectories containing .action.yaml + rule YAMLs.

module "rules" {
  source = "./modules/rules_engine"

  client_id = var.client_id

  config_dirs = [var.configs_dir]

  generated_rules = flatten([
    for req_id in keys(var.requests) : module.hits[req_id].rules
  ])
}

# ─── Action discovery (separate step) ────────────────────────────────────────
# Discovers new action scopes from API and generates .action.yaml files.
# Run with: terraform apply -var='discover_actions=true'
# This is a second-apply step — rules must exist first.

data "wallarm_actions" "discovery" {
  count     = var.discover_actions ? 1 : 0
  client_id = var.client_id
}

locals {
  known_action_hashes = { for hash, cfg in module.rules.action_map : hash => true }

  discovered_actions = var.discover_actions ? {
    for a in try(data.wallarm_actions.discovery[0].actions, []) :
    a.conditions_hash => a
    if !contains(keys(local.known_action_hashes), a.conditions_hash)
  } : {}
}

resource "local_file" "discovered_action_config" {
  for_each = local.discovered_actions

  filename        = "${var.configs_dir}/${each.value.dir_name}/.action.yaml"
  file_permission = "0644"

  content = yamlencode({
    conditions      = each.value.conditions
    conditions_hash = each.value.conditions_hash
    action_id       = each.value.action_id
    action_path     = each.value.endpoint_path
    action_domain   = each.value.endpoint_domain
    action_instance = each.value.endpoint_instance
  })

  lifecycle {
    ignore_changes = [content]
  }
}
