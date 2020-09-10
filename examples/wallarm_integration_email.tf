resource "wallarm_integration_email" "email_integration" {
  name = "New Terraform Integration"
  active = false
  emails = ["test@wallarm.com", "test2@wallarm.com"]
  event {
    event_type = "report_monthly"
    active = true
  }
  
  event {
    event_type = "vuln"
    active = true
  }
}