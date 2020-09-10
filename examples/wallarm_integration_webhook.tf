resource "wallarm_integration_webhook" "wh_integration" {
  name = "New Terraform WebHook Integration"
  webhook_url = "https://webhook.example.com"
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
    event_type = "vuln"
    active = true
  }

  headers = {
    HOST = "localhost"
    Content-Type = "application/json"
  }

}