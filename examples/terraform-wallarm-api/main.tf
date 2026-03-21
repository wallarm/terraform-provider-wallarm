# ─── Create rules ────────────────────────────────────────────────────────────

module "wallarm_rules" {
  source = "./modules/wallarm_rules"

  client_id         = var.client_id
  config_dir        = "${path.root}/rules_config/custom_rules"
  fp_config_dir     = "${path.root}/rules_config/fp_rules"
  import_config_dir = "${path.root}/rules_config/import_rules"

  # Optional
  requests          = var.requests
  hits_mode         = var.hits_mode
  is_importing      = var.is_importing
  import_rule_types = var.import_rule_types
}

# ─── Import existing resources from API ──────────────────────────────────────

module "wallarm_import" {
  source = "./modules/wallarm_import"

  client_id          = var.client_id
  is_importing       = var.is_importing
  subnet_import_mode = var.subnet_import_mode
}

# ─── Outputs ─────────────────────────────────────────────────────────────────

output "rule_ids" {
  value = module.wallarm_rules.rule_ids
}

# ─── Import outputs ──────────────────────────────────────────────────────────

output "total_rules" {
  value = length(module.wallarm_import.all_rules[*]) > 0 ? length(module.wallarm_import.all_rules) : null
}

output "imported_rules" {
  value = length(module.wallarm_import.all_rules[*]) > 0 ? module.wallarm_import.all_rules : null
}

output "imported_applications" {
  description = "All imported applications from the Wallarm API (when is_importing=true)"
  value       = module.wallarm_import.all_applications
}

output "imported_application_count" {
  description = "Number of imported applications (when is_importing=true)"
  value       = module.wallarm_import.application_count
}



# # Import

#   1. terraform init && terraform apply -auto-approve -var='is_importing=true'
#   2. terraform plan -var='is_importing=true' -generate-config-out=imported_rules.tf
#   3. terraform apply --auto-approve
