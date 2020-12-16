---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_insightconnect"
subcategory: "Integration"
description: |-
  Provides the resource to manage InsightConnect integrations.
---

# wallarm_integration_insightconnect

Provides the resource to manage integrations to send notifications to InsightConnect.

The types of events available to be sent to InsightConnect:
- Hits detected
- System related: newly added users, deleted or disabled integration
- Vulnerabilities detected
- Scope changed: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates the integration to send notifications to InsightConnect

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

* `client_id` - (Optional) ID of the client to apply the trigger to. The value is required for multi-tenant scenarios.
* `active` - (Optional) Indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent). 
Default: `false`
* `name` - (Optional) Integration name.
* `api_token` - (Required) InsightConnect API token. Sensitive.
* `api_url` - (Required) InsightConnect API URL with the schema (https://).

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (Optional) Event type. Can be:
  - `hit` - Hits detected
  - `vuln` - Vulnerabilities detected
  - `system` - System related
  - `scope` - Scope changed

  Default: `vuln`
* `active` - (Optional) Indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
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

* `integration_id` - Integer ID of the created integration.
* `created_by` - Email of the user which created the integration.
* `is_active` - Indicator of the integration status. Can be: `true` and `false`.
