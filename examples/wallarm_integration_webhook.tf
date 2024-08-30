resource "wallarm_integration_webhook" "wh_integration" {
  name = "New Terraform WebHook Integration"
  webhook_url = "https://example.com/webhook"
  http_method = "POST"

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

  headers = {
    HOST = "localhost"
    Content-Type = "application/json"
  }

}

