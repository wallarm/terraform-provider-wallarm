---
layout: "wallarm"
page_title: "Wallarm: wallarm_global_mode"
subcategory: "Common"
description: |-
  Provides the resource to set global modes for the filtration, scanner, and Active threat verification component.
---

# wallarm_global_mode

  Provides the resource to set global modes for the [filtration][1], [scanner][2], and [Active threat verification][3] component.

## Example Usage

```hcl
# Sets filtration mode to use configuration defined locally on each node
# Scanner is disabled
# Active threat verification component (rechecker) is turned on

resource "wallarm_global_mode" "global_block" {
  filtration_mode = "default"
  scanner_mode = "off"
  rechecker_mode = "on"
}

```

## Argument Reference

* `filtration_mode` - (optional) global [filtration mode][1]. Possible values: `default`, `monitoring`, `block`, `safe_blocking`, `off`.

  Default: `default`
* `scanner_mode` - (optional) scanner mode. Possible values: `off`, `on`.

  Default: `on`
* `rechecker_mode` - (optional) Active threat verification component mode. Possible values: `off`, `on`.

  Default: `off`

[1]: https://docs.wallarm.com/admin-en/configure-wallarm-mode/
[2]: https://docs.wallarm.com/user-guides/scanner/intro/
[3]: https://docs.wallarm.com/user-guides/scanner/intro/#active-threat-verification
