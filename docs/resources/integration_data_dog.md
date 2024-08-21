---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_data_dog"
subcategory: "Integration"
description: |-
  Provides the resource to manage DataDog integrations.
---

# wallarm_integration_data_dog

Provides the resource to manage integrations to send [notifications to DataDog][1].

The types of events available to be sent to DataDog:
- Detected hits
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Changes in exposed assets: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications to DataDog Logic

resource "wallarm_integration_data_dog" "data_dog_integration" {
  name = "New Terraform DataDog Integration"
  region = "US1"
  token = "eb7ddfc33acaaacaacaca55a398"

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
    active = false
  }
}

```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `token` - (required) DataDog token.
* `region` - (required) DataDog region.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (optional) event type. Can be:
  - `hit` - detected hits
  - `vuln_low` - detected vulnerabilities
  - `system` - system related
  - `scope` - scope changes

  Default: `vuln_low`
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

[1]: https://docs.wallarm.com/user-guides/settings/integrations/datadog/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
