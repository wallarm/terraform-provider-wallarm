output "rule_ids" {
  description = "Map of rule keys to their created rule IDs (all types)"
  value = merge(
    { for k, v in wallarm_rule_disable_stamp.this : k => v.id },
    { for k, v in wallarm_rule_disable_attack_type.this : k => v.id },
  )
}

output "config_files" {
  description = "Paths to generated config files"
  value = {
    points = { for k, v in local_file.rule_config : k => v.filename }
  }
}
