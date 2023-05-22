variable "api_host" {
  description = "Wallarm API address"
  type    = string
  default = "https://api.wallarm.com"
}

variable "api_token" {
  description = "Wallarm token to authorize in API"
  type    = string
}

variable "node_names" {
  description = "Create Node names"
  type        = list(string)
  default     = ["prod", "stage", "dev"]
}
