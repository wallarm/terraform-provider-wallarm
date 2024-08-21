---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_ms_teams"
subcategory: "Integration"
description: |-
  Provides the resource to manage MS Teams integrations.
---

# wallarm_integration_ms_teams

Provides the resource to manage [integrations via MS Teams][1].

The types of events available to be sent via MS Teams:
- Detected hits
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Changes in exposed assets: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications via
# MS Teams to the provided URL and corresponding HTTP method

resource "wallarm_integration_ms_teams" "teams_integration" {
  name = "New Terraform MS Teams Integration"
  webhook_url = "https://gar8347sk.webhook.office.com/webhookb2/a92734837"
  active = true
  
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
    event_type = "vuln"
    active = true
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
Default: `POST`
* `webhook_url` - (required) MS Teams URL with the schema (https://).
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
    event_type = "vuln"
    active = false
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user who created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/microsoft-teams/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
