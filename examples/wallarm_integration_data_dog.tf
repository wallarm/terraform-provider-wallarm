resource "wallarm_integration_data_dog" "data_dog_integration" {
  name = "New Terraform DataDog Integration"
  region = "US5"
  token = "eb7ddfc33acaaacaacaca55a39834dad"
  active = true

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