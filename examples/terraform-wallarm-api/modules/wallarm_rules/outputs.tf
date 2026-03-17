output "rule_ids_by_request" {
  description = "Rule IDs grouped by request_id (from fp_rules)"
  value = {
    for req_id in keys(local.rules_by_request) : req_id => try(module.fp_rules[req_id].rule_ids, {})
  }
}

output "custom_rule_ids" {
  description = "Rule IDs from custom rules defined in variables"
  value       = module.custom_rules.rule_ids
}

output "all_rule_ids" {
  description = "Flat map of all created rule IDs across every request_id, rule type, and custom rules"
  value = merge(concat(
    [for req_id, mod in module.fp_rules : mod.rule_ids],
    [module.custom_rules.rule_ids],
  )...)
}
