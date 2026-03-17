# wallarm_import

Parent module for importing existing Wallarm resources into Terraform state. Provides a framework for adopting API-managed resources into infrastructure-as-code.

## Overview

This module orchestrates child modules that fetch existing resources from the Wallarm API and generate Terraform import blocks. The import blocks are written to `.tf` files at the project root, enabling `terraform plan -generate-config-out` to produce the matching resource configurations.

Child modules:
- **import_rules** — imports existing WAF rules
- **import_applications** — imports existing application pools
- **import_ip_lists** — imports existing IP list entries (allowlist, denylist, graylist)

## Usage

```hcl
module "wallarm_import" {
  source = "./modules/wallarm_import"

  client_id    = 12345
  is_importing = true
}
```

### Import Workflow

```bash
# 1. Fetch all resources from API and generate import blocks
terraform apply -auto-approve -var='is_importing=true'

# 2. Generate Terraform resource configurations from import blocks
terraform plan -var='is_importing=true' -generate-config-out=imported.tf

# 3. Apply the imported configuration
terraform apply --auto-approve
```

After importing, set `is_importing = false` to disable the data source API calls.

### IP List Import Details

IP list entries are grouped for import based on their type:

| Rule type | Import strategy | Import ID format |
|-----------|----------------|------------------|
| `location` (country) | 1 API group = 1 resource | `{clientID}/{groupID}` |
| `datacenter` | 1 API group = 1 resource | `{clientID}/{groupID}` |
| `proxy_type` | 1 API group = 1 resource | `{clientID}/{groupID}` |
| `subnet` (IPs) | Merged by `expired_at` | `{clientID}/subnet/{expiredAt}` |

Subnets are merged by expiration time — all IPs with the same `expired_at` become one Terraform resource with `ip_range = [...]`. This is logically correct since IPs created together share the same expiration.

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `client_id` | `number` | — | yes | Wallarm client ID |
| `is_importing` | `bool` | `false` | no | Must be `true` to activate import functionality |

## Outputs

| Name | Description |
|------|-------------|
| `all_rules` | All imported rules from the Wallarm API (null when `is_importing=false`) |
| `all_applications` | All imported applications from the Wallarm API (null when `is_importing=false`) |
| `application_count` | Number of imported applications (null when `is_importing=false`) |
| `ip_list_entries` | IP list entries by list type: `{ denylist: [...], allowlist: [...], graylist: [...] }` |
| `ip_list_counts` | Number of entries per list type |

## Submodules

| Name | Description |
|------|-------------|
| [import_rules](modules/import_rules/) | Import existing WAF rules from the API |
| [import_applications](modules/import_applications/) | Import existing application pools from the API |
| [import_ip_lists](modules/import_ip_lists/) | Import existing IP list entries (allowlist, denylist, graylist) from the API |

## Generated Files

When `is_importing = true`, the following files are created at the project root:

| File | Contents |
|------|----------|
| `wallarm_rule_imports.tf` | Import blocks for all existing WAF rules |
| `wallarm_application_imports.tf` | Import blocks for all existing application pools |
| `wallarm_denylist_imports.tf` | Import blocks for denylist IP entries |
| `wallarm_allowlist_imports.tf` | Import blocks for allowlist IP entries |
| `wallarm_graylist_imports.tf` | Import blocks for graylist IP entries |
