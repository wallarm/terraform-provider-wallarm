---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_email"
subcategory: "Integration"
description: |-
  Provides the resource to manage Email integrations.
---

# wallarm_integration_email

Provides the resource to manage integrations to send email notifications.

The types of events available to be sent via email:
- System related: newly added users, deleted or disabled integrations
- Detected vulnerabilities
- Scope changes: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications to the specified emails

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

* `client_id` - (optional) ID of the client to apply the trigger to. The value is required for multi-tenant scenarios.
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `emails` - (**required**) list of emails where notifications should be sent to.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (optional) event type. Can be:
  - `vuln` - detected vulnerabilities
  - `system` - system related
  - `scope` - scope changes
  - `report_daily` - daily report
  - `report_weekly` - weekly report
  - `report_monthly` - monthly report

  Default: `vuln`
* `active` - (optional) indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
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

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user which created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.
