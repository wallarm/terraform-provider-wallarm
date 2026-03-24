output "rule_ids" {
  description = "Flat map of all created rule IDs"
  value       = module.rules.rule_ids
}

output "config_files" {
  description = "Paths to generated YAML config files"
  value       = module.rules.config_files
}

output "hints_index" {
  description = "Persistent rules index with is_managed computed from rule_ids"
  value = {
    for id, entry in module.hints_cache.hints_index :
    id => merge(entry, {
      is_managed = anytrue([
        contains(keys(module.rules.rule_ids), "imported_${entry.terraform_resource}_${id}"),
        contains(keys(module.rules.rule_ids), "imported_${entry.terraform_resource}_${id}_${try(entry.suffix, "")}"),
      ])
    })
  }
}

output "unmanaged_count" {
  description = "Number of rules in API not yet managed by Terraform"
  value = length([
    for id, entry in module.hints_cache.hints_index : id
    if !anytrue([
      contains(keys(module.rules.rule_ids), "imported_${entry.terraform_resource}_${id}"),
      contains(keys(module.rules.rule_ids), "imported_${entry.terraform_resource}_${id}_${try(entry.suffix, "")}"),
    ])
  ])
}

output "action_map" {
  description = "Action map from rules_engine"
  value       = module.rules.action_map
}
