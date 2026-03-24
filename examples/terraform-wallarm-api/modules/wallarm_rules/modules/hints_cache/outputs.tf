output "hints_index" {
  description = "Persistent index: rule_id → { action_id, import_id, terraform_resource, conditions_hash, action_dir_name, suffix }"
  value       = local.hints_index
}

output "import_rules" {
  description = "All rules in generated_rules format (ephemeral, only during refresh)"
  value       = local.import_rules
}

output "import_blocks_content" {
  description = "Import block HCL content for local_file"
  value       = local.import_blocks_content
}

output "action_dirs" {
  description = "Unique action directories from indexed rules"
  value       = local.action_dirs
}

output "rule_count" {
  description = "Total rules in index"
  value       = length(local.hints_index)
}
