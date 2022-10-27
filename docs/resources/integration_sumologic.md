---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_sumologic"
subcategory: "Integration"
description: |-
  Provides the resource to manage Sumo Logic integrations.
---

# wallarm_integration_sumologic

Provides the resource to manage integrations to send [notifications to Sumo Logic][1].

The types of events available to be sent to Sumo Logic:
- Detected hits
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Scope changes: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications to Sumo Logic

resource "wallarm_integration_sumologic" "sumologic_integration" {
  name = "New Terraform Sumo Logic Integration"
  sumologic_url = "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA=="

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
  
  event {
    event_type = "vuln"
    active = false
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the integration to. The value is required for [multi-tenant scenarios][2].
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `sumologic_url` - (**required**) Sumo Logic collector URL with the schema (https://).

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

[1]: https://docs.wallarm.com/user-guides/settings/integrations/sumologic/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
