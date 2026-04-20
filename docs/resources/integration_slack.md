---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_slack"
subcategory: "Integrations"
description: |-
  Provides the resource to manage Slack integrations.
---

# wallarm_integration_slack

Provides the resource to manage integrations to send [notifications to Slack][1].

The types of events available to be sent to Slack:
- System related: newly added users, deleted or disabled integrations
- Rule and trigger changes
- Security issues (critical, high, medium, low, info)

## Example Usage

```hcl
# Creates an integration to send notifications to Slack

resource "wallarm_integration_slack" "slack_integration" {
  name = "New Terraform Slack Integration"
  webhook_url = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

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
    active = false
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `webhook_url` - (**required**) Slack Webhook URL. Sensitive.

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
    active = true
  }

  event {
    event_type = "security_issue_critical"
    active = false
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user who created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

## Import

```
$ terraform import wallarm_integration_slack.example 1111/slack/2222
```

* `1111` - Client ID.
* `slack` - Integration type constant (literal).
* `2222` - Integration ID.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/slack/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
