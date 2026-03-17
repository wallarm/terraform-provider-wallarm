# # import {
# #   to = wallarm_tenant.my_tenant
# #   id = "32324"
# # }

# # resource "wallarm_tenant" "my_tenant" {
# #   client_id = 1324
# #   name = "tenant"
# # }

# resource "wallarm_user" "my_user" {
#   email = "testuser6088@wallarm.com"
#   password = "IfIHKJdJIndjksj54_2r"
#   permissions = "admin"
#   realname = "Test 1"
# }

# resource "wallarm_rules_settings" "rules_settings" {
# #   client_id = 3482
#   min_lom_format = null # Default value - recommended
#   max_lom_format = null # Default value - recommended
#   max_lom_size = 100000000 # 100Mb limit for large rulesets
#   lom_disabled = false
#   lom_compilation_delay = 10 # Recommended to avoid operations of resources blocking
#   rules_snapshot_enabled = true
#   rules_snapshot_max_count = 5
# }

# # import {
# #   to = wallarm_application.default_app
# #   id = "3482/-1"
# # }
# # resource "wallarm_application" "default_app" {
# #   name = "Default Application"
# #   app_id = -1
# # }

# resource "wallarm_application" "one_x_4_app" {
#   name = "Application 1 1 1 1"
#   app_id = 1111
# }

# resource "wallarm_global_mode" "global_block" {
#   filtration_mode = "default" # Global filtration mode
#   rechecker_mode = "off" # Threat Replay
#   overlimit_time = 1000 # Default recommended
#   overlimit_mode = "monitoring" # Default
# }

# data wallarm_security_issues security_issues {
#     client_id = 3482
# }

# data wallarm_hits hits_test-1 {
#     client_id = 3482
#     request_id = "0a9be49e7e8e810ee7a6ef5f26e15534"
# }

# resource "wallarm_denylist" "denylist_ip_random" {
#   ip_range = ["73.247.168.209","7.128.126.121","254.238.236.138","162.67.180.20","219.200.243.97","166.160.49.3","244.183.231.140","133.64.237.72","222.184.143.140","208.2.237.220","190.142.117.33","119.23.239.169","131.21.138.217","186.166.11.221","33.179.38.106","20.254.121.242","12.236.147.130","151.91.112.83","29.182.132.175","158.253.163.113","227.199.62.174","118.181.111.119","102.230.86.175","69.165.200.153","150.48.46.107","234.57.250.122","142.187.147.46","229.233.53.220","57.150.78.227","237.29.243.63","43.159.200.161","253.161.253.13","184.19.142.129","172.3.211.37","57.53.186.12","238.19.119.75","47.219.10.66","97.223.191.237","87.60.178.151","218.181.152.34","39.136.181.251","55.126.68.138","195.213.198.134","92.161.78.139","165.222.22.48","84.58.203.172","42.152.85.162","99.196.125.87","14.13.159.201","149.150.17.250","63.15.215.217","100.136.80.135","173.140.151.234","160.192.111.209","7.145.6.235","13.198.202.55","72.243.84.92","173.242.194.87","223.255.142.181","34.86.204.208","207.138.185.145","30.70.102.107","173.0.141.242","125.171.43.173","80.75.104.218","12.39.27.107","81.81.56.83","193.173.52.248","243.228.31.151","131.96.91.68","111.130.125.25","87.204.125.204","48.64.185.143","223.254.82.149","185.38.37.109","131.58.232.200","190.234.76.89","140.1.80.159","176.170.55.243","119.115.129.42","243.35.194.175","122.117.186.56","196.225.170.31","161.223.229.80","224.95.232.95","178.41.180.201","16.169.84.143","96.242.24.66","70.214.227.43","59.19.191.196","105.202.218.144","124.219.80.42","40.164.248.103","34.49.37.206","88.107.155.35","158.185.93.217","227.173.201.40"]
#   application = [1]
#   reason = "TEST DENYLIST"
#   time_format = "rfc3339"
#   time = "2026-06-10T23:33:00+07:00"
# }

# resource "wallarm_denylist" "denylist_countries" {
#   country = ["AF","AX","AL","DZ","AS","AD","AO","AI","AG","AR","AM","AW","AU","AT","AZ","BS","BH","BD","BB","BY","BE","BZ","BJ","BM","BT","BO","BQ","BA","BW","BV","BR","IO","BN","BG","BF","BI","CV","KH","CM","CA","KY","CF","TD","CL","CN","CX","CC","CO","KM","CG","CD","CK","CR","CI","HR","CU","CW","CY","CZ","DK","DJ","DM","DO","DNO","EC","EG","SV","GQ","ER","EE","ET","FK","FO","FJ","FI","FR","GF","PF","TF","GA","GM","GE","DE","GH","GI","GB","GR","GL","GD","GP","GU","GT","GG","GN","GW","GY","HT","HM","VA","HN","HK","HU","IS","IN","ID","IR","IQ","IE","IM","IL","IT","JM","JP","JE","JO","KZ","KE","KI","KP","KR","KW","KG","LA","LV","LB","LS","LR","LY","LI","LT","LHO","LU","MO","MK","MG","MW","MY","MV","ML","MT","MH","MQ","MR","MU","YT","MX","FM","MD","MC","MN", "ME","MS","MA","MZ","MM","NA","NR","NP","NL","NC","NZ","NI","NE","NG","NU","NF","MP","NO","OM","PK","PW","PS","PA","PG","PY","PE","PH","PN","PL","PT","PR","QA","ROC","RE","RO","RU","RW","BL","SH","KN","LC","MF","PM","VC","WS","SM","ST","SA","SN","RS","SC","SL","SG","SX","SK","SI","SB","SO","ZA","GS","SS","ES","LK","SD","SR","SJ","SZ","SE","CH","SY","TW","TJ","TZ","TH","TL","TG","TK","TO","TT","TN","TR","TM","TC","TV","UG","UA","AE","US","UY","UZ","VU","VE","VN","VG","VI","WF","EH","YE","ZM","ZW"]
#   application = [0] # All Applications
#   reason = "TEST DENYLIST"
#   time_format = "forever"
# }

# resource "wallarm_allowlist" "allowlist_ip_subnets" {
#   ip_range = ["172.16.16.0/16","10.0.0.0/8","192.168.10.1/24","8.8.8.8","1.1.1.1"]
#   application = [-1] # Default
#   reason = "TEST ALLOWLIST"
#   time_format = "rfc3339"
#   time = "2029-06-12T00:00:00Z"
# }

# resource "wallarm_denylist" "denylist_minutes_tor" {
#   application = [0] # All Applications
#   proxy_type = ["TOR"]
#   reason = "TEST DENYLIST TOR"
#   time_format = "Minutes"
#   time = 60
# }

# resource "wallarm_graylist" "graylist_datacenters" {
#   application = [100]
#   datacenter = ["aws", "docean"]
#   reason = "TEST GRAYLIST DATACENTER"
#   time_format = "Minutes"
#   time = 30
# }

# resource "wallarm_integration_email" "email_integration_test" {
#   name = "New Terraform Integration"
#   active = false
#   emails = ["test@wallarm.com", "test2@wallarm.com"]

#   event {
#     event_type = "system"
#     active = true
#   }
#   event {
#     event_type = "aasm_report"
#     active = true
#   }
#   event {
#     event_type = "api_discovery_hourly_changes_report"
#     active = true
#   }
#   event {
#     event_type = "api_discovery_daily_changes_report"
#     active = true
#   }
#   event {
#     event_type = "report_daily"
#     active = true
#   }
#   event {
#     event_type = "report_weekly"
#     active = true
#   }
#   event {
#     event_type = "report_monthly"
#     active = true
#   }
# }

# resource "wallarm_integration_data_dog" "data_dog_integration" {
#   name = "New Terraform DataDog Integration"
#   region = "US5"
#   token = "eb7ddfc33acaaacaacaca55a39834dad"
#   active = true

#   event {
#     event_type   = "siem"
#     active       = true
#     with_headers = true
#   }
#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "number_of_requests_per_hour"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_integration_insightconnect" "insight_integration" {
#   name = "New Terraform InsightConnect Integration"
#   api_url = "https://us.api.insight.rapid7.com/connect/v1/workflows/d1763a97-e41b-1020-a651-26c1427657081/events/execute"
#   api_token = "c038033e-550a-0260-aa00-a102e5b356a7"

#   event {
#     event_type   = "siem"
#     active       = true
#     with_headers = true
#   }
#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "number_of_requests_per_hour"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_integration_opsgenie" "opsgenie_integration" {
#   name = "New Terraform OpsGenie Integration"
#   api_url = "https://api.opsgenie.com/v2/alerts"
#   api_token = "eb7ddfc33acaaacaacaca55a39834aaa"
#   active = true

#   event {
#     event_type = "rules_and_triggers"
#     active = false
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = false
#   }
#   event {
#     event_type = "security_issue_high"
#     active = false
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = false
#   }
#   event {
#     event_type = "security_issue_low"
#     active = false
#   }
#   event {
#     event_type = "security_issue_info"
#     active = false
#   }
#   event {
#     event_type = "system"
#     active = false
#   }
# }

# resource "wallarm_integration_pagerduty" "pagerduty_integration" {
#   name = "Terraform Pagerduty Integration"
#   integration_key = "eb7ddfc33acaaacaacaca55a39834dad"
#   active = true

#   event {
#     event_type = "rules_and_triggers"
#     active = false
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = false
#   }
#   event {
#     event_type = "security_issue_high"
#     active = false
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = false
#   }
#   event {
#     event_type = "security_issue_low"
#     active = false
#   }
#   event {
#     event_type = "security_issue_info"
#     active = false
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_integration_splunk" "splunk_integration" {
#   name = "Terraform Splunk Integration X"
#   api_url = "https://httpbin.org:443"
#   api_token = "B5A79AAD-D822-46CC-80D1-819F80D7BFB0"
#   active = true

#   event {
#     event_type   = "siem"
#     active       = false
#     with_headers = false
#   }
#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "number_of_requests_per_hour"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_integration_webhook" "webhook_integration" {
#   name = "Terraform Webhook Integration"
#   active = true
#   webhook_url = "https://httpbin.org:443"
#   http_method =  "POST"
#   format = "json"
#   headers = {
#     "INTEGRATION" = "B5A79AAD"
#     "Content-Type" = "application/json"
#   }
#   ca_file = "-----BEGIN CERTIFICATE-----\nMIIDQTCCAimgAwIBAgITBmyfz5m/jAo54vB4ikPmljZbyjANBgkqhkiG9w0BAQsF\nADA5MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6\nb24gUm9vdCBDQSAxMB4XDTE1MDUyNjAwMDAwMFoXDTM4MDExNzAwMDAwMFowOTEL\nMAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJv\nb3QgQ0EgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALJ4gHHKeNXj\nca9HgFB0fW7Y14h29Jlo91ghYPl0hAEvrAIthtOgQ3pOsqTQNroBvo3bSMgHFzZM\n9O6II8c+6zf1tRn4SWiw3te5djgdYZ6k/oI2peVKVuRF4fn9tBb6dNqcmzU5L/qw\nIFAGbHrQgLKm+a/sRxmPUDgH3KKHOVj4utWp+UhnMJbulHheb4mjUcAwhmahRWa6\nVOujw5H5SNz/0egwLX0tdHA114gk957EWW67c4cX8jJGKLhD+rcdqsq08p8kDi1L\n93FcXmn/6pUCyziKrlA4b9v7LWIbxcceVOF34GfID5yHI9Y/QCB/IIDEgEw+OyQm\njgSubJrIqg0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMC\nAYYwHQYDVR0OBBYEFIQYzIU07LwMlJQuCFmcx7IQTgoIMA0GCSqGSIb3DQEBCwUA\nA4IBAQCY8jdaQZChGsV2USggNiMOruYou6r4lK5IpDB/G/wkjUu0yKGX9rbxenDI\nU5PMCCjjmCXPI6T53iHTfIUJrU6adTrCC2qJeHZERxhlbI1Bjjt/msv0tadQ1wUs\nN+gDS63pYaACbvXy8MWy7Vu33PqUXHeeE6V/Uq2V8viTO96LXFvKWlJbYK8U90vv\no/ufQJVtMVT8QtPHRh8jrdkPSHCa2XV4cdFyQzR1bldZwgJcJmApzyMZFo6IQ6XU\n5MsI+yMRQ+hDKXJioaldXgjUkK642M4UwtBV8ob2xJNDd2ZhwLnoQdeXeGADbkpy\nrqXRfboQnoZsG4q5WTP468SQvvG5\n-----END CERTIFICATE-----"
#   ca_verify = true
#   timeout = 10
#   open_timeout = 20

#   event {
#     event_type   = "siem"
#     active       = true
#     with_headers = true
#   }
#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "number_of_requests_per_hour"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_integration_slack" "slack_integration" {
#   name = "Terraform Slack Integration X"
#   webhook_url = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
#   active = true

#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_integration_teams" "teams_integration" {
#   name = "Terraform MS Teams Integration"
#   webhook_url = "https://xxxxx.webhook.office.com/xxxxxxxxx"
#   active = true

#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_integration_telegram" "telegram_integration" {
#   name = "Terraform Telegram Integration"
#   telegram_username = "WallarmIntegrationTest"
#   chat_data = "+y86q0LOQ4QG3hK9QgVDfw=="
#   active = true

#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
#   event {
# 		event_type = "report_daily"
# 		active = false
# 	}
# 	event {
# 		event_type = "report_weekly"
# 		active = false
# 	}
# 	event {
# 		event_type = "report_monthly"
# 		active = false
# 	}
# }

# resource "wallarm_integration_sumologic" "sumologic_integration" {
#   name = "Terraform SumoLogic Integration"
#   sumologic_url = "http://sumologic.com/changed/once/again"

#   event {
#     event_type   = "siem"
#     active       = true
#     with_headers = true
#   }
#   event {
#     event_type = "rules_and_triggers"
#     active = true
#   }
#   event {
#     event_type = "number_of_requests_per_hour"
#     active = true
#   }
#   event {
#     event_type = "security_issue_critical"
#     active = true
#   }
#   event {
#     event_type = "security_issue_high"
#     active = true
#   }
#   event {
#     event_type = "security_issue_medium"
#     active = true
#   }
#   event {
#     event_type = "security_issue_low"
#     active = true
#   }
#   event {
#     event_type = "security_issue_info"
#     active = true
#   }
#   event {
#     event_type = "system"
#     active = true
#   }
# }

# resource "wallarm_rule_dirbust_counter" "dirbust_counter" {
# }

# resource "wallarm_trigger" "dirbust_trigger" {
#   template_id = "forced_browsing_started"

#   filters {
#     filter_id = "hint_tag"
#     operator = "eq"
#     value = [wallarm_rule_dirbust_counter.dirbust_counter.counter]
#   }

#   actions {
#     action_id = "mark_as_brute"
#   }

#   actions {
#     action_id = "block_ips"
#     lock_time = 2592000
#   }

#   threshold = {
#     period = 30
#     operator = "gt"
#     count = 30
#   }
# }

# resource "wallarm_rule_bruteforce_counter" "brute_counter" {
#   action {
#     point = {
#       instance = 9
#     }
#   }

#   action {
#     type = "absent"
#     point = {
#       path = 0
#      }
#   }

#   action {
#     type = "equal"
#     point = {
#       action_name = "login"
#     }
#   }

#   action {
#     type = "absent"
#     point = {
#       action_ext = ""
#     }
#   }

#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   action {
#     type = "equal"
#     value = "admin"
#     point = {
#       query = "user"
#     }
#   }

# }

# resource "wallarm_trigger" "brute_trigger" {
#   template_id = "bruteforce_started"

#   filters {
#     filter_id = "hint_tag"
#     operator = "eq"
#     value = [wallarm_rule_bruteforce_counter.brute_counter.counter]
#   }

#   actions {
#     action_id = "mark_as_brute"
#   }

#   actions {
#     action_id = "block_ips"
#     lock_time = 600
#   }

#   threshold = {
#     period = 30
#     operator = "gt"
#     count = 1
#   }
# }

# resource "wallarm_integration_email" "email_integration" {
#   name = "New Terraform Integration"
#   emails = ["test1@example.com", "test2@example.com"]

#   event {
#     event_type = "report_monthly"
#     active = true
#   }

#   event {
#     event_type = "aasm_report"
#     active = true
#   }

# }

# resource "wallarm_trigger" "attack_trigger" {
#   name = "New Terraform Trigger Email"
#   enabled = false
#   template_id = "attacks_exceeded"

#   filters {
#     filter_id = "ip_address"
#     operator = "eq"
#     value = ["2.2.2.2"]
#   }

#   threshold = {
#     period = 86400
#     operator = "gt"
#     count = 10000
#   }

#   actions {
#     action_id = "send_notification"
#     integration_id = [wallarm_integration_email.email_integration.integration_id]
#   }

#   depends_on = [
#     wallarm_integration_email.email_integration
#   ]
# }
