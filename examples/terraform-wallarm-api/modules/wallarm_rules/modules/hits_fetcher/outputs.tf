output "action" {
  description = "Rule action conditions derived from the hit (host, path, instance)"
  value       = local.effective.action
}

output "action_hash" {
  description = "SHA256 hash of sorted action conditions"
  value       = local.effective.action_hash
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
