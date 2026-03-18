---
layout: "wallarm"
page_title: "Wallarm: wallarm_rules"
subcategory: "Data Source"
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

* `rules` - List of rule objects, each containing:
  * `rule_id` - (Int) Rule (hint) ID.
  * `action_id` - (Int) Action ID.
  * `client_id` - (Int) Client ID.
  * `type` - (String) Rule type (API type name).
  * `resource_type` - (String) Corresponding Terraform resource type name (e.g., `wallarm_rule_mode`).
  * `import_id` - (String) Pre-computed import ID for use in `terraform import` commands. Format varies by type: 3-part (`{clientID}/{actionID}/{ruleID}`) or 4-part (`{clientID}/{actionID}/{ruleID}/{mode}`).


**Important:** Rules created using Terraform cannot be modified by other types of rules, such as those employing middleware, variative_values, or variative_by_regex.

This restriction exists because Terraform is built to maintain stable configurations and is not designed for external modifications.

Similarly, during the import process, imported rules will be updated automatically to align with this approach, which is necessary to preserve the stability of the Terraform state.
