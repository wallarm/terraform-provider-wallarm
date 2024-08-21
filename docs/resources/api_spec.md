---
layout: "wallarm"
page_title: "Wallarm: wallarm_api_spec"
subcategory: "API Specification"
description: |-
  Provides the resource to manage API Specs[1] in Wallarm.
---

# wallarm_api_spec

Provides the resource to manage API Spec in Wallarm.

## Example Usage

```hcl
# Creates an API specification for Wallarm
resource "wallarm_api_spec" "api_spec" {
  client_id          = 1
  title              = "Example API Spec"
  description        = "This is an example API specification created by Terraform."
  file_remote_url    = "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/main/examples/v3.0/api-with-examples.yaml"
  regular_file_update = true
  api_detection      = true
  domains            = ["ex.com"]
  instances          = [1]
}
```

## Argument Reference

* `client_id` -  (required) ID of the client to apply the API specification to.
* `title` - (required) The title of the API specification.
* `description` - (optional) A description of the API specification.
* `file_remote_url` - (required) The remote URL to the API specification file. This is useful for pulling specifications from external sources.
* `regular_file_update` - (optional) Indicator of whether the API specification file should be regularly updated from the file_remote_url. Can be true or false. Default: false.
* `api_detection` - (optional) Indicator of whether Wallarm should automatically detect APIs based on this specification.
* `domains` - (required) List of domains associated with the API.
* `instances` - (required) List of Wallarm node instances where the API specification should be applied.

## Attributes Reference
* `api_spec_id` - Integer ID of the created API specification.
[1]: https://docs.wallarm.com/api-specification-enforcement/overview/
