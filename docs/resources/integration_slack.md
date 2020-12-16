---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_slack"
subcategory: "Integration"
description: |-
  Provides the resource to manage Slack integrations.
---

# wallarm_integration_slack

Provides the resource to manage integrations to send notifications to Slack.

The types of events available to be sent to Slack:
- System related: newly added users, deleted or disabled integration
- Vulnerabilities detected
- Scope changed: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates the integration to send notifications to Slack

resource "wallarm_integration_slack" "slack_integration" {
  name = "New Terraform Slack Integration"
  webhook_url = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

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
* `webhook_url` - (Required) Slack Webhook URL. Sensitive.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (Optional) Event type. Can be:
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
