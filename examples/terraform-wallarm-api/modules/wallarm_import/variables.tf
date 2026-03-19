variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Must be true to activate rules import functionality."
}

variable "subnet_import_mode" {
  type        = string
  default     = "grouped"
  description = "How to import subnet entries: 'grouped' merges by expired_at into resources of max 1000 IPs; 'individual' creates one resource per IP."
}
