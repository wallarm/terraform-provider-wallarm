provider "wallarm" {
  api_uuid = "${var.api_uuid}"
  api_secret = "${var.api_secret}"
  api_host = "${var.api_host}"
  api_client_logging = true
  client_id = 6039
}

# resource "wallarm_user" "user" {
#   email = "testuser6039@wallarm.com"
#   # password = "IGIHKJHJIndjksj54_2."
#   # phone = "+7 900 123 45 67"
#   permissions = "analyst"
#   realname = "Test User"
# }

# output "password" {
#   value = wallarm_user.user.onetime_password
# }

# resource "wallarm_blacklist" "blacklist" {
#   ip_range = ["1.1.1.1/32"]
#   application = [1]
#   reason = "TEST BLACKLIST"
#   time_format = "Minutes"
#   time = 60
# }

# resource "wallarm_rule_vpatch" "default" {
#   attack_type =  ["sqli"]
#   point = [["get_all"]]
# }

# resource "wallarm_rule_vpatch" "vpatch" {
#   attack_type =  ["redir"]
#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   action {
#     type = "equal"
#     value = "api"
#     point = {
#       path = 0
#     }
#   }
#   action {
#     type = "regex"
#     value = "logon"
#     point = {
#       path = 1
#     }
#   }
#   action {
#     type = "equal"
#     point = {
#       method = "POST"
#     }
#   }
#   action {
#     point = {
#       instance = "1"
#     }
#   }
#   action {
#     type = "regex"
#     point = {
#       scheme = "https"
#     }
#   }
#   point = [["post"],["xml"],["hash","user"]]
# }

# resource "wallarm_rule_vpatch" "splunk" {
#   attack_type =  ["sqli", "nosqli"]
#   action {
#     type = "iequal"
#     value = "splunk.wallarm-demo.com:88"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["get_all"]]
# }

# resource "wallarm_rule_vpatch" "tiredful_api" {
#   attack_type =  ["any"]
#   action {
#     point = {
#       instance = "9"
#     }
#   }
#   action {
#     type = "absent"
#     point = {
#       path = 0
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       action_name = "formmail"
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       action_ext = "cgi"
#     }
#   }

#   point = [["uri"]]
# }

# resource "wallarm_rule_vpatch" "env_sample" {
#   attack_type =  ["any"]

#   action {
#     type = "equal"
#     point = {
#       action_name = ".env.sample"
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       action_ext = "php"
#     }
#   }

#   point = [["uri"]]
# }


# resource "wallarm_rule_mode" "wp_mode" {
#   mode =  "block"

#   action {
#     point = {
#       instance = "6"
#     }
#   }

#   action {
#     type = "iequal"
#     value = "monitor"
#     point = {
#       path = 0
#     }
#   }
# }

# resource "wallarm_rule_mode" "tiredful_api_mode" {
#   mode =  "monitoring"

#   action {
#     point = {
#       instance = "9"
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       action_name = "formmail"
#     }
#   }
# }


# resource "wallarm_rule_mode" "ad_mode" {
#   mode =  "default"

#   action {
#     type = "equal"
#     value = "api"
#     point = {
#       path = 0
#     }
#   }

#   action {
#     type = "equal"
#     value = "active-directory"
#     point = {
#       path = 1
#     }
#   }
# }

# resource "wallarm_rule_mode" "dvwa_mode" {
#   mode =  "block"

#   action {
#     type = "equal"
#     value = "dvwa.wallarm-demo.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       method = "GET"
#     }
#   }
# }

# resource "wallarm_rule_masking" "dvwa_sensitive" {

#   action {
#     point = {
#       instance = "5"
#     }
#   }

#   point = [["header", "X-KEY"]]
# }

# resource "wallarm_rule_masking" "masking" {

#   action {
#     type = "absent"
#     point = {
#       path = 0
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       action_name = "masking"
#     }
#   }

#   action {
#     type = "absent"
#     point = {
#       action_ext = ""
#     }
#   }
#   point = [["header", "X-KEY"]]
# }

# resource "wallarm_rule_masking" "masking_json" {

#   action {
#     type = "absent"
#     point = {
#       path = 0
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       action_name = "masking"
#     }
#   }

#   action {
#     type = "absent"
#     point = {
#       action_ext = ""
#     }
#   }
#   point = [["post"],["json_doc"],["hash", "field"]]
# }

# resource "wallarm_rule_regex" "regex" {
#   regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
#   experimental = true
#   attack_type =  "redir"

#   action {
#     type = "iequal"
#     value = "tiredful-api.wallarm-demo.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["header", "X-AUTHENTICATION"]]
# }

# resource "wallarm_rule_ignore_regex" "ingore_regex" {
#   regex_id = wallarm_rule_regex.regex.regex_id
#   action {
#   }
#   point = [["header", "X-LOGIN"]]

#   depends_on = [wallarm_rule_regex.regex]
# }

# resource "wallarm_waf_mode" "monitoring" {
#   mode = "block"
# }

# variable "node_names" {
#   description = "Create Node names"
#   type        = list(string)
#   default     = ["prod", "stage", "dev"]
# }

# resource "wallarm_node" "cloud_node" {
#   count = 3
#   hostname = "tf-${var.node_names[count.index]}"
# }


# data "wallarm_node" "waf" {

#   filter {
#     type = "cloud_node"
#     # hostname = "4f5f7b48bf13"
#     # uuid = "b161e6f9-33d2-491e-a584-513522d312db"
#   }
# }

# # output "waf_nodes" {
# #   value = data.wallarm_node.waf.nodes
# # }


# resource "wallarm_scanner" "scan" {
#     element = ["1.1.1.0", "example.com"]
#     disabled = false
# }

# output "scan_id" {
#   value = wallarm_scanner.scan.resource_id
# }

# resource "wallarm_application" "tf_app" {
#   name = "New Terraform Application"
#   app_id = 43
# }

# resource "wallarm_integration_email" "email_integration" {
#   name = "New Terraform Integration"
#   active = false
#   emails = ["kokk@wallarm.com", "lal@wallarm.com"]
#   event {
#     event_type = "report_monthly"
#     active = true
#   }
  
#   event {
#     event_type = "vuln"
#     active = true
#   }
# }

# # resource "wallarm_trigger" "attack_trigger" {
# #   name = "New Terraform Trigger"
# #   enabled = false
# #   template_id = "attacks_exceeded"

# #   filters {
# #     filter_id = "ip_address"
# #     operator = "eq"
# #     value = ["2.2.2.2"]
# #   }

# #   filters {
# #     filter_id = "pool"
# #     operator = "eq"
# #     value = [wallarm_application.tf_app.app_id]
# #   }

# #   threshold = {
# #     period = 86400
# #     operator = "gt"
# #     count = 10000
# #   }

# #   actions {
# #     action_id = "send_notification"
# #     integration_id = [wallarm_integration_email.email_integration.integration_id]
# #   }
# #   depends_on = [
# #     "wallarm_integration_email.email_integration",
# #     "wallarm_application.tf_app",
# #   ]
# # }


# resource "wallarm_trigger" "user_trigger" {
#   client_id = 6039
#   name = "The Updated Terraform Trigger User"
#   comment = "This is the updated description set by Terraform"
#   enabled = true
#   template_id = "user_created"

#   actions {
#     action_id = "send_notification"
#     integration_id = [597]
#   }
# }

# resource "wallarm_integration_opsgenie" "opsgenie_integration" {
#   name = "New Terraform OpsGenie Integration"
#   api_token = "SuPER_SecREt-TokEN-APIfor_OPSGenie"
#   active = true

#   event {
#     event_type = "hit"
#     active = true
#   }
  
#   event {
#     event_type = "vuln"
#     active = true
#   }
# }

# resource "wallarm_integration_pagerduty" "pagerduty_integration" {
#   name = "New Terraform PagerDuty Integration"
#   integration_key = "48c8f1999cbf4b91a3dbd0fac79bfc6b"

#   event {
#     event_type = "hit"
#     active = true
#   }

#   event {
#     event_type = "scope"
#     active = true
#   }

#   event {
#     event_type = "system"
#     active = true
#   }
  
#   event {
#     event_type = "vuln"
#     active = true
#   }
# }

# resource "wallarm_integration_sumologic" "sumologic_integration" {
#   name = "New Terraform SumoLogic Integration"
#   sumologic_url = "http://sumologic.com/changed/once/again"

#   event {
#     event_type = "hit"
#     active = true
#   }

#   event {
#     event_type = "scope"
#     active = true
#   }

#   event {
#     event_type = "system"
#     active = false
#   }
  
#   event {
#     event_type = "vuln"
#     active = false
#   }
# }

# resource "wallarm_trigger" "vector_trigger" {
#   name = "New Terraform Trigger"
#   enabled = true
#   template_id = "vector_attack"

#   filters {
#     filter_id = "ip_address"
#     operator = "eq"
#     value = ["2.2.2.2"]
#   }

#   threshold = {
#     operator = "gt"
#     count = 5
#     period = 3600
#   }

#   actions {
#     action_id = "block_ips"
#     lock_time = 10000
#   }

# }

# resource "wallarm_integration_insightconnect" "insight_integration" {
#   name = "New Terraform InsightConnect Integration"
#   api_url = "https://us.api.insight.rapid7.com/connect/v1/workflows/d1763a97-e41a-1020-a651-26c1437657081/events/execute"
#   api_token = "c038033e-550a-0390-aa00-a102e5b356a7"

#   event {
#     event_type = "hit"
#     active = true
#   }

#   event {
#     event_type = "scope"
#     active = true
#   }

#   event {
#     event_type = "system"
#     active = true
#   }
  
#   event {
#     event_type = "vuln"
#     active = true
#   }
# }

# resource "wallarm_integration_splunk" "splunk_integration" {
#   name = "New Terraform Splunk Integration"
#   api_url = "http://splunk.wallarm.com"
#   api_token = "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"

#   event {
#     event_type = "hit"
#     active = true
#   }

#   event {
#     event_type = "scope"
#     active = true
#   }

#   event {
#     event_type = "system"
#     active = true
#   }
  
#   event {
#     event_type = "vuln"
#     active = true
#   }
# }

#  resource "wallarm_integration_webhook" "wh_integration" {
#   name = "New Terraform WebHook Integration"
#   webhook_url = "https://webhook.wallarm.com"
#   http_method = "POST"

#   event {
#     event_type = "hit"
#     active = true
#   }

#   event {
#     event_type = "scope"
#     active = true
#   }

#   event {
#     event_type = "system"
#     active = true
#   }
  
#   event {
#     event_type = "vuln"
#     active = true
#   }

#   headers = {
#     HOST = "localhost"
#     Content-Type = "application/json"
#   }

# }

# # data "wallarm_vuln" "vulns" {

# #   filter {
# #     status = "open"
# #     limit = 1000
# #   }
# # }

# # output "vulns" {
# #   value = data.wallarm_vuln.vulns.vuln
# # }


# resource "wallarm_rule_attack_rechecker_rewrite" "default_rewrite" {
#   rules =  "my.awesome-application.com"
#   point = [["header", "HOST"]]
# }

# resource "wallarm_rule_attack_rechecker" "enable_rechecker" {
#   enabled =  false

#   action {
#     point = {
#       instance = "6"
#     }
#   }

# }

# resource "wallarm_rule_set_response_header" "resp_headers" {
#   mode = "replace"

#   action {
#     point = {
#       instance = "6"
#     }
#   }

#   headers = {
#     Server = "Wallarm"
#     Blocked = "Yes, you are"
#   }

# }


# resource "wallarm_node" "cloud_node" {
#   count = 1
#   hostname = "tf-test"
# }