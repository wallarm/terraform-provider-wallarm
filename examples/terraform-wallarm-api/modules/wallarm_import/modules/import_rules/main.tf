variable "client_id" {
  type = number
}
variable "is_importing" {
  type        = bool
}
variable "rule_types" {
  type        = list(string)
  default     = []
  description = "Filter by API rule type(s). Empty list means all types."
}
data "wallarm_rules" "all" {
  client_id = var.client_id
  type      = var.rule_types
  count     = var.is_importing ? 1 : 0
}

locals {
  # Build import blocks for all fetched/filtered rules
  import_blocks = var.is_importing ? join("\n", [
    for rule in data.wallarm_rules.all[0].rules :
    "import {\n  to = ${rule.terraform_resource}.rule_${rule.rule_id}\n  id = \"${rule.import_id}\"\n}"
  ]) : null
}

resource "local_file" "imports" {
  depends_on      = [data.wallarm_rules.all]
  filename        = "${path.root}/wallarm_rule_imports.tf"
  content         = local.import_blocks
  file_permission = "0644"
  count           = var.is_importing ? 1 : 0
}

output "all_rules" {
  value = length(data.wallarm_rules.all) > 0 ? data.wallarm_rules.all[0].rules : null
}
