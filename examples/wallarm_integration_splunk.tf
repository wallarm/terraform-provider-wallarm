resource "wallarm_integration_splunk" "splunk_integration" {
  name = "New Terraform Splunk Integration"
  api_url = "https://example.com:8088"
  api_token = "b1e2d6dc-e4b5-400d-9d4e-270c39d5daa2"

  event {
    event_type = "hit"
    active = true
  }

  event {
    event_type = "scope"
    active = true
  }

  event {
    event_type = "system"
    active = true
  }
  
  event {
    event_type = "vuln_low"
    active = true
  }
}
