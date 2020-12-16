---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_opsgenie"
subcategory: "Integration"
description: |-
  Provides the resource to manage OpsGenie integrations.
---

# wallarm_integration_opsgenie

Provides the resource to manage integrations to send notifications to OpsGenie.

The types of events available to be sent to OpsGenie:
- Hits detected
- Vulnerabilities detected

## Example Usage

```hcl
# Creates the integration to send notifications to OpsGenie

resource "wallarm_integration_opsgenie" "opsgenie_integration" {
  name = "New Terraform OpsGenie Integration"
  api_token = "b035033e-540a-0390-aa00-a102e5b556a7"
  active = true

  event {
    event_type = "hit"
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
* `api_token` - (Required) OpsGenie API token. Sensitive.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (Optional) Event type. Can be:
  - `hit` - Hits detected
  - `vuln` - Vulnerabilities detected

  Default: `vuln`
* `active` - (Optional) Indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
Default: `true`


Example:

```hcl
  # ... omitted

  event {
    event_type = "hit"
    active = false
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
