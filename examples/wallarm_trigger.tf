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
    event_type = "vuln"
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