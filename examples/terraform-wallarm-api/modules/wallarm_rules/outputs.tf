output "rule_ids" {
  description = "Flat map of all created rule IDs"
  value       = module.rules.rule_ids
}

output "config_files" {
  description = "Paths to generated YAML config files (from hits)"
  value       = module.rules.config_files
}

output "import_blocks" {
  description = "Import block content (written to wallarm_rule_imports.tf when is_importing=true)"
  value       = module.import_generator.import_blocks
}

output "moved_blocks" {
  description = "Moved block content (written to wallarm_moved_blocks.tf when convert_imports=true)"
  value       = module.import_generator.moved_blocks
}
