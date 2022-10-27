---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_webhook"
subcategory: "Integration"
description: |-
  Provides the resource to manage Webhook integrations.
---

# wallarm_integration_webhook

Provides the resource to manage [integrations via generic webhooks][1]. Webhooks can be used as system log sources. The number of log sources depends on the system complexity: the more components in the system, the greater number of log sources and logs.

The types of events available to be sent via WebHooks:
- Detected hits
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Scope changes: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications via
# webhooks to the provided URL and corresponding HTTP method

resource "wallarm_integration_webhook" "wh_integration" {
  name = "New Terraform WebHook Integration"
  webhook_url = "https://example.com/api/v1/webhook/"
  http_method = "POST"
  active = true
  
  event {
    event_type = "hit"
    active = true
  }

  event {
    event_type = "scope"
    active = true
  }

  event {
    event_type = "system"
    active = true
  }
  
  event {
    event_type = "vuln"
    active = true
  }

  headers = {
    Authorization = "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg=="
    Content-Type = "application/xml"
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `http_method` - (optional) HTTP method via which requests are to be sent. Can be: `POST`, `PUT`. 
Default: `POST`
* `webhook_url` - (optional) Webhook URL with the schema (https://).
* `ca_file` - (optional) CA certificate if needed by webhook collector.
* `ca_verify` - (optional) indicator of the SSL/TLS certificate verification. Can be: `true` or `false`.
Default: `true`
* `timeout` - (optional) time in seconds to raise a timeout error whilst connecting to the specified Webhook URL. 
Default: 15
* `open_timeout` - (optional) time in seconds to raise a timeout error while opening a TCP connection to the specified Webhook URL.
Default: 20
* `headers` - (optional) HTTP headers required by the Webhook endpoint. For instance, basic authentication can be set. 
Type: `map`

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (optional) event type. Can be:
  - `hit` - detected hits
  - `vuln` - detected vulnerabilities
  - `system` - system related
  - `scope` - scope changes

  Default: `vuln`
* `active` - (optional) indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
Default: `true`


Example:

```hcl
  # ... omitted

  event {
    event_type = "hit"
    active = true
  }

  event {
    event_type = "scope"
    active = false
  }

  event {
    event_type = "system"
    active = true
  }
  
  event {
    event_type = "vuln"
    active = false
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user who created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/webhook/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
