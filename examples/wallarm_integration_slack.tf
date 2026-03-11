resource "wallarm_integration_slack" "slack_integration" {
  name = "Terraform Slack Integration X"
  webhook_url = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
  active = true

  event {
    event_type = "rules_and_triggers"
    active = true
  }
  event {
    event_type = "security_issue_critical"
    active = true
  }
  event {
    event_type = "security_issue_high"
    active = true
  }
  event {
    event_type = "security_issue_medium"
    active = true
  }
  event {
    event_type = "security_issue_low"
    active = true
  }
  event {
    event_type = "security_issue_info"
    active = true
  }
  event {
    event_type = "system"
    active = true
  }
}