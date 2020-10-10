---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_sumologic"
sidebar_current: "docs-wallarm-resource-integration-sumologic"
description: |-
  Provides the resource to manage SumoLogic integrations.
---

# wallarm_integration_sumologic

Provides the resource to manage integrations to send notifications to SumoLogic.

The types of events available to be sent to SumoLogic:
- Hits detected
- System related: newly added users, deleted or disabled integration
- Vulnerabilities detected
- Scope changed: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates the integration to send notifications to SumoLogic

resource "wallarm_integration_sumologic" "sumologic_integration" {
  name = "New Terraform SumoLogic Integration"
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

* `client_id` - (Optional) ID of the client to apply the trigger to. The value is required for multi-tenant scenarios.
* `active` - (Optional) Indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent). 
Default: `false`
* `name` - (Optional) Integration name.
* `sumologic_url` - (Required) SumoLogic collector URL with the schema (https://).

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (Optional) Event type. Can be:
  - `hit` - Hits detected
  - `vuln` - Vulnerabilities detected
  - `system` - System related
  - `scope` - Scope changed

  Default: `vuln`
* `active` - (Optional) Indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
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

* `integration_id` - Integer ID of the created integration.
* `created_by` - Email of the user which created the integration.
* `is_active` - Indicator of the integration status. Can be: `true` and `false`.
