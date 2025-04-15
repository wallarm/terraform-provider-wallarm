---
layout: "wallarm"
page_title: "Wallarm: wallarm_integration_telegram"
subcategory: "Integration"
description: |-
  Provides the resource to manage Telegram integrations.
---

# wallarm_integration_telegram

Provides the resource to manage integrations to send [notifications to Telegram][1].

The types of events available to be sent to Telegram:
- System related: newly added users, deleted or disabled integration
- Detected vulnerabilities
- Changes in exposed assets: updates in hosts, services, and domains

## Example Usage

```hcl
# Creates an integration to send notifications to Slack

resource "wallarm_integration_telegram" "telegram_integration" {
  name = "my_tg_chat_name"
  active = false
  chat_data = "Z72S5kwWBPYpTnLrIfvYIQ=="
  token = "df8e6c1050180c3ce1b84fe7d6940070"

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
* `name` - (**required**) should be the same as telegram chat name
* `token` - (**required**) telegram token
* `chat_data` - (**required**) chat ID provided by Telegram

## Event

`event` are events for integration to monitor. Can be:

* `event_type` - (optional) event type. Can be:
  - `vuln` - detected vulnerabilities
  - `system` - System related
  - `scope` - scope changes

  Default: `vuln`
* `active` - (optional) indicator of the event type status. Can be: `true` for active events and `false` for disabled events (notifications are not sent). 
Default: `true`


Example:

```hcl
  # ... omitted

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

[1]: https://docs.wallarm.com/user-guides/settings/integrations/telegram/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
