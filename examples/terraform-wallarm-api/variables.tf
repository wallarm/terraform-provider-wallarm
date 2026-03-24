variable "api_host" {
  type = string
}

variable "api_token" {
  type = string
}

variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "requests" {
  type        = map(list(string))
  default     = {}
  description = "Static map of request_id => list of rule types. Example: { \"abc123\" = [\"disable_stamp\", \"disable_attack_type\"] }"
}

variable "hits_mode" {
  type        = string
  default     = "request"
  description = "Fetch mode for hits data source: 'request' (direct hits only) or 'attack' (expand to all related hits by attack_id)"
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Import IP lists, applications, and other non-rule resources."
}

variable "import_rules" {
  type        = bool
  default     = false
  description = "Import rules from API via hints_cache."
}

variable "subnet_import_mode" {
  type        = string
  default     = "grouped"
  description = "How to import subnet entries: 'grouped' merges by expired_at into resources of max 1000 IPs; 'individual' creates one resource per IP."
}

variable "import_rule_types" {
  type    = list(string)
  default = []
}

variable "discover_actions" {
  type        = bool
  default     = false
  description = "Discover new action scopes from API and generate .action.yaml files."
}

variable "fetch_hits" {
  type        = bool
  default     = false
  description = "Force re-fetch hits from API even if YAML configs exist."
}
