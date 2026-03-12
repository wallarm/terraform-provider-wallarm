resource "wallarm_integration_telegram" "telegram_integration" {
  name = "Terraform Telegram Integration"
  telegram_username = "WallarmIntegrationTest"
  chat_data = "+y86q0LOQ4QG3hK9QgVDfw=="
  active = true

  event {
    event_type = "rules_and_triggers"
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
  event {
		event_type = "report_daily"
		active = false
	}
	event {
		event_type = "report_weekly"
		active = false
	}
	event {
		event_type = "report_monthly"
		active = false
	}
}
