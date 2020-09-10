---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_email"
sidebar_current: "docs-wallarm-resource-integration-insightconnect"
description: |-
  Provides the resource to manage Email integrations.
---

# wallarm_integration_email

Provides the resource to manage integrations to send notifications to Email.

The types of events available to be sent to Email:
- System related: newly added users, deleted or disabled integration
- Vulnerabilities detected
- Scope changed: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates the integration to send notifications to defined Emails

resource "wallarm_integration_email" "email_integration" {
  name = "New Email Terraform Integration"
  active = true
  emails = ["test@wallarm.com", "test2@wallarm.com"]
  
  event {
    event_type = "report_monthly"
    active = true
  }
  
  event {
    event_type = "report_weekly"
    active = true
  }

  event {
    event_type = "report_daily"
    active = true
  }

  event {
    event_type = "system"
    active = true
  }

  event {
    event_type = "scope"
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
* `emails` - (Required) List of emails where notifications should be sent to.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (Optional) Event type. Can be:
  - `vuln` - Vulnerabilities detected
  - `system` - System related
  - `scope` - Scope changed
  - `report_daily` - Daily report
  - `report_weekly` - Weekly report
  - `report_monthly` - Monthly report

  Default: `vuln`
* `active` - (Optional) Indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
Default: `true`


Example:

```hcl
  # ... omitted

  event {
    event_type = "report_monthly"
    active = true
  }
  
  event {
    event_type = "report_weekly"
    active = true
  }

  event {
    event_type = "report_daily"
    active = true
  }

  event {
    event_type = "system"
    active = true
  }

  event {
    event_type = "scope"
    active = true
  }
  
  event {
    event_type = "vuln"
    active = true
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - Integer ID of the created integration.
* `created_by` - Email of the user which created the integration.
* `is_active` - Indicator of the integration status. Can be: `true` and `false`.
