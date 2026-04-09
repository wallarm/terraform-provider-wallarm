---
layout: "wallarm"
page_title: "Wallarm: wallarm_rules"
subcategory: "Rules"
description: |-
  Reads existing Wallarm rules from the API.
---

# wallarm_rules

Reads all Wallarm rules for the specified client from the API. Returns rule metadata including IDs, types, and action conditions. Used by the import module to discover existing rules and generate Terraform import blocks.

Results are served from the hint cache when available, avoiding redundant API calls.

## Example Usage

```hcl
data "wallarm_rules" "all" {}

output "rule_count" {
  value = length(data.wallarm_rules.all.rules)
}
```

### Filter by Rule Type

```hcl
data "wallarm_rules" "all" {
  type = ["wallarm_mode", "vpatch", "masking"]
}
```

## Argument Reference

* `client_id` - (Optional) ID of the client to query. Defaults to the provider's default client ID.
* `type` - (Optional) List of rule type names to filter by. When omitted, returns all rule types. Valid values include: `wallarm_mode`, `vpatch`, `masking`, `ignore_regex`, `regex`, `binary_data`, `disable_attack_type`, `disable_stamp`, `set_response_header`, `parser_state`, `uploads`, `brute`, `bruteforce_counter`, `dirbust_counter`, `bola`, `bola_counter`, `rate_limit`, `rate_limit_enum`, `enum`, `forced_browsing`, `graphql_detection`, `file_upload_size_limit`, `overlimit_res_settings`, `credential_stuffing_regex`, `credential_stuffing_point`.

## Attributes Reference

* `rules` - List of rule objects for import workflows, each containing:
  * `rule_id` - (Int) Rule (hint) ID.
  * `action_id` - (Int) Action ID.
  * `client_id` - (Int) Client ID.
  * `type` - (String) Rule type (API type name).
  * `terraform_resource` - (String) Corresponding Terraform resource type name (e.g., `wallarm_rule_mode`).
  * `import_id` - (String) Pre-computed import ID for `terraform import`.

* `rules_export` - Full rule details with reverse-mapped scope fields. Used for config generation and export workflows. Each entry contains:

  **Identifiers:**
  * `rule_id`, `action_id`, `client_id` - Rule and action IDs.
  * `api_type` - API rule type name.
  * `terraform_resource` - Terraform resource type name.
  * `import_id` - Pre-computed import ID.

  **Reverse-mapped scope (from action conditions):**
  * `path` - URL path reconstructed from path/action_name/action_ext conditions (e.g., `/api/v1/users`).
  * `domain` - Domain from HOST header condition.
  * `instance`, `method`, `scheme`, `proto` - Other scope fields.
  * `conditions_hash` - SHA256 hash of action conditions (Ruby-compatible).
  * `action_dir_name` - Computed directory name for this action scope.

  **Serialized fields (JSON strings):**
  * `query_json` - Query parameter conditions as JSON array.
  * `headers_json` - Custom header conditions as JSON array.
  * `action_json` - Raw action conditions as JSON.
  * `point_json` - Detection point as JSON.

  **Rule-specific fields:** `comment`, `attack_type`, `stamp`, `mode`, `regex`, `parser`, `state`, `file_type`, `size`, `size_unit`, `delay`, `burst`, `rate`, `time_unit`, `overlimit_time`, `header_name`, `header_values_json`, `variativity_disabled`, and GraphQL/credential stuffing/threshold fields as applicable.


**Important:** Rules created using Terraform cannot be modified by other types of rules, such as those employing middleware, variative_values, or variative_by_regex.

This restriction exists because Terraform is built to maintain stable configurations and is not designed for external modifications.

Similarly, during the import process, imported rules will be updated automatically to align with this approach, which is necessary to preserve the stability of the Terraform state.
