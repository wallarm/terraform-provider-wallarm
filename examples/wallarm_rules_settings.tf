resource "wallarm_rules_settings" "rules_settings" {
  client_id = 123
  min_lom_format = null # Default value - recommended
  max_lom_format = null # Default value - recommended
  max_lom_size = 100000000 # 100Mb limit for large rulesets, 10Mb for small
  lom_disabled = false # Set true to stop rules compilation
  lom_compilation_delay = 10 # Recommended to avoid resource create/update/delete operations block by snapshot
  rules_snapshot_enabled = true # Automatic rules backups
}
