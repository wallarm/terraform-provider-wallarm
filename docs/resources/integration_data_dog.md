---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_data_dog"
subcategory: "Integrations"
description: |-
  Provides the resource to manage DataDog integrations.
---

# wallarm_integration_data_dog

Provides the resource to manage integrations to send [notifications to DataDog][1].

The types of events available to be sent to DataDog:
- SIEM events (detected hits and related data)
- System related: newly added users, deleted or disabled integrations
- Rule and trigger changes
- Security issues (critical, high, medium, low, info)
- Request volume monitoring

## Example Usage

```hcl
# Creates an integration to send notifications to DataDog

resource "wallarm_integration_data_dog" "data_dog_integration" {
  name = "New Terraform DataDog Integration"
  region = "US1"
  token = "eb7ddfc33acaaacaacaca55a398abcdef01"

  event {
    event_type = "siem"
    active = true
    with_headers = true
  }

  event {
    event_type = "system"
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
* `token` - (required) DataDog API key. Must be 32-40 characters. Sensitive.
* `region` - (required) DataDog region.

## Event

`event` are events for integration to monitor. Can be:
```
event {

* `event_type`: (optional) event type. Options:
  - `siem` - SIEM events: Detected hits, original request data, and malicious payloads.
  - `rules_and_triggers` - rule and trigger changes
  - `number_of_requests_per_hour` - number of requests per hour
  - `security_issue_critical` - critical security issues
  - `security_issue_high` - high severity security issues
  - `security_issue_medium` - medium severity security issues
  - `security_issue_low` - low severity security issues
  - `security_issue_info` - informational security issues
  - `system` - system related (newly added users, deleted or disabled integrations)
* `with_headers`: (Optional) Include request headers in event data (for `siem` event type only). Can be `true` or `false`. Default: `false`.
* `active`: `true` for active events (notifications sent), `false` for disabled events (no notifications). Default: `true`.

}
```


Example:

```hcl
  # ... omitted

  event {
    event_type = "siem"
    active = true
    with_headers = true
  }

  event {
    event_type = "system"
    active = true
  }

  event {
    event_type = "security_issue_critical"
    active = false
  }

  event {
    event_type = "security_issue_high"
    active = false
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user who created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/datadog/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
