---
layout: "wallarm"
page_title: "Wallarm: wallarm_global_mode"
subcategory: "Common"
description: |-
  Provides the resource to set global modes for the filtration and Active threat verification component.
---

# wallarm_global_mode

Provides the resource to set global modes for the [filtration][1] and [Active threat verification][2] component.

## Example Usage

```hcl
# Sets filtration mode to use configuration defined locally on each node
# Scanner is disabled
# Active threat verification component (rechecker) is turned on

resource "wallarm_global_mode" "global_block" {
  filtration_mode = "default"
  rechecker_mode = "on"
}

```

## Argument Reference

* `filtration_mode` - (optional) global [filtration mode][1]. Possible values: `default`, `monitoring`, `block`, `safe_blocking`, `off`.

  Default: `default`

  Default: `on`
* `rechecker_mode` - (optional) the Active threat verification component mode. Possible values: `off`, `on`.

  Default: `off`

[1]: https://docs.wallarm.com/admin-en/configure-wallarm-mode/
[2]: https://docs.wallarm.com/user-guides/scanner/intro/#active-threat-verification
