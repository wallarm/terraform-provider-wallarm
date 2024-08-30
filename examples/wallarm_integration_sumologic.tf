resource "wallarm_integration_sumologic" "sumologic_integration" {
  name = "New Terraform SumoLogic Integration"
  sumologic_url = "http://sumologic.com/changed/once/again"

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
    active = false
  }
  
  event {
    event_type = "vuln_low"
    active = false
  }
}
