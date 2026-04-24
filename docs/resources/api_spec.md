---
layout: "wallarm"
page_title: "Wallarm: wallarm_api_spec"
subcategory: "API Specification Enforcement"
description: |-
  Manages an uploaded OpenAPI specification used by Wallarm's API Specification Enforcement.
---

# wallarm_api_spec

Provides the resource to upload and manage an [OpenAPI specification][1] used by Wallarm's [API Specification Enforcement][2] feature. Wallarm validates incoming requests against the spec and can apply a policy (see `wallarm_api_spec_policy`) that decides what to do on each violation type.

Only URL-hosted specs are supported by this provider. For specs uploaded directly via the Wallarm console, import the existing resource (see the Import section).

## Example Usage

```hcl
resource "wallarm_api_spec" "petstore" {
  client_id           = 6039
  title               = "Petstore"
  description         = "Public-facing Petstore API"
  file_remote_url     = "https://raw.githubusercontent.com/acme/petstore/main/openapi.yaml"
  regular_file_update = true
  api_detection       = true
  domains             = ["petstore.example.com"]
  instances           = [1]

  auth_headers {
    key   = "X-Source-Token"
    value = var.source_token
  }
}
```

## Argument Reference

### Required

* `client_id` - (required, ForceNew) ID of the client to apply changes. Immutable.
* `title` - (required) human-readable spec title.
* `file_remote_url` - (required) URL that serves the OpenAPI 3.0 / 3.1 spec in JSON or YAML.

### Optional

* `description` - (optional) free-text description.
* `regular_file_update` - (optional) when `true`, Wallarm refreshes the spec from `file_remote_url` hourly. Default: `false`.
* `api_detection` - (optional) when `true`, Wallarm uses the spec for API discovery. Default: `false`.
* `domains` - (optional) list of domains the spec applies to.
* `instances` - (optional) list of instance (application) IDs the spec applies to.
* `auth_headers` - (optional) list of `{key, value}` blocks sent when Wallarm fetches `file_remote_url`. `value` is marked Sensitive.

## Attributes Reference

* `api_spec_id` - ID of the uploaded specification.
* `status` - upload status, e.g. `ready`.
* `spec_version` - version declared in the OpenAPI spec's `info.version`.
* `openapi_version` - OpenAPI format version (`3.0.0`, `3.1.0`, ...).
* `endpoints_count` - number of endpoints parsed from the spec.
* `shadow_endpoints_count`, `orphan_endpoints_count`, `zombie_endpoints_count` - Wallarm-specific endpoint categorization counts.
* `format` - Wallarm-internal spec format discriminator.
* `version` - internal version counter incremented on spec updates.
* `node_sync_version` - internal sync version used by filtering nodes.
* `last_synced_at`, `last_compared_at`, `updated_at`, `created_at`, `file_changed_at` - timestamps (RFC 3339).
* `file` - metadata of the stored spec file: `name`, `signed_url`, `checksum`, `mime_type`, `version`. `signed_url` is Sensitive and **regenerates on every Read** (expires ~10 minutes after the last refresh). Use it immediately after `terraform refresh`; for durable access, download via the Wallarm console.

## Import

```
$ terraform import wallarm_api_spec.petstore 6039/134172
```

* `6039` - Client ID.
* `134172` - API Spec ID.

Every field is populated on import; the only field excluded from `ImportStateVerify` is `file.0.signed_url` (see note above).

## Limitations

File-upload mode (local `.yaml`/`.json` file without a URL) is not supported by this provider in v2.3.6. Use a URL-hosted spec or import a console-uploaded spec.

[1]: https://spec.openapis.org/oas/latest.html
[2]: https://docs.wallarm.com/api-specification-enforcement/overview/
