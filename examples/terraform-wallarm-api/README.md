# terraform-wallarm-api

Terraform module for managing Wallarm WAF rules and importing existing resources from the Wallarm API.

## Features

- **Rule creation from attack hits** — Fetch hits by request ID (once), aggregate by detection point, persist in Terraform state, and auto-create `disable_stamp` / `disable_attack_type` rules (false-positive suppression). Supports **attack mode** to expand from a single request to all related hits in the same attack campaign.
- **Custom rule creation** — Define rules in `terraform.tfvars` with a `path` field that auto-expands into Wallarm action conditions. Supports **25 resource types** (masking, vpatch, regex, rate_limit, brute, bola, mode, and more)
- **Import existing resources** — Import rules and applications already managed in the Wallarm API into Terraform state using `terraform plan -generate-config-out`
- **Editable config files** — Every rule generates a YAML or HCL config file that can be manually edited. Variable values always override config file values (variables-first pattern)
- **Wildcard paths** — `*` matches any single segment, `**` matches any depth

## Architecture

```
┌────────────────────────────────────────────────────────────────────────┐
│                          Root Module                                   │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  module "wallarm_rules"                                          │  │
│  │                                                                  │  │
│  │  hits_fetcher → terraform_data → fp_rules (false-positive rules) │  │
│  │                                  custom_rules (25 resource types)│  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  module "wallarm_import"                                         │  │
│  │                                                                  │  │
│  │  import_rules         (existing WAF rules from API)              │  │
│  │  import_applications  (existing application pools from API)      │  │
│  │  import_ip_lists      (existing IP list entries from API)        │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────┘
```

## Usage

### Basic — Custom Rules Only

```hcl
module "wallarm" {
  source = "wallarm/api/wallarm"

  api_host  = "https://us1.api.wallarm.com"
  api_token = var.wallarm_api_token
  client_id = 12345

  requests = {}

  custom_rules = [
    {
      name          = "block_admin"
      resource_type = "wallarm_rule_mode"
      mode          = "block"
      path          = "/admin"
      domain        = "example.com"
    },
    {
      name          = "mask_passwords"
      resource_type = "wallarm_rule_masking"
      point         = [["post"], ["json_doc"], ["hash", "password"]]
      path          = "/api/auth/login"
      domain        = "example.com"
    },
  ]
}
```

### False-Positive Rules from Hits

```hcl
module "wallarm" {
  source = "wallarm/api/wallarm"

  api_host  = "https://us1.api.wallarm.com"
  api_token = var.wallarm_api_token
  client_id = 12345

  hits_mode = "attack"  # Expand to all related hits by attack_id

  requests = {
    "abc123def456" = ["disable_stamp", "disable_attack_type"]
    "789ghi012jkl" = ["disable_stamp"]
  }

  custom_rules = []
}
```

Hits are fetched from the API automatically on first apply for each request_id. Subsequent applies use persisted data from Terraform state — no API calls needed.

**Hits mode:**
- `"request"` (default) — fetch hits for the specific request_id only
- `"attack"` — fetch hits for the request_id, then expand to all related hits sharing the same attack_id, filtered by allowed attack types and matching action (Host + path)

### Import Existing Resources

```bash
# 1. Fetch all rules, applications, and IP lists from API, generate import blocks
terraform apply -auto-approve -var='is_importing=true'

# 2. Generate Terraform resource configurations from import blocks
terraform plan -var='is_importing=true' -generate-config-out=imported.tf

# 3. Apply the imported configuration
terraform apply --auto-approve
```

The import module generates import blocks for:
- **Rules** — all existing WAF rules
- **Applications** — all application pools (except default app_id=-1)
- **IP Lists** — all allowlist/denylist/graylist entries. Grouped types (country/datacenter/proxy) import by group ID. Subnets are merged by expiration time into one resource per unique `expired_at`.

## Requirements

| Name      | Version   |
|-----------|-----------|
| terraform | >= 1.10.1 |
| wallarm   | 2.0.0     |

## Providers

| Name    | Version |
|---------|---------|
| wallarm | 2.0.0   |

## Variables

| Name | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `api_host` | `string` | — | yes | Wallarm API endpoint URL |
| `api_token` | `string` | — | yes | API authentication token |
| `client_id` | `number` | — | yes | Wallarm client ID |
| `requests` | `map(list(string))` | — | yes | Map of `request_id => [rule_types]` for hit-based rules |
| `hits_mode` | `string` | `"request"` | no | Fetch mode: `"request"` (direct hits) or `"attack"` (expand to related hits by attack_id) |
| `custom_rules` | `list(object)` | `[]` | no | Custom rules defined in variables (see [custom_rules module](modules/wallarm_rules/modules/custom_rules/)) |
| `is_importing` | `bool` | `false` | no | Activate import functionality |
| `subnet_import_mode` | `string` | `"grouped"` | no | `"grouped"` merges subnets by expiry (max 1000/resource); `"individual"` creates one resource per IP |
| `config_format` | `string` | `"yaml"` | no | Config file format: `"yaml"` or `"hcl"` |

## Outputs

| Name | Description |
|------|-------------|
| `rule_ids_by_request` | Rule IDs grouped by request_id (from false-positive rules) |
| `custom_rule_ids` | Map of custom rule names to their created rule IDs |
| `all_rule_ids` | Flat map of all created rule IDs |
| `total_rules` | Total number of imported rules (when `is_importing=true`) |
| `imported_rules` | All imported rules (when `is_importing=true`) |
| `imported_applications` | All imported applications (when `is_importing=true`) |
| `imported_application_count` | Number of imported applications (when `is_importing=true`) |

## Modules

| Name | Description |
|------|-------------|
| [wallarm_rules](modules/wallarm_rules/) | Rule creation pipeline: hits fetching, state persistence, false-positive rules, and custom rules |
| [wallarm_import](modules/wallarm_import/) | Import existing Wallarm resources (rules, applications) into Terraform state |

## File Structure

```
terraform-wallarm-api/
├── main.tf                           # Root orchestration
├── variables.tf                      # Root inputs
├── terraform.tf                      # Provider requirements
├── providers.tf                      # Provider configuration
├── terraform.tfvars                  # API credentials (api_host, api_token, client_id)
├── requests.auto.tfvars              # Hit-based rule request IDs
├── custom_rules.auto.tfvars          # Custom rule definitions (25 resource types)
├── Makefile                          # Convenience targets
├── rules_config/                     # Custom rule config files (runtime)
│   └── wallarm_rule_*.<yaml|tf>      # Custom rule configs
├── fp-rules-configs/                 # FP rule config files (runtime)
│   └── <request_id>/                 # Per-request subdirectory
│       ├── <hash>_disable_stamp.yaml
│       └── <hash>_disable_attack_type.yaml
│
└── modules/
    ├── wallarm_rules/                # Rule creation
    │   └── modules/
    │       ├── hits_fetcher/         # Fetch + aggregate hits, persist in state
    │       ├── fp_rules/             # False-positive rules from hits
    │       └── custom_rules/         # 25 resource types with path expansion
    │
    └── wallarm_import/               # Import existing resources
        └── modules/
            ├── import_rules/         # Import WAF rules
            └── import_applications/  # Import application pools
```

## Makefile Targets

```bash
make init     # terraform init
make apply    # terraform apply (auto-detects new requests)
make clean    # rm state, config files, .terraform
make state    # terraform state list
```

## State Migration

### Upgrading from the Old 4-Module Structure

If upgrading from the previous flat module layout (4 top-level modules), run:

```bash
# Move hits into wallarm_rules parent
terraform state mv 'module.hits' 'module.wallarm_rules.module.hits'

# Move wallarm_rules (old) into fp_rules child
terraform state mv 'module.wallarm_rules' 'module.wallarm_rules.module.fp_rules'

# Move custom_rules into wallarm_rules parent
terraform state mv 'module.custom_rules' 'module.wallarm_rules.module.custom_rules'

# Move import_rules into wallarm_import parent
terraform state mv 'module.import_rules' 'module.wallarm_import.module.import_rules'

# Reinitialize
terraform init -upgrade
terraform plan  # verify no destroy/recreate
```

## License

See [LICENSE](LICENSE) for details.
