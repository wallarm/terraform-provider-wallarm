---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_splunk"
subcategory: "Integration"
description: |-
  Provides the resource to manage Splunk integrations.
---

# wallarm_integration_splunk

Provides the resource to manage integrations to send [alerts to Splunk][1].

The types of events available to be sent to Splunk:
- Detected hits
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Scope changes: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications to Splunk

resource "wallarm_integration_splunk" "splunk_integration" {
  name = "New Terraform Splunk Integration"
  api_url = "https://example.com:8088"
  api_token = "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"

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
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the trigger to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `api_token` - (**required**) Splunk API token. Sensitive.
* `api_url` - (**required**) Splunk API URL with the schema (https://).

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
* `created_by` - email of the user which created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/splunk/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
