variable "client_id" {
  type    = number
  default = null
}

variable "is_importing" {
  type = bool
}

data "wallarm_applications" "all" {
  client_id = var.client_id
  count     = var.is_importing ? 1 : 0
}

locals {
  import_blocks = var.is_importing ? join("\n", [
    for app in data.wallarm_applications.all[0].applications :
    "import {\n  to = wallarm_application.app_${app.app_id}\n  id = \"${app.client_id}/${app.app_id}\"\n}"
    if app.app_id != -1
  ]) : null
}

resource "local_file" "imports" {
  depends_on      = [data.wallarm_applications.all]
  filename        = "${path.root}/wallarm_application_imports.tf"
  content         = local.import_blocks
  file_permission = "0644"
  count           = var.is_importing ? 1 : 0
}

output "all_applications" {
  value = length(data.wallarm_applications.all) > 0 ? [
    for app in data.wallarm_applications.all[0].applications : app
    if app.app_id != -1
  ] : null
}

output "application_count" {
  value = length(data.wallarm_applications.all) > 0 ? length([
    for app in data.wallarm_applications.all[0].applications : app
    if app.app_id != -1
  ]) : null
}
