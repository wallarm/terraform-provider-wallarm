---
layout: "wallarm"
page_title: "Importing Wallarm Rules"
description: |-
  How to import existing Wallarm rules into Terraform.
---

# Importing Wallarm Rules

Wallarm rules created via the Console UI or API can be imported into Terraform to bring them under infrastructure-as-code management.

## Import ID

Each rule resource uses a composite import ID. Most resources use a 3-part format:

```
{client_id}/{action_id}/{rule_id}
```

Two resources require an additional suffix:

| Resource | Format | Suffix |
|----------|--------|--------|
| `wallarm_rule_mode` | `{client_id}/{action_id}/{rule_id}/{mode}` | Mode value: `monitoring`, `block`, `off`, or `default` |
| `wallarm_rule_regex` | `{client_id}/{action_id}/{rule_id}/{rule_type}` | `regex` or `experimental_regex` |

Refer to the individual resource documentation for the exact import command and ID breakdown.

## Import Options

Terraform supports two ways to import:

- **CLI command** — `terraform import <resource_address> <import_id>` — imports a single resource into state
- **Import blocks** (Terraform v1.5+) — declarative `import {}` blocks in HCL, processed during `terraform plan`/`apply`. Supports the `-generate-config-out` flag to auto-generate resource configuration

## Configuration

The full configuration below supports all import workflows: discovery, import block generation, sync status checking, and HCL generation for complex rules. Copy it as a starting point and adjust as needed.

```hcl
# Provider configuration omitted — see the provider documentation for setup.

variable "import_rules" {
  type        = bool
  default     = false
  description = "Must be true to activate rules import functionality."
}

variable "rule_types" {
  type        = list(string)
  default     = []
  description = "Include only these API rule type(s). Empty list means all types."
}

variable "exclude_rule_types" {
  type        = list(string)
  default     = []
  description = "Exclude these API rule type(s) from import blocks and sync status."
}

variable "sync_status" {
  type        = bool
  default     = false
  description = "Must be true to activate rules sync status (API <-> TF state)."
}

variable "generate_configs" {
  type        = bool
  default     = false
  description = "Generate HCL configs via wallarm_rule_generator. Use as fallback when -generate-config-out fails on complex point values."
}

# ─── Fetch rules from API ──────────────────────────────────────────────────

data "wallarm_rules" "all" {
  type  = var.rule_types
  count = var.import_rules || var.sync_status ? 1 : 0
}

# ─── Generate import blocks ────────────────────────────────────────────────

locals {
  import_blocks = var.import_rules ? join("\n", [
    for rule in data.wallarm_rules.all[0].rules :
    "import {\n  to = ${rule.terraform_resource}.rule_${rule.rule_id}\n  id = \"${rule.import_id}\"\n}"
    if !contains(var.exclude_rule_types, rule.type)
  ]) : null
}

resource "local_file" "imports" {
  filename        = "${path.root}/import_rule_blocks.tf"
  content         = local.import_blocks
  file_permission = "0644"
  count           = var.import_rules ? 1 : 0
}

output "import_blocks" {
  value       = local.import_blocks
  description = "Import blocks for existing rules"
}

# ─── Sync status: compare API rules vs Terraform state ────────────────────
# Requires: jq (https://jqlang.github.io/jq/)

data "external" "sync_check" {
  count = var.sync_status ? 1 : 0
  program = ["bash", "-c", <<-EOF
    ids=$(terraform show -json 2>/dev/null \
      | jq -r '.values.root_module.resources[]
        | select(.type | startswith("wallarm_rule_"))
        | .values.rule_id' \
      | sort -un \
      | paste -sd, - || echo "")
    printf '{"ids":"%s"}' "$ids"
  EOF
  ]
}

locals {
  state_rule_ids = var.sync_status ? toset(compact(split(",",
    try(data.external.sync_check[0].result.ids, "")
  ))) : toset([])

  api_rule_ids = var.sync_status ? toset([
    for r in data.wallarm_rules.all[0].rules : tostring(r.rule_id)
    if !contains(var.exclude_rule_types, r.type)
  ]) : toset([])

  unmanaged_ids = setsubtract(local.api_rule_ids, local.state_rule_ids)
}

output "sync_status" {
  value = var.sync_status ? {
    total_in_api = length(local.api_rule_ids)
    in_state     = length(local.state_rule_ids)
    unmanaged    = length(local.unmanaged_ids)
  } : null
}

# ─── Generate HCL configs via rule_generator (fallback) ───────────────────
# For disable_stamp rules with complex point values (e.g. XML namespace URIs)
# where -generate-config-out fails to generate the point field.

resource "wallarm_rule_generator" "from_api" {
  count           = var.generate_configs ? 1 : 0
  source          = "api"
  output_dir      = "./"
  output_filename = "import_rule_stamp_configs.tf"
  rule_types      = ["disable_stamp"]
  split           = false
}
```

## Workflow: Import All Rules

This is the standard workflow for importing all existing rules into Terraform.

**Step 1.** Fetch rules from the API and generate import blocks:

```bash
terraform apply -var='import_rules=true'
```

This creates `import_rule_blocks.tf` containing an `import {}` block for each rule.

**Step 2.** Generate resource configurations from the import blocks:

```bash
terraform plan -generate-config-out=import_rule_configs.tf
```

Terraform reads the import blocks and writes a matching resource block for each one into `import_rule_configs.tf`.

**Step 3.** Fix generated configs. Terraform generates `null` for optional fields and incorrect defaults for some fields. Run this to clean them up:

```bash
sed -E \
  -e 's/(variativity_disabled[[:space:]]*)=[[:space:]]*false/\1= true/' \
  -e 's/(comment[[:space:]]*)=[[:space:]]*null/\1= "Managed by Terraform"/' \
  -e '/=[[:space:]]*null/d' \
  import_rule_configs.tf > import_rule_configs.tf.tmp && \
  mv import_rule_configs.tf.tmp import_rule_configs.tf
```

**Step 4.** Review the generated configs, then import into state:

```bash
terraform apply
```

Re-importing existing resources is safe — Terraform generates configs only for resources not already in state.

## Workflow: Import with Generator Fallback

For `disable_stamp` rules with complex point values (e.g. XML namespace URIs, special characters), Terraform's `-generate-config-out` may fail to generate the `point` field correctly. Use the `wallarm_rule_generator` resource as a fallback for stamps, and native import for everything else.

**Step 1.** Import all rules except stamps:

```bash
terraform apply -var='import_rules=true' -var='exclude_rule_types=["disable_stamp"]'
terraform plan -generate-config-out=import_rule_configs.tf
```

Fix generated configs with the `sed` command from the previous section, then apply:

```bash
terraform apply
```

**Step 2.** Generate stamp configs via `wallarm_rule_generator` and import blocks for stamps only:

```bash
terraform apply -var='generate_configs=true' -var='import_rules=true' -var='rule_types=["disable_stamp"]'
```

This writes `import_rule_stamp_configs.tf` (HCL resource blocks via the generator, which handles complex point values) and `import_rule_blocks.tf` (import blocks for stamp rules only).

**Step 3.** Import stamp rules:

```bash
terraform apply
```

## Filtering by Rule Type

To import only specific rule types, use the `rule_types` variable:

```bash
terraform apply -var='import_rules=true' -var='rule_types=["wallarm_mode"]'
```

To exclude specific types from import and sync status, use `exclude_rule_types`:

```bash
terraform apply -var='import_rules=true' -var='exclude_rule_types=["disable_stamp", "disable_attack_type"]'
```

See the [`wallarm_rules` data source documentation](../data-sources/rules) for the full list of valid type values.

## Sync Status

Check how many API rules are not yet managed by Terraform — a read-only operation that makes no changes.

**Prerequisites:** [jq](https://jqlang.github.io/jq/) must be installed. The sync status uses `terraform show -json` and `jq` to extract rule IDs from the current state.

```bash
terraform plan -refresh=false -var='sync_status=true'
```

This compares rule IDs in the API against rule IDs in Terraform state and outputs the difference:

```
sync_status = {
  total_in_api = 142
  in_state     = 98
  unmanaged    = 44
}
```

Use this to track import progress or detect rules created outside of Terraform. You can filter with `exclude_rule_types` to ignore rule types you don't plan to manage:

```bash
terraform plan -refresh=false -var='sync_status=true' -var='exclude_rule_types=["disable_stamp"]'
```

## Important Notes

- **Plan after import:** Always run `terraform plan` after importing to verify the configuration matches actual state. Adjust your HCL until the plan shows no changes.
- **State stability:** Imported rules are automatically updated to prevent external modification by other rule types (middleware, variative_values, variative_by_regex). This preserves Terraform state consistency.
- **Multi-tenant:** Specify `client_id` in the data source to discover rules for a specific tenant.
- **Re-import is safe:** Terraform generates configs only for resources not already in state, so running the import workflow again won't duplicate resources.
