resource "wallarm_integration_insightconnect" "insight_integration" {
  name = "New Terraform InsightConnect Integration"
  api_url = "https://us.api.insight.rapid7.com/connect/v1/workflows/d1763a97-e41b-1020-a651-26c1427657081/events/execute"
  api_token = "c038033e-550a-0260-aa00-a102e5b356a7"

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
