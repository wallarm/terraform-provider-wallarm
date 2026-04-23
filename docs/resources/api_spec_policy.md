---
layout: "wallarm"
page_title: "Wallarm: wallarm_api_spec_policy"
subcategory: "API Specification Enforcement"
description: |-
  Manages the enforcement policy attached to an uploaded API specification.
---

# wallarm_api_spec_policy

Provides the resource to configure [API Specification Enforcement][1] for an uploaded spec — how Wallarm reacts when a request violates the OpenAPI contract (undefined endpoints, missing parameters, schema mismatches, oversized bodies, enforcement timeouts). The parent spec is managed by [`wallarm_api_spec`](./api_spec.md); each spec has at most one policy.

Every violation/threshold category has its own mode, letting you pick which classes of misuse to block, which to only log (`monitor`), and which to pass through (`ignore`). Scope the policy with `conditions` to limit enforcement to specific hosts, paths, methods, or applications.

## Example Usage

```hcl
resource "wallarm_api_spec" "petstore" {
  client_id       = 6039
  title           = "Petstore"
  file_remote_url = "https://raw.githubusercontent.com/acme/petstore/main/openapi.yaml"
  domains         = ["petstore.example.com"]
  instances       = [1]
}

resource "wallarm_api_spec_policy" "petstore" {
  client_id   = wallarm_api_spec.petstore.client_id
  api_spec_id = wallarm_api_spec.petstore.api_spec_id

  # Block the most common spec violations.
  undefined_endpoint_mode      = "block"
  undefined_parameter_mode     = "monitor"
  missing_parameter_mode       = "monitor"
  invalid_parameter_value_mode = "monitor"
  missing_auth_mode            = "block"
  invalid_request_mode         = "monitor"

  # Enforcement thresholds — fall back to monitor on runaway processing.
  timeout_mode          = "monitor"
  max_request_size_mode = "monitor"
  timeout               = 50
  max_request_size      = 1024

  # Restrict enforcement to the production host.
  conditions {
    type  = "iequal"
    value = "petstore.example.com"
    point = {
      header = "HOST"
    }
  }
}
```

## Argument Reference

### Required

* `client_id` - (required, ForceNew) ID of the client that owns the parent API spec.
* `api_spec_id` - (required, ForceNew) ID of the spec this policy enforces. Reference `wallarm_api_spec.<name>.api_spec_id` to create both in a single apply.

### Optional

* `enabled` - (optional) whether enforcement is actively applied. Defaults to `true`. Setting to `false` keeps all other fields stored on the spec but disables the policy at runtime.

#### Scope

* `conditions` - (optional) set of scope conditions limiting where the policy applies. Same schema and semantics as the `action {}` block on rule resources — see the [Action Guide](../guides/action) for full documentation on condition types (`equal`, `iequal`, `regex`, `absent`) and points (`header`, `path`, `method`, `instance`, etc.). If omitted, the policy applies to all traffic covered by the parent spec's `domains` / `instances`.

#### Violation Modes

Each violation mode accepts `block`, `monitor`, or `ignore` and defaults to `monitor`.

* `undefined_endpoint_mode` - request hits a URL not declared in the spec.
* `undefined_parameter_mode` - request carries a parameter not declared on the matched endpoint.
* `missing_parameter_mode` - a parameter marked `required` in the spec is absent from the request.
* `invalid_parameter_value_mode` - a parameter value does not match its declared type, format, or enum.
* `missing_auth_mode` - request lacks the authentication declared by the matched endpoint (API key, bearer token, etc.).
* `invalid_request_mode` - request body does not validate against the declared request schema.

#### Threshold Limits

Threshold modes accept `block` or `monitor` (default `monitor`). `ignore` is **not** supported for thresholds.

* `timeout_mode` - action when spec-enforcement processing exceeds `timeout`.
* `max_request_size_mode` - action when the request body exceeds `max_request_size`.
* `timeout` - (optional) max spec-processing time per request, in milliseconds. Default: `50`.
* `max_request_size` - (optional) max inspected request body size, in kilobytes. Default: `1024`.

## Attributes Reference

No computed-only attributes. Every field round-trips through Read — the resource ID is `{client_id}/{api_spec_id}` and ingestion metadata (status, endpoint counts, etc.) lives on the parent `wallarm_api_spec`.

## Import

```
$ terraform import wallarm_api_spec_policy.petstore 6039/134172
```

* `6039` - Client ID.
* `134172` - API Spec ID.

All fields are populated on import; the resource is fully compatible with `ImportStateVerify`.

## Notes & Limitations

* **Soft delete.** `terraform destroy` on this resource does **not** remove the policy record — it issues a PUT with `enabled = false` and preserves every other setting. Re-creating the resource restores enforcement without needing to reconfigure individual modes. To remove the policy record entirely, destroy the parent `wallarm_api_spec`.
* **One policy per spec.** The API permits at most one policy per `api_spec_id`. Creating a second `wallarm_api_spec_policy` pointing at the same spec overwrites the existing one.
* **Threshold modes do not accept `ignore`.** Wallarm always takes some action on threshold breaches — either blocking or logging.

[1]: https://docs.wallarm.com/api-specification-enforcement/overview/
