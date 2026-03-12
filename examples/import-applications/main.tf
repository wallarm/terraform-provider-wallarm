# Import Applications Helper
#
# Generates import blocks for all existing Wallarm applications,
# allowing you to bring them under Terraform management.
#
# Usage:
#
# Option A (recommended):
#   1. terraform init && terraform apply
#      -> Only creates the data source in state (reads apps from API)
#
#   2. terraform output -raw import_blocks > wallarm_application_imports.tf
#      -> Writes import blocks to a file (nothing stored in state)
#
#   3. terraform plan -generate-config-out=imported_apps.tf
#      -> Terraform generates resource configs matching each import block
#
#   4. Review imported_apps.tf, then: terraform apply
#      -> All applications are imported into Terraform state
#
#   5. Move the generated resource blocks to your main config,
#      remove this helper and wallarm_application_imports.tf
#
# Option B:
#   Alternatively, you can configure the module to save import blocks to a file using
#   the local_file resource. This configuration will also add a file resource to the state.
#
# resource "local_file" "import_blocks" {
#   filename = "${path.module}/wallarm_application_imports.tf"
#   content = join("\n", concat(
#     [
#       "# Auto-generated import blocks for Wallarm applications.",
#       "# Run: terraform plan -generate-config-out=imported_apps.tf",
#       ""
#     ],
#     [
#       for app in data.wallarm_applications.all.applications :
#       <<-EOT
#       import {
#         to = wallarm_application.app_${app.app_id}
#         id = "${app.client_id}/${app.app_id}"
#       }
#       EOT
#       if app.app_id != -1
#     ]
#   ))
# }

terraform {
  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
    }
  }
}

provider "wallarm" {
  api_token = var.api_token
  api_host  = var.api_host
}

variable "api_token" {
  type      = string
  sensitive = true
}

variable "api_host" {
  type    = string
  default = "https://us1.api.wallarm.com"
}

variable "client_id" {
  type    = number
  default = null
}

data "wallarm_applications" "all" {
  client_id = var.client_id
}

output "import_blocks" {
  value = join("\n", [
    for app in data.wallarm_applications.all.applications :
    "import {\n  to = wallarm_application.app_${app.app_id}\n  id = \"${app.client_id}/${app.app_id}\"\n}"
    if app.app_id != -1
  ])
  description = "Import blocks for all existing applications"
}

output "application_count" {
  value = length([
    for app in data.wallarm_applications.all.applications : app
    if app.app_id != -1
  ])
  description = "Number of applications to import (excluding default app)"
}
