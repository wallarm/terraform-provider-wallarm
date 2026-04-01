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

## Automated Import with the Rules Data Source

The [`wallarm_rules`](../data-sources/rules) data source discovers all existing rules and provides pre-computed `import_id` values for each one. This enables a workflow where you fetch rules, generate import blocks, and let Terraform generate the resource configuration.

### Example: Import All Rules

**1. Create a discovery configuration:**

```hcl
# import_discovery/main.tf

data "wallarm_rules" "all" {}

resource "local_file" "import_blocks" {
  filename = "${path.module}/imports.tf"
  content = join("\n", [
    for rule in data.wallarm_rules.all.rules :
    <<-EOT
    import {
      to = ${rule.resource_type}.rule_${rule.rule_id}
      id = "${rule.import_id}"
    }
    EOT
  ])
}
```

**2. Apply to generate the `imports.tf` file:**

```
$ terraform apply
```

**3. Copy `imports.tf` to your target configuration directory, then generate resource configs:**

```
$ terraform plan -generate-config-out=generated.tf
```

Terraform reads the import blocks and writes a matching resource block for each one into `generated.tf`. Review the output — adjust resource names or arguments as needed, then apply.

### Filter by Rule Type

To import only specific rule types, use the `type` filter:

```hcl
data "wallarm_rules" "vpatch_only" {
  type = ["vpatch"]
}
```

See the [`wallarm_rules` data source documentation](../data-sources/rules) for the full list of valid type values.

## Sync Status

Check how many API rules are not yet managed by Terraform — a read-only operation that makes no changes:

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

Use this to track import progress or detect rules created outside of Terraform. The `sync_status` variable is available in the `examples/import-rules` example module. You can also filter by rule type using `exclude_rule_types` to ignore types you don't plan to manage.

## Important Notes

- **Plan after import:** Always run `terraform plan` after importing to verify the configuration matches actual state. Adjust your HCL until the plan shows no changes.
- **State stability:** Imported rules are automatically updated to prevent external modification by other rule types (middleware, variative_values, variative_by_regex). This preserves Terraform state consistency.
- **Multi-tenant:** Specify `client_id` in the data source to discover rules for a specific tenant.
