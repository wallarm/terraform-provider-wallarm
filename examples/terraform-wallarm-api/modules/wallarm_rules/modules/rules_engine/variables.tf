variable "client_id" {
  type        = number
  description = "Wallarm client ID"
}

variable "config_dirs" {
  type        = list(string)
  description = "Directories to discover YAML config files from (recursive **/*.yaml)"
}

variable "generated_rules" {
  type        = any
  default     = []
  description = <<-EOT
    Rules from generators (hits). Each object has the same fields as a YAML
    config plus a _config_dir field specifying where the YAML should be written.
    On first apply: data is used directly + YAML is created.
    On subsequent applies: YAML file is read (user edits preserved).
  EOT
}
