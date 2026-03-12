resource "wallarm_integration_sumologic" "sumologic_integration" {
  name = "Terraform SumoLogic Integration"
  sumologic_url = "http://sumologic.com/changed/once/again"

  event {
    event_type   = "siem"
    active       = true
    with_headers = true
  }
  event {
    event_type = "rules_and_triggers"
    active = true
  }
  event {
    event_type = "number_of_requests_per_hour"
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