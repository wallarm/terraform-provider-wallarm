output "action_hash" {
  description = "SHA256 hash of action conditions (Ruby-compatible)"
  value       = local.effective.action_hash
}

output "action_dir_name" {
  description = "Computed directory name for this action scope"
  value       = local.effective.action_dir_name
}

output "domain" {
  description = "Request domain from hit"
  value       = local.effective.domain
}

output "path" {
  description = "Request path from hit"
  value       = local.effective.path
}

output "poolid" {
  description = "Application pool ID from hit"
  value       = local.effective.poolid
}

output "points" {
  description = "Map of point_hash => { point_wrapped, stamps, attack_types, attack_ids }"
  value       = local.effective.points
}

output "has_hits" {
  description = "Whether any hits are available"
  value       = length(try(local.effective.points, {})) > 0
}

output "rules" {
  description = "List of rule objects in universal format for the rules_engine"
  value       = local.rules
}
