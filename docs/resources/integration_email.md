---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_email"
subcategory: "Integrations"
description: |-
  Provides the resource to manage Email integrations.
---

# wallarm_integration_email

Provides the resource to manage integrations to send [email notifications][1].

The types of events available to be sent via email:
- System related: newly added users, deleted or disabled integrations
- Security issues discovered by the platform
- API discovery reports (hourly and daily changes)
- Scheduled reports (daily, weekly, monthly)

## Example Usage

```hcl
# Creates an integration to send notifications to the specified emails

resource "wallarm_integration_email" "email_integration" {
  name = "New Email Terraform Integration"
  active = true
  emails = ["test@wallarm.com", "test2@wallarm.com"]

  event {
    event_type = "system"
    active = true
  }

  event {
    event_type = "aasm_report"
    active = true
  }

  event {
    event_type = "report_daily"
    active = true
  }

  event {
    event_type = "report_weekly"
    active = true
  }

  event {
    event_type = "report_monthly"
    active = true
  }

  event {
    event_type = "api_discovery_hourly_changes_report"
    active = true
  }

  event {
    event_type = "api_discovery_daily_changes_report"
    active = true
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `emails` - (**required**) list of emails where notifications should be sent to.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (**required**) event type. Can be:
  - `system` - system related (newly added users, deleted or disabled integrations)
  - `aasm_report` - automated attack summary report
  - `api_discovery_hourly_changes_report` - hourly API discovery changes report
  - `api_discovery_daily_changes_report` - daily API discovery changes report
  - `report_daily` - daily report
  - `report_weekly` - weekly report
  - `report_monthly` - monthly report

  MaxItems: 7
* `active` - (optional) indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent).
Default: `true`


Example:

```hcl
  # ... omitted

  event {
    event_type = "system"
    active = true
  }

  event {
    event_type = "aasm_report"
    active = true
  }

  event {
    event_type = "report_daily"
    active = true
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user who created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

## Import

```
$ terraform import wallarm_integration_email.example 1111/email/2222
```

* `1111` - Client ID.
* `email` - Integration type constant (literal).
* `2222` - Integration ID.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/email/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
