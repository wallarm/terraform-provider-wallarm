---
layout: "wallarm"
page_title: "Wallarm: wallarm_global_mode"
subcategory: "Common"
description: |-
  Provides the resource to set global modes of scanner, attack rechecker and WAF.
---

# wallarm_user

  Provides the resource to set global modes of scanner, attack rechecker and WAF.

## Example Usage

```hcl
# Sets WAF mode to use local defined configuration
# Scanner is disabled
# Attack Rechecker is turned on

resource "wallarm_global_mode" "global_block" {
  waf_mode = "default"
  scanner_mode = "off"
  rechecker_mode = "on"
}

```

## Argument Reference

* `waf_mode` - (Optional) WAF mode. Possible values: `default`, `monitoring`, `block`
* `scanner_mode` - (Optional) Scanner mode. Possible values: `off`, `on`
* `rechecker_mode` - (Optional) Attack rechecker mode. Possible values: `off`, `on`
