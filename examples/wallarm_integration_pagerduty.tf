resource "wallarm_integration_pagerduty" "pagerduty_integration" {
  name = "Terraform Pagerduty Integration"
  integration_key = "eb7ddfc33acaaacaacaca55a39834dad"
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
    active = true
  }
}