---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_pagerduty"
subcategory: "Integration"
description: |-
  Provides the resource to manage PagerDuty integrations.
---

# wallarm_integration_pagerduty

Provides the resource to manage integrations to send [notifications to PagerDuty][1].

The types of events available to be sent to PagerDuty:
- System related: newly added users, deleted or disabled integrations
- Rule and trigger changes
- Security issues (critical, high, medium, low, info)

## Example Usage

```hcl
# Creates an integration to send notifications to PagerDuty

resource "wallarm_integration_pagerduty" "pagerduty_integration" {
  name = "New Terraform PagerDuty Integration"
  integration_key = "48c8f1999cbf4b91a3dbd0fac79bfc6b"

  event {
    event_type = "system"
    active = true
  }

  event {
    event_type = "rules_and_triggers"
    active = true
  }

  event {
    event_type = "security_issue_critical"
    active = true
  }

  event {
    event_type = "security_issue_high"
    active = true
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `integration_key` - (**required**) PagerDuty Integration key. Sensitive.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (optional) event type. Can be:
  - `system` - system related (newly added users, deleted or disabled integrations)
  - `rules_and_triggers` - rule and trigger changes
  - `security_issue_critical` - critical security issues
  - `security_issue_high` - high severity security issues
  - `security_issue_medium` - medium severity security issues
  - `security_issue_low` - low severity security issues
  - `security_issue_info` - informational security issues

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
    event_type = "rules_and_triggers"
    active = false
  }

  event {
    event_type = "security_issue_critical"
    active = true
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user who created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/pagerduty/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
