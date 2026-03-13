variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "is_importing" {
  type        = bool
  default     = false
  description = "Must be true to activate rules import functionality."
}
