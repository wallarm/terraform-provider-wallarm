---
layout: "wallarm"
page_title: "Wallarm: wallarm_trigger"
subcategory: "Common"
description: |-
  Provides the resource to manage triggers.
---

# wallarm_trigger

Provides the resource to manage [triggers][1]. Triggers are tools that are used to set up custom notifications and reactions to events. Using triggers, you can receive alerts on major events via the tools you use for your day-to-day workflow, for example via corporate messengers or incident management systems.

To reduce the amount of noise, you can also configure the parameters of events to be notified about. The following events are available for setup:

- Attacks
- Incidents
- Hits
- Users added to the account
- Brute-force attacks

To receive notifications and reports, you can use Slack, email, Sumo Logic and other [integrations](https://docs.wallarm.com/user-guides/settings/integrations/integrations-intro/).

## Example Usage

```hcl
# Creates the trigger to send a notification to
# Slack (existing integration ID is 521) if a new user
# is added to the company account in Wallarm Console.

resource "wallarm_trigger" "user_trigger" {
  name = "New Terraform Trigger Slack"
  comment = "This is a description set by Terraform"
  enabled = true
  template_id = "user_created"

  actions {
    action_id = "send_notification"
    integration_id = [521]
  }
}
```

Trigger with the integration together:

```hcl
# Create an application to use it further

resource "wallarm_application" "tf_app" {
  name = "New Terraform Application"
  app_id = 43
}

# Create an email integration to get its ID afterward

resource "wallarm_integration_email" "email_integration" {
  name = "New Terraform Integration"
  emails = ["test1@example.com", "test2@example.com"]

  event {
    event_type = "report_monthly"
    active = true
  }
  
  event {
    event_type = "vuln"
    active = true
  }

}

# Create a trigger:
# When the number of attacks from IP address 2.2.2.2 and
# directed to the application with ID 43 exceeds 10000 for 86400 seconds
# then send notification to the email integration.

resource "wallarm_trigger" "attack_trigger" {
  name = "New Terraform Trigger Email"
  enabled = false
  template_id = "attacks_exceeded"

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_email.email_integration.integration_id]
  }

  depends_on = [
    wallarm_integration_email.email_integration,
    wallarm_application.tf_app,
  ]
}
```


## Argument Reference

* `client_id` - (optional) ID of the client to apply the trigger to. The value is required for [multi-tenant scenarios][2].
* `template_id` - (**required**) trigger condition. A condition is a system event to be notified about. Can be:
  - `user_created` for a user added to the company account in Wallarm Console
  - `attacks_exceeded` for detected attacks number exceeded the specified value
  - `hits_exceeded` for detected hits number exceeded the specified value
  - `incidents_exceeded` for detected incidents number exceeded the specified value
  - `vector_attack` for detected malicious payloads number exceeded the specified value
  - `bruteforce_started` for detected attack to be identified as brute-force
  - `bola_search_started` for detected attack identified as BOLA
  - `blacklist_ip_added` for IP list added to denylist
  - `api_structure_changed` for detected new, changed or unused endpoints
  - `attack_ip_grouping` for threshold to group hits from the same IP into one address
  - `compromised_logins` for attempts to use compromised user credentials
  - `rogue_api_detected` for Shadow, Orphan or Zombie API detected
* `enabled` - (optional) indicator of the trigger status. Can be: `true` for enabled trigger and `false` for disabled trigger (notifications are not sent).
* `name` - (optional) Trigger name.
* `comment` - (optional) Trigger description.
* `filters` - (optional) Filters for trigger conditions. Possible attributes are described below.
* `threshold` - (optional) Limitations for trigger conditions. Possible attributes are described below.
* `actions` - (optional) Trigger actions. Possible attributes are described below.

## Filters

`filters` are filters for trigger conditions. Can be:

* `filter_id` - (optional) Filter name. Can be:
  - `ip_address` - IP address from which the request is sent
  - `pool` - ID of the [application](https://docs.wallarm.com/user-guides/settings/applications/) that receives the request or in which an event is detected.
  - `attack_type` - type of the attack detected in the request or a type of vulnerability the request is directed to.
  - `domain` - Domain that receives the request or in which an event is detected.
  - `target` - Application architecture part that the attack is directed at or in which the incident is detected. Can be:
    * `Server`
    * `Client`
    * `Database`
* `response_status` - Integer response code returned to the request.
* `hint_tag` - Arbitrary tag of any request tuned in by a rule.
* `pii` - Array of fields that are considered sensitive (PII).
* `change_type` - Array of change types for API structure changed trigger. Can be:
    * `added`
    * `changed`
    * `excluded`
* `deviation_type` - Array of Rogue API types for Rogue api detected trigger. Values in the array can be:
  * `shadow`
  * `orphan`
  * `zombie`
* `operator` - (optional) Operator to compare the specified filter value and a real value. Can be:
    * `eq` - Equal
    * `ne` - Not equal
* `value` - (optional) Filter value.

Example:

```hcl
  # ... omitted

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [1]
  }

  # ... omitted
```

## Threshold

`threshold` argument shares the available conditions which can be applied.  It must **NOT** be specified when the `user_created` template is used. The conditions are:
  - `period` - The period of time to count (in seconds by default).
  - `time_format` - Time units for period. Can be:
    * `Seconds` - By default, to measure in seconds.
    * `Minutes` - To measure period in minutes.
  - `count` - The number of such events.
  - `operator` - (optional) The comparison operator. Valid values:
    * `gt` - Greater than

Example:

```hcl
  # ... omitted

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  threshold = {
    operator = "gt"
    count = 5
    period = 3600
  }

  # ... omitted
```

`actions` argument shares the available conditions which can be applied. The conditions are:
  - `action_id` - (**required**) the type of action when triggered.
    * `send_notification` - send notification to existing integration resource.
    * `block_ips` - block the IP addresses from which the requests originated.
    * `mark_as_brute` - to mark as bruteforce.
    * `add_to_graylist` - to add to graylist.
    * `group_attack_by_ip` - to group an attack by the same IP.
  - `integration_id` - the identificator of the existing integration.
  - `lock_time` - The time for which to block IP addresses in case of usage `block_ips`.
  - `lock_time_format` - Time units to setup lock time. In seconds by default. Can be: 
    * `Minutes` - Time in minutes (e.g. `60` is to block for 60 minutes).
    * `Hours` - Time in hours (e.g. `5` is to block for 5 hours).
    * `Days` - Time in days (e.g. `7` is to block for 7 days).
    * `Weeks` - Time in weeks (e.g. `4` is to block for 4 weeks).
    * `Months` - Time in weeks (e.g. `12` is to block for 12 months).

Example:

```hcl
  # ... other configuration

  actions {
    action_id = "send_notification"
    integration_id = [123]
  }

  actions {
    action_id = "block_ips"
    lock_time = 10000
  }

  # ... skipped
```

## Attributes Reference

* `trigger_id` - ID of the created trigger.

[1]: https://docs.wallarm.com/user-guides/triggers/triggers/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
