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
- Detected hits
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Changes in exposed assets: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications to PagerDuty

resource "wallarm_integration_pagerduty" "pagerduty_integration" {
  name = "New Terraform PagerDuty Integration"
  integration_key = "48c8f1999cbf4b91a3dbd0fac79bfc6b"

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
    event_type = "vuln_high"
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
  - `hit` - detected hits
  - `vuln` - detected vulnerabilities
  - `system` - system related
  - `scope` - scope changes

  Default: `vuln`
* `active` - (optional) indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
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
    event_type = "vuln_high"
    active = false
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user who created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/pagerduty/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
