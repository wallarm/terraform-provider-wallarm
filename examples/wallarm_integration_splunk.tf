resource "wallarm_integration_splunk" "splunk_integration" {
  name = "Terraform Splunk Integration X"
  api_url = "https://httpbin.org:443"
  api_token = "B5A79AAD-D822-46CC-80D1-819F80D7BFB0"
  active = true

  event {
    event_type   = "siem"
    active       = false
    with_headers = false
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