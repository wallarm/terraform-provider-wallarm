---
layout: "wallarm"
page_title: "Provider: Wallarm"
description: |-
  The Wallarm provider is used to interact with the Wallarm platform resources. The provider needs to be configured with the proper authentication credentials before it can be used.
---

# Wallarm Provider

The Wallarm provider is used to interact with the [Wallarm platform](https://docs.wallarm.com/) resources. The provider needs to be configured with the proper authentication credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Sets up Wallarm authentication credentials
# You can use environment variables instead
provider "wallarm" {
  api_token = "yXccqbq8o0zznJ5wMxzGmjvQ2RvmFAJZ6mFKF5Ka6n8fFpYaZBJHWIFBNXdeDhIG"
  api_host = "https://api.wallarm.com"
  client_id = 1111
}

}

# Adds a domain to the Scanner scope
resource "wallarm_scanner" "scope" {
  # ...
}

# Creates a rule to block the requests
resource "wallarm_rule_vpatch" "vpatch" {
  # ...
}
```

## Argument Reference

The following arguments are supported in `provider "wallarm"`:

* `api_token` - (**required**) your Wallarm [API token](https://docs.wallarm.com/user-guides/settings/api-tokens/). Note that the most operations with Wallarm API are allowed only for the users with the **Administrator** role. This can also be specified with the `WALLARM_API_TOKEN` shell environment variable.
* `api_host` - (optional) Wallarm API URL. Can be: `https://us1.api.wallarm.com` for the [US Cloud](https://docs.wallarm.com/about-wallarm/overview/#us-cloud), `https://api.wallarm.com` for the [EU Cloud](https://docs.wallarm.com/about-wallarm/overview/#eu-cloud). This can also be specified with the `WALLARM_API_HOST` shell environment variable. Default: `https://api.wallarm.com`.
* `client_id` - (optional) ID of the client (tenant). The value is required for [multi-tenant scenarios][2]. This can also be specified with the `WALLARM_API_CLIENT_ID` shell environment variable. Default: client ID of the authenticated user defined by api_token.
* `retries` - (optional) maximum number of retries to perform when an API request fails. Default: 3. This can also be specified with the `WALLARM_API_RETRIES` shell environment variable.
* `min_backoff` - (optional) minimum backoff period in seconds after failed API calls. Default: 1. This can also be specified with the `WALLARM_API_MIN_BACKOFF` shell environment variable.
* `max_backoff` - (optional) maximum backoff period in seconds after failed API calls Default: 30. This can also be specified with the `WALLARM_API_MAX_BACKOFF` shell environment variable.
* `api_client_logging` - (optional) whether to print logs from the API client (using the default log library logger). Default: false. This can also be specified with the `WALLARM_API_CLIENT_LOGGING` shell environment variable.

[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
