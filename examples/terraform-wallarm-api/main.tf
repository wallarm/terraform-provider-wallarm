# ─── Create rules ────────────────────────────────────────────────────────────

module "wallarm_rules" {
  source = "./modules/wallarm_rules"

  client_id           = var.client_id
  requests            = var.requests
  hits_mode           = var.hits_mode
  custom_rules        = var.custom_rules
  config_dir          = "${path.root}/rules_config"
  fp_config_dir       = "${path.root}/fp-rules-configs"
  config_format       = var.config_format
}

# ─── Import existing resources from API ──────────────────────────────────────

module "wallarm_import" {
  source = "./modules/wallarm_import"

  client_id    = var.client_id
  is_importing = var.is_importing
}

# ─── Outputs ─────────────────────────────────────────────────────────────────

output "rule_ids_by_request" {
  description = "Rule IDs grouped by request_id"
  value       = module.wallarm_rules.rule_ids_by_request
}

output "custom_rule_ids" {
  description = "Rule IDs from custom rules defined in variables"
  value       = module.wallarm_rules.custom_rule_ids
}

output "all_rule_ids" {
  description = "Flat map of all created rule IDs across every request_id, rule type, and custom rules"
  value       = module.wallarm_rules.all_rule_ids
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

#   1. terraform init && terraform apply -auto-approve -var='is_importing=true'
#   2. terraform plan -var='is_importing=true' -generate-config-out=imported_rules.tf
#   3. terraform apply --auto-approve
