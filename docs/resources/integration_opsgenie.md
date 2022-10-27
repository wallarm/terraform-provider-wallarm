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
- Detected hits
- Detected vulnerabilities

## Example Usage

```hcl
# Creates an integration to send notifications to OpsGenie

resource "wallarm_integration_opsgenie" "opsgenie_integration" {
  name = "New Terraform OpsGenie Integration"
  api_url = "https://api.opsgenie.com/v2/alerts"
  api_token = "b035033e-540a-0390-aa00-a102e5b556a7"
  active = true

  event {
    event_type = "hit"
    active = true
  }
  
  event {
    event_type = "vuln"
    active = true
  }
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the trigger to. The value is required for multi-tenant scenarios.
* `active` - (optional) indicator of the integration status. Can be: `true` for active integration and `false` for disabled integration (notifications are not sent).

  Default: `false`
* `name` - (optional) integration name.
* `api_token` - (**required**) OpsGenie API token. Sensitive.
* `api_url` - (**required**) OpsGenie alerts API endpoint. If you're using the [EU instance](https://support.atlassian.com/opsgenie/docs/european-service-region) of OpsGenie, set the value to https://api.eu.opsgenie.com/v2/alerts. Otherwise, set it to https://api.opsgenie.com/v2/alerts.

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (optional) event type. Can be:
  - `hit` - detected hits
  - `vuln` - detected vulnerabilities

  Default: `vuln`
* `active` - (optional) indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
Default: `true`


Example:

```hcl
  # ... omitted

  event {
    event_type = "hit"
    active = false
  }
  
  event {
    event_type = "vuln"
    active = false
  }

  # ... omitted
```

## Attributes Reference

* `integration_id` - integer ID of the created integration.
* `created_by` - email of the user which created the integration.
* `is_active` - indicator of the integration status. Can be: `true` and `false`.

[1]: https://docs.wallarm.com/user-guides/settings/integrations/opsgenie/
