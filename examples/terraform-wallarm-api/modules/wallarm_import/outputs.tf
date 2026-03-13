output "all_rules" {
  description = "All imported rules from the Wallarm API"
  value       = module.import_rules.all_rules
}

output "all_applications" {
  description = "All imported applications from the Wallarm API"
  value       = module.import_applications.all_applications
}

output "application_count" {
  description = "Number of imported applications"
  value       = module.import_applications.application_count
}

output "ip_list_entries" {
  description = "IP list entries by list type"
  value = {
    denylist  = module.import_ip_lists.denylist_entries
    allowlist = module.import_ip_lists.allowlist_entries
    graylist  = module.import_ip_lists.graylist_entries
  }
}

output "ip_list_counts" {
  description = "Number of IP list entries by list type"
  value       = module.import_ip_lists.entry_counts
}
