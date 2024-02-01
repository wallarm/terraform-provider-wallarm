resource "wallarm_rules_settings" "rules_settings" {
  client_id = 123
  min_lom_format = 50
	max_lom_format = 54
	max_lom_size = 10240
	lom_disabled = false
	lom_compilation_delay = 0
	rules_snapshot_enabled = true
	rules_snapshot_max_count = 5
	rules_manipulation_locked = false
	heavy_lom = false
	parameters_count_weight = 6
	path_variativity_weight = 6
	pii_weight = 8
	request_content_weight = 6
	open_vulns_weight = 9
	serialized_data_weight = 6
	risk_score_algo = "maximum"
	pii_fallback = false
}
