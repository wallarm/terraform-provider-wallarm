resource "wallarm_integration_teams" "teams_integration" {
  name = "Terraform MS Teams Integration"
  webhook_url = "https://xxxxx.webhook.office.com/xxxxxxxxx"
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