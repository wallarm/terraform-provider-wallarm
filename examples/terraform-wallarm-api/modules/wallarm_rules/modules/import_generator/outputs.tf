output "import_blocks" {
  description = "Import block content for wallarm_rule_imports.tf"
  value       = local.import_blocks_content
}

output "moved_blocks" {
  description = "Moved block content for migrating state to rules_engine"
  value       = local.moved_blocks_content
}
