---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_opsgenie"
subcategory: "Integration"
description: |-
  Provides the resource to manage OpsGenie integrations.
---

# wallarm_integration_opsgenie

Provides the resource to manage integrations to send [notifications to OpsGenie][1].

The types of events available to be sent to OpsGenie:
- System related: newly added users, deleted or disabled integrations
- Rule and trigger changes
- Security issues (critical, high, medium, low, info)

## Example Usage

```hcl
# Creates an integration to send notifications to OpsGenie

resource "wallarm_integration_opsgenie" "opsgenie_integration" {
  name = "New Terraform OpsGenie Integration"
  api_url = "https://api.opsgenie.com/v2/alerts"
  api_token = "b035033e-540a-0390-aa00-a102e5b556a7"
  active = true

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
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `api_token` - (**required**) OpsGenie API token. Sensitive.
* `api_url` - (**required**) OpsGenie alerts API endpoint. If you're using the [EU instance](https://support.atlassian.com/opsgenie/docs/european-service-region) of OpsGenie, set the value to https://api.eu.opsgenie.com/v2/alerts. Otherwise, set it to https://api.opsgenie.com/v2/alerts.

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

[1]: https://docs.wallarm.com/user-guides/settings/integrations/opsgenie/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
