# import_rules

Fetches all existing WAF rules from the Wallarm API and generates Terraform import blocks for state adoption.

## Overview

This module calls `data.wallarm_rules` to retrieve all rules managed in the Wallarm API, then generates `import {}` blocks that enable `terraform plan -generate-config-out` to produce the matching resource configurations.

## How It Works

When `is_importing = true`:

1. `data.wallarm_rules.all` fetches all rules from the API
2. For each rule, an import block is generated:
   ```hcl
   import {
     to = wallarm_rule_disable_stamp.rule_12345
     id = "8649/12345"
   }
   ```
3. Import blocks are written to `wallarm_rule_imports.tf` at the project root
4. Run `terraform plan -generate-config-out=imported_rules.tf` to generate resource configs
5. Run `terraform apply` to reconcile state

When `is_importing = false`, the data source and local file are skipped entirely (via `count = 0`).

## Usage

Called by the `wallarm_import` parent module:

```hcl
module "import_rules" {
  source       = "./modules/import_rules"
  is_importing = var.is_importing
  client_id    = var.client_id
}
```

### Step-by-Step Import

```bash
# 1. Fetch all rules from API and generate import blocks
terraform apply -auto-approve -var='is_importing=true'

# 2. Generate Terraform resource configurations from import blocks
terraform plan -var='is_importing=true' -generate-config-out=imported_rules.tf

# 3. Apply the imported configuration
terraform apply --auto-approve
```

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | — | yes | Wallarm client ID |
| `is_importing` | `bool` | — | yes | Must be `true` to fetch rules from API |

## Outputs

| Name | Description |
|------|-------------|
| `all_rules` | All rules fetched from the Wallarm API (null when `is_importing=false`) |

## Generated File

`wallarm_rule_imports.tf` at the project root — contains one `import {}` block per rule with `to` (resource address) and `id` (client_id/rule_id).
