output "rule_ids" {
  description = "Map of custom rule keys to their created rule IDs (all resource types)"
  value = merge(
    { for k, v in wallarm_rule_binary_data.this : k => v.id },
    { for k, v in wallarm_rule_masking.this : k => v.id },
    { for k, v in wallarm_rule_disable_attack_type.this : k => v.id },
    { for k, v in wallarm_rule_disable_stamp.this : k => v.id },
    { for k, v in wallarm_rule_vpatch.this : k => v.id },
    { for k, v in wallarm_rule_uploads.this : k => v.id },
    { for k, v in wallarm_rule_ignore_regex.this : k => v.id },
    { for k, v in wallarm_rule_parser_state.this : k => v.id },
    { for k, v in wallarm_rule_regex.this : k => v.id },
    { for k, v in wallarm_rule_file_upload_size_limit.this : k => v.id },
    { for k, v in wallarm_rule_rate_limit.this : k => v.id },
    { for k, v in wallarm_rule_credential_stuffing_point.this : k => v.id },
    { for k, v in wallarm_rule_credential_stuffing_regex.this : k => v.id },
    { for k, v in wallarm_rule_mode.this : k => v.id },
    { for k, v in wallarm_rule_set_response_header.this : k => v.id },
    { for k, v in wallarm_rule_overlimit_res_settings.this : k => v.id },
    { for k, v in wallarm_rule_graphql_detection.this : k => v.id },
    { for k, v in wallarm_rule_brute.this : k => v.id },
    { for k, v in wallarm_rule_bruteforce_counter.this : k => v.id },
    { for k, v in wallarm_rule_dirbust_counter.this : k => v.id },
    { for k, v in wallarm_rule_bola.this : k => v.id },
    { for k, v in wallarm_rule_bola_counter.this : k => v.id },
    { for k, v in wallarm_rule_enum.this : k => v.id },
    { for k, v in wallarm_rule_rate_limit_enum.this : k => v.id },
    { for k, v in wallarm_rule_forced_browsing.this : k => v.id },
  )
}

output "config_files" {
  description = "Paths to generated config files"
  value = { for k, v in local_file.rule_config : k => v.filename }
}
