---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_insightconnect"
subcategory: "Integration"
description: |-
  Provides the resource to manage InsightConnect integrations.
---

# wallarm_integration_insightconnect

Provides the resource to manage integrations to send [notifications to InsightConnect][1].

The types of events available to be sent to InsightConnect:
- Detected hits
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Scope changes: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications to InsightConnect

resource "wallarm_integration_insightconnect" "insight_integration" {
  name = "New Terraform InsightConnect Integration"
  api_url = "https://us.api.insight.rapid7.com/connect/v1/workflows/d1757a97-e41a-2030-a641-26c1435657081/events/execute"
  api_token = "b035033e-540a-0390-aa00-a102e5b556a7"

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
* `api_token` - (**required**) InsightConnect API token. Sensitive.
* `api_url` - (**required**) InsightConnect API URL with the schema (https://).

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

[1]: https://docs.wallarm.com/user-guides/settings/integrations/insightconnect/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
