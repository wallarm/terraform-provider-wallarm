---
layout: "wallarm"
page_title: "Wallarm: wallarm_rules_settings"
subcategory: "Common"
description: |-
  Provides the resource to manage rules settings of the company.
---

# wallarm_rules_settings

Provides the resource to manage rules settings of the company.

## Duplicates
Every client can have only one wallarm_rules_settings resource.
Unfortunately, terraform doesn't support singleton resources.
Therefore, you must ensure that there is no more than one resource per client in your configuration.
Otherwise, the terraform provider will merge them and send them as one resource. Priority for identical fields will be random.

Let's break down the merge rules using the following configuration:
```hcl
# default client_id = 1

resource "wallarm_rules_settings" "rules_settings1" {
  min_lom_format = 50
  max_lom_format = 70
}

resource "wallarm_rules_settings" "rules_settings2" {
  min_lom_format = 51
  max_lom_size = 10000000
}

resource "wallarm_rules_settings" "rules_settings3" {
  client_id = 1
  min_lom_format = 52
}

resource "wallarm_rules_settings" "rules_settings4" {
  client_id = 2
  min_lom_format = 53
}

resource "wallarm_rules_settings" "rules_settings5" {
  client_id = 2
  min_lom_format = 54
}

resource "wallarm_rules_settings" "rules_settings6" {
  client_id = 3
  min_lom_format = 55
}
```

The provider will merge them inside and the final configuration will be equivalent to:
```hcl
# based on rules_settings1, rules_settings2 and rules_settings3
resource "wallarm_rules_settings" "new_rules_settings1" {
  client_id = 1
  min_lom_format = 51 # random value from 50, 51, 52
  max_lom_format = 70
  max_lom_size = 10000000
}

# based on rules_settings3 and rules_settings4
resource "wallarm_rules_settings" "new_rules_settings2" {
  client_id = 2
  min_lom_format = 52 # random value from 52, 53
}

resource "wallarm_rules_settings" "rules_settings5" {
  client_id = 3
  min_lom_format = 55
}
```

## Example Usage

```hcl
# Configure rules settings

resource "wallarm_rules_settings" "rules_settings" {
  client_id = 123
  min_lom_format = 50
	max_lom_format = 54
	max_lom_size = 10240
	lom_disabled = false
	lom_compilation_delay = 0
	rules_snapshot_enabled = true
	rules_snapshot_max_count = 5
	rules_manipulation_locked = false
	heavy_lom = false
	parameters_count_weight = 6
	path_variativity_weight = 6
	pii_weight = 8
	request_content_weight = 6
	open_vulns_weight = 9
	serialized_data_weight = 6
	risk_score_algo = "maximum"
	pii_fallback = false
}
```

## Argument Reference

* `client_id` - (optional) ID of the client which is a partner for the created tenant. By default, this argument has the value of the current client ID.
* `min_lom_format` - (optional) Minimal Custom Ruleset format that will be compiled.
* `max_lom_format` - (optional) Maximum Custom Ruleset format that will be compiled.
* `max_lom_size` - (optional) Maximum size of Custom Ruleset size in bytes.
* `lom_disabled` - (optional) Field determining whether Custom Ruleset is compiled.
* `lom_compilation_delay` - (optional) Delay before Custom Ruleset compilition.
* `rules_snapshot_enabled` - (optional) Field determining whether rules snapshots are created during Custom Ruleset compilation.
* `rules_snapshot_max_count` - (optional) Maximum count of rules snapshot stored in wallarm.
* `rules_manipulation_locked` - (optional) Field determining whether rules might changed.
* `heavy_lom` - (optional) Field determining whether Custom Ruleset is compiled in special queue for huge rulesets.
* `parameters_count_weight` - (optional) [Risk Score][1] weight of query and body parameters. The more parameters, the more potential malicious payloads.
* `path_variativity_weight` - (optional) [Risk Score][1] weight of potential vulnerabilites to BOLA. Variable path parts make the endpoint a potential target of BOLA (IDOR) attacks.
* `pii_weight` - (optional) [Risk Score][1] weight of parameters with sensitive data. Parameters with sensitive data are always at risk of exposure.
* `request_content_weight` - (optional) [Risk Score][1] weight of uploading files to server. Attackers may be able to attack servers by uploading files containing malicious code.
* `open_vulns_weight` - (optional) [Risk Score][1] weight of active vulnerabilities. Active vulnerabilities may result in unauthorized data access or corruption.
* `serialized_data_weight` - (optional) [Risk Score][1] weight of accepting XML / JSON objects. XML / JSON objects are often used to transfer malicious payloads to attack servers.
* `risk_score_algo` - (optional) Method of [Risk Score][1] calulation. Specify how the risk score calculation should be performed. Available values: maximum, average.
* `pii_fallback` - (optional) Field determining whether fallback mechanism for PII detection is active.

[1]: https://docs.wallarm.com/api-discovery/overview/#endpoint-risk-score
