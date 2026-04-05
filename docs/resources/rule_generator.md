---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_generator"
subcategory: "Common"
description: |-
  Generates HCL config files for Wallarm rules from cached rule data or existing API rules.
---

# wallarm_rule_generator

Generates HCL `.tf` files for Wallarm rules. Supports two source modes:

- **`rules`** (default) -- generates HCL from pre-built rules (e.g., cached rules from the hits-to-rules workflow)
- **`api`** -- fetches existing rules from the Wallarm API and generates HCL

Generated files persist on disk after the resource is removed from state. This is a **state-only delete** -- removing the resource does not delete the generated files.

## Example Usage

### From pre-built rules (recommended for hits-to-rules workflow)

```hcl
resource "wallarm_rule_generator" "from_rules" {
  source     = "rules"
  output_dir = "./generated_rules"
  split      = true
  moved_from = "this"
  rules_json = jsonencode(local._all_rules)
}
```

### From existing API rules

```hcl
resource "wallarm_rule_generator" "from_api" {
  source     = "api"
  output_dir = "./imported_rules"
  rule_types = ["disable_stamp", "disable_attack_type"]
}
```

### With moved blocks for migration

```hcl
resource "wallarm_rule_generator" "migrate" {
  source     = "rules"
  moved_from = "this"
  output_dir = "./migrated_rules"
  rules_json = jsonencode(local._all_rules)
}
```

## Argument Reference

* `client_id` - (Optional) Client ID for generated resource blocks. Defaults to the provider's client ID.
* `output_dir` - (Required, ForceNew) Directory to write generated `.tf` files.
* `output_filename` - (Optional) Filename when `split = false`. Defaults to `{prefix}_rules.tf`.
* `source` - (Optional) Source of rules: `rules` or `api`. Default: `rules`.
* `rules_json` - (Optional) JSON-encoded list of pre-built rules. Required when `source = "rules"`.
* `rule_types` - (Optional) Filter by rule type. Possible values: `disable_stamp`, `disable_attack_type`. Default: all types.
* `resource_prefix` - (Optional) Prefix for resource names. Default: `fp` for rules, `rule` for api.
* `split` - (Optional) One file per rule when true, all in one file when false. Default: `false`.
* `comment` - (Optional) Comment for generated resources. Default: `Managed by Terraform`.
* `moved_from` - (Optional) Resource name to generate `moved` blocks from (for migration from `for_each`-based resources).

## Attributes Reference

* `generated_files` - List of paths of generated `.tf` files.
* `rules_count` - Number of generated rules.
