resource "wallarm_integration_opsgenie" "opsgenie_integration" {
  name = "New Terraform OpsGenie Integration"
  api_url = "https://api.opsgenie.com/v2/alerts"
  api_token = "eb7ddfc33acaaacaacaca55a39834aaa"
  active = true

  event {
    event_type = "rules_and_triggers"
    active = false
  }
  event {
    event_type = "security_issue_critical"
    active = false
  }
  event {
    event_type = "security_issue_high"
    active = false
  }
  event {
    event_type = "security_issue_medium"
    active = false
  }
  event {
    event_type = "security_issue_low"
    active = false
  }
  event {
    event_type = "security_issue_info"
    active = false
  }
  event {
    event_type = "system"
    active = false
  }
}