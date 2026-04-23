# Terraform Provider for Wallarm

[![Unit Tests](https://github.com/wallarm/terraform-provider-wallarm/actions/workflows/unit-tests.yml/badge.svg?branch=master)](https://github.com/wallarm/terraform-provider-wallarm/actions/workflows/unit-tests.yml)
[![Acceptance Tests](https://github.com/wallarm/terraform-provider-wallarm/actions/workflows/acceptance-tests.yml/badge.svg?branch=master)](https://github.com/wallarm/terraform-provider-wallarm/actions/workflows/acceptance-tests.yml)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL--2.0-blue.svg)](LICENSE)

The Wallarm Terraform provider manages resources on the [Wallarm](https://www.wallarm.com/) API security platform — rules, IP lists, integrations, applications, nodes, tenants, and triggers.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.5
- [Go](https://golang.org/doc/install) >= 1.24 (only to build from source)

## Installation

Install the provider from the [Terraform Registry](https://registry.terraform.io/providers/wallarm/wallarm/latest):

```hcl
terraform {
  required_version = ">= 1.5"

  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
    }
  }
}
```

Then run `terraform init`.

## Authentication

The provider authenticates via an API token. Set it as an environment variable (recommended):

```sh
export WALLARM_API_TOKEN="your-api-token"
```

Or configure it in the provider block:

```hcl
provider "wallarm" {
  api_token = var.api_token
}

variable "api_token" {
  type      = string
  sensitive = true
}
```

The API host defaults to `https://api.wallarm.com`. Override with `WALLARM_API_HOST` or the `api_host` attribute for other regions (e.g. `https://us1.api.wallarm.com`).

For multi-tenant setups, set `client_id` on the provider or individual resources to target specific tenant accounts.

> **Note:** Never commit API tokens to version control. Use environment variables, Terraform variables with `.tfvars` files (added to `.gitignore`), or a secrets manager.

## Usage Example

Enable blocking mode for a specific host:

```hcl
resource "wallarm_rule_mode" "block_example" {
  mode = "block"

  action {
    type  = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
}
```

Deny traffic from specific IPs:

```hcl
resource "wallarm_denylist" "bad_ips" {
  ip_range    = ["1.2.3.4", "5.6.7.0/24"]
  reason      = "Blocked by Terraform"
  time_format = "RFC3339"
  time        = "2027-01-01T00:00:00+00:00"
}
```

## Resources

### Mitigation Controls (8 resources)

Session-based rules for real-time threat mitigation. In the Wallarm Console UI these appear under a dedicated Mitigation Controls section.

| Resource | Description |
|----------|-------------|
| `wallarm_rule_mode` | Real-time blocking mode |
| `wallarm_rule_graphql_detection` | GraphQL API protection |
| `wallarm_rule_enum` | Enumeration attack protection |
| `wallarm_rule_bola` | BOLA/IDOR protection |
| `wallarm_rule_forced_browsing` | Forced browsing protection |
| `wallarm_rule_brute` | Brute force protection |
| `wallarm_rule_rate_limit_enum` | DoS protection (rate limiting) |
| `wallarm_rule_file_upload_size_limit` | File upload restriction policy |

### Rules (21 resources)

Request-level rules for detection tuning, virtual patching, and data handling. These appear under the Rules section in the Wallarm Console UI.

> **Note:** `wallarm_rule_mode`, `wallarm_rule_graphql_detection`, and `wallarm_rule_file_upload_size_limit` appear in both tables because they are shown in both the Rules and Mitigation Controls sections of the Wallarm Console UI. They are a single Terraform resource — listed in both places here to match the UI experience.

| Resource | Description |
|----------|-------------|
| `wallarm_rule_mode` | Filtration mode (monitor/block) |
| `wallarm_rule_vpatch` | Virtual patches |
| `wallarm_rule_regex` | Custom regex detection |
| `wallarm_rule_ignore_regex` | Regex-based false positive suppression |
| `wallarm_rule_masking` | Sensitive data masking |
| `wallarm_rule_binary_data` | Binary data markers |
| `wallarm_rule_parser_state` | Parser control (enable/disable) |
| `wallarm_rule_set_response_header` | Response header injection |
| `wallarm_rule_disable_attack_type` | Disable specific attack type detection |
| `wallarm_rule_disable_stamp` | Disable specific attack signatures |
| `wallarm_rule_uploads` | File upload handling |
| `wallarm_rule_graphql_detection` | GraphQL detection settings |
| `wallarm_rule_file_upload_size_limit` | File upload size limits |
| `wallarm_rule_overlimit_res_settings` | Overlimit request settings |
| `wallarm_rule_rate_limit` | Rate limiting |
| `wallarm_rule_bruteforce_counter` | Brute force counter |
| `wallarm_rule_dirbust_counter` | Directory traversal counter |
| `wallarm_rule_bola_counter` | BOLA counter |
| `wallarm_rule_credential_stuffing_regex` | Credential stuffing regex |
| `wallarm_rule_credential_stuffing_point` | Credential stuffing detection points |
| `wallarm_rule_api_abuse_mode` | Toggle API Abuse Prevention per request scope |

### IP Lists (3 resources)

| Resource | Description |
|----------|-------------|
| `wallarm_denylist` | Block IPs, countries, datacenters, or proxy types |
| `wallarm_allowlist` | Allow specific traffic sources |
| `wallarm_graylist` | Graylist for behavioral analysis |

### Integrations (11 resources)

| Resource | Description |
|----------|-------------|
| `wallarm_integration_email` | Email notifications |
| `wallarm_integration_slack` | Slack notifications |
| `wallarm_integration_teams` | Microsoft Teams notifications |
| `wallarm_integration_telegram` | Telegram notifications |
| `wallarm_integration_pagerduty` | PagerDuty alerts |
| `wallarm_integration_opsgenie` | OpsGenie alerts |
| `wallarm_integration_splunk` | Splunk log forwarding |
| `wallarm_integration_sumologic` | Sumo Logic log forwarding |
| `wallarm_integration_data_dog` | Datadog log forwarding |
| `wallarm_integration_insightconnect` | InsightConnect integration |
| `wallarm_integration_webhook` | Custom webhook notifications |

### Infrastructure & Tooling (12 resources)

| Resource | Description |
|----------|-------------|
| `wallarm_application` | Application (pool) management |
| `wallarm_node` | Filtering node management |
| `wallarm_tenant` | Tenant (multi-tenancy) management |
| `wallarm_user` | User management |
| `wallarm_trigger` | Trigger configuration |
| `wallarm_global_mode` | Global filtration mode |
| `wallarm_rules_settings` | Rules engine settings |
| `wallarm_api_spec` | API specification management |
| `wallarm_api_spec_policy` | API Specification Enforcement policy attached to an uploaded spec |
| `wallarm_action` | Rule action scope tracking |
| `wallarm_rule_generator` | Generate HCL config files from hits or existing API rules |
| `wallarm_hits_index` | Track fetched request IDs for the [hits-to-rules workflow](docs/guides/hits_to_rules.md) |

### Data Sources (7 data sources)

| Data Source | Description |
|-------------|-------------|
| `wallarm_node` | Look up filtering nodes |
| `wallarm_applications` | List applications (supports bulk import) |
| `wallarm_actions` | Discover rule action scopes |
| `wallarm_rules` | Read all rules (hints) |
| `wallarm_hits` | Fetch detected hits for FP analysis |
| `wallarm_ip_lists` | Read IP list entries |
| `wallarm_security_issues` | Query security issues |

## Import

Most resources support `terraform import`. Bulk import is supported for:
- **IP lists** — via the `wallarm_ip_lists` data source. See the [IP List Import Guide](docs/guides/ip_list_import.md).
- **Rules** — via the `wallarm_rules` data source. See the [Rules Import Guide](docs/guides/rules_import.md).
- **Applications** — via the `wallarm_applications` data source to discover existing apps and generate import blocks.

## Guides

| Guide | Description |
|-------|-------------|
| [Action Scopes](docs/guides/action.md) | How action conditions define rule scope (Host, path, method, etc.) |
| [Detection Points](docs/guides/point.md) | Point syntax for specifying request parts in rules |
| [Mitigation Controls](docs/guides/mitigation_controls.md) | Configuring brute force, BOLA, enumeration, and other session-based protections |
| [Hits to Rules](docs/guides/hits_to_rules.md) | Creating false positive suppression rules from hit data |
| [Rules Import](docs/guides/rules_import.md) | Bulk-importing existing rules from the Wallarm API |
| [IP List Import](docs/guides/ip_list_import.md) | Bulk-importing IP list entries |

## Documentation

Full resource and data source documentation is in the [`docs/`](docs/) directory and on the [Terraform Registry](https://registry.terraform.io/providers/wallarm/wallarm/latest/docs).

## Building from Source

```sh
git clone https://github.com/wallarm/terraform-provider-wallarm.git
cd terraform-provider-wallarm
make build    # Build binary locally
make install  # Install to $GOPATH/bin
```

## Development

```sh
make test     # Unit tests
make testacc  # Acceptance tests (requires WALLARM_API_TOKEN, WALLARM_API_HOST, TF_ACC=1)
make lint     # Run golangci-lint
make fmt      # Format code
```

Run a single acceptance test:

```sh
go test -v -run TestAccRuleWmodeCreate_Basic ./wallarm/provider/ -timeout=120m
```

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.

## License

[Mozilla Public License 2.0](LICENSE)
