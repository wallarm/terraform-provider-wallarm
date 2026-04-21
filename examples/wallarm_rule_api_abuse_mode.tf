# Enable API Abuse Prevention for a specific application (instance).
resource "wallarm_rule_api_abuse_mode" "tiredful_api_abuse_mode" {
  mode = "enabled"

  action {
    point = {
      instance = 9
    }
  }
}

# Enable API Abuse Prevention for all traffic to a specific host.
resource "wallarm_rule_api_abuse_mode" "dvwa_abuse_mode" {
  mode = "enabled"

  action {
    type  = "iequal"
    value = "dvwa.wallarm-demo.com"
    point = {
      header = "HOST"
    }
  }
}

# Allowlist trusted Pinterest crawler traffic under /api/v{N} routes by
# disabling API Abuse Prevention for requests that match the crawler's
# User-Agent and path shape.
resource "wallarm_rule_api_abuse_mode" "pinterest" {
  mode    = "disabled"
  title   = "Allow Pinterest"
  comment = "Allow Pinterest through API Abuse Prevention"

  action {
    type  = "regex"
    value = ".*(Pinterest|Pinterestbot)/(0.2|1.0);?\\s[(]?[+]https?://www[.]pinterest[.]com/bot[.]html[)].*"
    point = {
      header = "USER-AGENT"
    }
  }

  action {
    type  = "equal"
    value = "api"
    point = {
      path = 0
    }
  }

  action {
    type  = "regex"
    value = "v\\d"
    point = {
      path = 1
    }
  }

  action {
    type = "absent"
    point = {
      action_ext = ""
    }
  }
}
