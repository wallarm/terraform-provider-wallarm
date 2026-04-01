---
layout: "wallarm"
page_title: "Wallarm: wallarm_global_mode"
subcategory: "Common"
description: |-
  Provides the resource to manage global filtration mode, rechecker mode, and overlimit request settings.
---

# wallarm_global_mode

Provides the resource to manage global [filtration mode][1], attack rechecker mode, and overlimit request processing settings.

## Example Usage

```hcl
# Set filtration mode to blocking, rechecker off, overlimit monitoring
resource "wallarm_global_mode" "settings" {
  filtration_mode = "block"
  rechecker_mode  = "off"
  overlimit_time  = 1000
  overlimit_mode  = "monitoring"
}
```

```hcl
# Multi-tenant: configure per client
resource "wallarm_global_mode" "tenant_settings" {
  client_id       = 8649
  filtration_mode = "monitoring"
  rechecker_mode  = "off"
}
```

## Argument Reference

* `client_id` - (Optional) Client ID. Defaults to the provider's client ID.

* `filtration_mode` - (Optional) Global [filtration mode][1]. Possible values: `default`, `monitoring`, `block`, `safe_blocking`, `off`. Default: `default`.

* `rechecker_mode` - (Optional) Attack rechecker mode. Possible values: `on`, `off`. Default: `off`.

* `overlimit_time` - (Optional) Time limit for single request processing in milliseconds. Range: 0–10000.

* `overlimit_mode` - (Optional) Action when `overlimit_time` is exceeded. Possible values: `blocking`, `monitoring`.

## Attributes Reference

* `client_id` - The Client ID for this resource.

## Import

```bash
terraform import wallarm_global_mode.settings 8649/global_mode
```

The import ID format is `{clientID}/global_mode`.

[1]: https://docs.wallarm.com/admin-en/configure-wallarm-mode/
