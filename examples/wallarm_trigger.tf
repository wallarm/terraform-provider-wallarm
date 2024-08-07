resource "wallarm_application" "tf_app" {
  name = "New Terraform Application"
  app_id = 43
}

resource "wallarm_integration_email" "email_integration" {
  name = "New Terraform Integration"
  emails = ["test1@wallarm.com", "test2@wallarm.com"]
  event {
    event_type = "report_monthly"
    active = true
  }
  
  event {
    event_type = "vuln_low"
    active = true
  }
}

resource "wallarm_trigger" "attack_trigger" {
  name = "New Terraform Trigger"
  enabled = false
  template_id = "attacks_exceeded"

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_email.email_integration.integration_id]
  }
  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
  ]
}

resource "wallarm_trigger" "user_trigger" {
  name = "New Terraform Trigger Telegram"
  comment = "This is a description set by Terraform"
  enabled = true
  template_id = "user_created"

  actions {
    action_id = "send_notification"
    integration_id = [521]
  }
}

resource "wallarm_trigger" "vector_trigger" {
  name = "New Terraform Trigger"
  enabled = true
  template_id = "vector_attack"

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  threshold = {
    operator = "gt"
    count = 5
    period = 3600
  }

  actions {
    action_id = "block_ips"
    lock_time = 10000
  }
}

resource "wallarm_rule_bruteforce_counter" "brute_counter" {
    action {
		type = "iequal"
		value = "example.com"
		point = {
			header = "HOST"
		}
	}

	action {
		type = "iequal"
		value = "foobar"
		point = {
			path = 0
		}
	}
}

resource "wallarm_trigger" "brute_trigger" {
	template_id = "bruteforce_started"

	filters {
		filter_id = "hint_tag"
		operator = "eq"
		value = [wallarm_rule_bruteforce_counter.brute_counter.counter]
	}

	actions {
		action_id = "mark_as_brute"
	}

	actions {
		action_id = "block_ips"
		lock_time = 2592000
	}

	threshold = {
		period = 30
		operator = "gt"
		count = 30
	}
}

# BOLA
resource "wallarm_rule_bola_counter" "bola_counter" {
    action {
		type = "iequal"
		value = "example.com"
		point = {
			header = "HOST"
		}
	}

	action {
		type = "iequal"
		value = "foobar"
		point = {
			path = 0
		}
	}
}

resource "wallarm_trigger" "bola_trigger" {
	template_id = "bola_search_started"

	filters {
		filter_id = "hint_tag"
		operator = "eq"
		value = [wallarm_rule_bola_counter.bola_counter.counter]
	}

	actions {
		action_id = "mark_as_brute"
	}

	actions {
		action_id = "block_ips"
		lock_time = 2592000
	}

	threshold = {
		period = 30
		operator = "gt"
		count = 30
	}
}

# Number of malicious payloads
resource "wallarm_trigger" "vector_attack_trigger" {
  name = "New Terraform Vector Attack Trigger"
  enabled = false
  template_id = "vector_attack"

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  filters {
    filter_id = "attack_type"
    operator = "eq"
    value = ["sqli"]
  }

  filters {
    filter_id = "domain"
    operator = "eq"
    value = ["ex.com"]
  }

  filters {
    filter_id = "response_status"
    operator = "eq"
    value = ["5xx"]
  }

  filters {
    filter_id = "target"
    operator = "eq"
    value = ["client"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  actions {
		action_id = "block_ips"
		lock_time = 2592000
	}

  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
  ]
}

# Number of attacks
resource "wallarm_trigger" "attacks_exceeded_trigger" {
  name = "New Terraform Attacks Exceeded Trigger"
  enabled = false
  template_id = "attacks_exceeded"

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  filters {
    filter_id = "attack_type"
    operator = "eq"
    value = ["sqli"]
  }

  filters {
    filter_id = "domain"
    operator = "eq"
    value = ["ex.com"]
  }

  filters {
    filter_id = "response_status"
    operator = "eq"
    value = ["5xx"]
  }

  filters {
    filter_id = "target"
    operator = "eq"
    value = ["client"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_email.email_integration.integration_id]
  }

  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
  ]
}

# Number of hits
resource "wallarm_trigger" "hits_exceeded_trigger" {
  name = "New Terraform Hits Exceeded Trigger"
  enabled = false
  template_id = "hits_exceeded"

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  filters {
    filter_id = "attack_type"
    operator = "eq"
    value = ["sqli"]
  }

  filters {
    filter_id = "domain"
    operator = "eq"
    value = ["ex.com"]
  }

  filters {
    filter_id = "response_status"
    operator = "eq"
    value = ["5xx"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_email.email_integration.integration_id]
  }

  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
  ]
}

# Number of incidents
resource "wallarm_trigger" "incidents_exceeded_trigger" {
  name = "New Terraform Incidents Exceeded Trigger"
  enabled = false
  template_id = "incidents_exceeded"

  filters {
    filter_id = "ip_address"
    operator = "eq"
    value = ["2.2.2.2"]
  }

  filters {
    filter_id = "attack_type"
    operator = "eq"
    value = ["sqli"]
  }

  filters {
    filter_id = "domain"
    operator = "eq"
    value = ["ex.com"]
  }

  filters {
    filter_id = "response_status"
    operator = "eq"
    value = ["5xx"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_email.email_integration.integration_id]
  }

  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
  ]
}

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

# Denylisted IP
resource "wallarm_trigger" "blacklist_ip_added_trigger" {
  name = "New Terraform Blacklist IP Added Trigger"
  enabled = false
  template_id = "blacklist_ip_added"

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_webhook.wh_integration.integration_id]
  }

  depends_on = [
    "wallarm_integration_webhook.wh_integration",
    "wallarm_application.tf_app",
  ]
}

# Changes in API
resource "wallarm_trigger" "api_structure_changed_trigger" {
  name = "New Terraform Api Structure Changed Trigger"
  enabled = false
  template_id = "api_structure_changed"

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  filters {
    filter_id = "domain"
    operator = "eq"
    value = ["ex.com"]
  }

  filters {
    filter_id = "change_type"
    operator = "eq"
    value = ["added", "changed", "excluded"]
  }

  filters {
    filter_id = "pii"
    operator = "eq"
    value = ["password",
      "secret",
      "credit_card_number",
      "CVV",
      "IBAN",
      "driver_license_number",
      "email",
      "location",
      "login",
      "passport_number",
      "personal_name",
      "phone",
      "SSN",
      "IP",
      "MAC"]
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_email.email_integration.integration_id]
  }

  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
  ]
}

# Hits from the same IP
resource "wallarm_trigger" "attack_ip_grouping_trigger" {
  name = "New Terraform Attack IP grouping Trigger"
  enabled = false
  template_id = "attack_ip_grouping"

  threshold = {
    period = 86400
    operator = "gt"
    count = 10000
  }

  actions {
    action_id = "group_attack_by_ip"
  }

  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
  ]
}

resource "wallarm_api_spec" "api_spec" {
  client_id          = 106662
  title              = "Example API Spec"
  description        = "This is an example API specification created by Terraform."
  file_remote_url    = "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/main/examples/v3.0/api-with-examples.yaml"
  regular_file_update = true
  api_detection      = true
  domains = ["ex.com"]
  instances = [1]
}

# Rogue API detected
resource "wallarm_trigger" "rogue_api_detected_trigger" {
  name = "New Terraform Rogue API detected Trigger"
  enabled = false
  template_id = "rogue_api_detected"

  filters {
    filter_id = "domain"
    operator = "eq"
    value = ["ex.com"]
  }

  filters {
    filter_id = "pool"
    operator = "eq"
    value = [wallarm_application.tf_app.app_id]
  }

  filters {
    filter_id = "deviation_type"
    operator = "eq"
    value = ["shadow", "orphan", "zombie"]
  }

  filters {
    filter_id = "api_spec_ids"
    operator = "eq"
    value = [wallarm_api_spec.api_spec.api_spec_id]
  }

  actions {
    action_id = "send_notification"
    integration_id = [wallarm_integration_email.email_integration.integration_id]
  }

  depends_on = [
    "wallarm_integration_email.email_integration",
    "wallarm_application.tf_app",
    "wallarm_api_spec.api_spec"
  ]
}
