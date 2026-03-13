# import_applications

Fetches all existing application pools from the Wallarm API and generates Terraform import blocks for state adoption.

## Overview

This module calls `data.wallarm_applications` to retrieve all application pools managed in the Wallarm API, filters out the default application (`app_id = -1`), and generates `import {}` blocks that enable `terraform plan -generate-config-out` to produce the matching resource configurations.

## How It Works

When `is_importing = true`:

1. `data.wallarm_applications.all` fetches all applications from the API
2. The default/global application (`app_id = -1`) is excluded
3. For each application, an import block is generated:
   ```hcl
   import {
     to = wallarm_application.app_42
     id = "8649/42"
   }
   ```
4. Import blocks are written to `wallarm_application_imports.tf` at the project root
5. Run `terraform plan -generate-config-out=imported_applications.tf` to generate resource configs
6. Run `terraform apply` to reconcile state

When `is_importing = false`, the data source and local file are skipped entirely (via `count = 0`).

## Usage

Called by the `wallarm_import` parent module:

```hcl
module "import_applications" {
  source       = "./modules/import_applications"
  is_importing = var.is_importing
  client_id    = var.client_id
}
```

### Step-by-Step Import

```bash
# 1. Fetch all applications from API and generate import blocks
terraform apply -auto-approve -var='is_importing=true'

# 2. Generate Terraform resource configurations from import blocks
terraform plan -var='is_importing=true' -generate-config-out=imported_applications.tf

# 3. Apply the imported configuration
terraform apply --auto-approve
```

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | `null` | no | Wallarm client ID |
| `is_importing` | `bool` | — | yes | Must be `true` to fetch applications from API |

## Outputs

| Name | Description |
|------|-------------|
| `all_applications` | All applications from the API, excluding the default app (null when `is_importing=false`) |
| `application_count` | Number of applications, excluding the default app (null when `is_importing=false`) |

## Generated File

`wallarm_application_imports.tf` at the project root — contains one `import {}` block per application with `to = wallarm_application.app_<app_id>` and `id = "<client_id>/<app_id>"`.
