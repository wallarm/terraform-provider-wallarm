#
# Define the source of a provider.
#
terraform {
  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
      version = ">= 1.0.0"
    }
  }
}

#
# Define provider parameters
#
provider "wallarm" {
  api_token = var.api_token
  api_host = var.api_host
  client_id = 133
}

# #
# # User section
# #
# resource "wallarm_user" "user" {
#   email = "testuser+6039@wallarm.com"
#   permissions = "deploy"
#   realname = "Test Deploy"
#   password = "vJdlSKJ_sdh2749sj!"
# }


# #
# # Denylist section
# #
# resource "wallarm_denylist" "denylist" {
#   ip_range = ["1.1.1.1/24"]
#   application = [1]
#   reason = "TEST DENYLIST"
#   time_format = "Minutes"
#   time = 60
# }


# #
# # Vpatch rule section
# #
# resource "wallarm_rule_vpatch" "default" {
#   attack_type =  ["sqli"]
#   point = [["get", "query"]]
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


# #
# # Regular expression rule section
# #
# resource "wallarm_rule_regex" "regex_curltool" {
#   regex = ".*curltool.*"
#   experimental = false
#   attack_type =  "vpatch"

#   action {
#     type = "iequal"
#     value = "tiredful-api.wallarm-demo.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   point = [["uri"]]
# }

# resource "wallarm_rule_regex" "scanner_rule" {
#   regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
#   experimental = true
#   attack_type =  "scanner"
#   action {
#     point = {
#       instance = 5
#     }
#   }
#   point = [["header", "X-AUTHENTICATION"]]
# }


# #
# # Ignore regex rule section
# #
# resource "wallarm_rule_ignore_regex" "ignore_regex" {
#   regex_id = wallarm_rule_regex.scanner_rule.regex_id
#   action {
#  }
#   point = [["header", "X-LOGIN"]]
#   depends_on = [wallarm_rule_regex.scanner_rule]
# }


# #
# # Mark information as sensitive rule section
# #
# resource "wallarm_rule_masking" "masking_json" {

#   action {
#     type = "equal"
#     point = {
#       action_name = "masking"
#     }
#   }

#   action {
#     type = "absent"
#     point = {
#       path = 0
#      }
#   }

#   action {
#     type = "absent"
#     point = {
#       action_ext = ""
#     }
#   }

#   point = [["post"],["json_doc"],["hash", "field"]]
# }


# #
# # WAF mode rule section
# #
# resource "wallarm_rule_mode" "tiredful_api_mode" {
#   mode =  "monitoring"

#   action {
#     point = {
#       instance = 9
#     }
#   }

#   action {
#     type = "regex"
#     point = {
#       scheme = "https"
#     }
#   }

# }


# #
# # Integration webhook section
# #
# resource "wallarm_integration_webhook" "wh_integration" {
#   name = "New Terraform WebHook Integration"
#   webhook_url = "https://webhook.wallarm.com"
#   http_method = "POST"
#   active = true

#   event {
#     event_type = "hit"
#     active = true
#   }
#   event {
#     event_type = "vuln"
#     active = true
#   }
#   headers = {
#     Authorization = "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg=="
#     Content-Type = "application/xml"
#   }
# }


# #
# # Trigger section
# #
# resource "wallarm_rule_bruteforce_counter" "example_counter" {
#   action {
# 	  type = "iequal"
#       value = "example.com"
#       point = {
#       header = "HOST"
#     }
#   }

#   action {
#     type = "iequal"
#     value = "foo"
#     point = {
#       path = 0
#     }
#   }
#
#   action {
#     type = "iequal"
#     value = "bar"
#     point = {
#       path = 1
#     }
#   }
# }

# resource "wallarm_trigger" "user_trigger" {
#   name = "New Terraform Trigger Telegram"
#   comment = "This is a description set by Terraform"
#   enabled = true
#   template_id = "bruteforce_started"

#   filters {
#     filter_id = "hint_tag"
#     operator = "eq"
#     value = [wallarm_rule_bruteforce_counter.example_counter.counter]
#   }

#   filters {
#     filter_id = "ip_address"
#     operator = "eq"
#     value = ["1.1.1.1"]
#   }

#   threshold = {
# 		period = 30
# 		operator = "gt"
# 		count = 30
# 	}

#   actions {
#     action_id = "mark_as_brute"
#   }

#   actions {
#     action_id = "block_ips"
#     lock_time = 60
#   }

# }

# #
# # Scanner scope section
# #
# resource "wallarm_scanner" "scan" {
#     element = ["1.1.1.1", "example.com", "2.2.2.2/31"]
#     disabled = true
# }

# output "scan_id" {
#   value = wallarm_scanner.scan.resource_id
# }


# #
# # Global mode section
# #
# resource "wallarm_global_mode" "global_block" {
#   filtration_mode = "block"
#   scanner_mode = "off"
#   rechecker_mode = "on"
# }


# #
# # Application section
# #
# resource "wallarm_application" "tf_app" {
#   name = "New Terraform Application"
#   app_id = 42
# }


# #
# # Node section
# #
# variable "node_names" {
#   description = "Create Node names"
#   type        = list(string)
#   default     = ["prod", "stage", "dev"]
# }

# resource "wallarm_node" "cloud_node" {
#   # count = 3
#   # hostname = "tf-${var.node_names[count.index]}"
#   hostname = "tf-test"
# }


# # #######Extra Rules#######


# #
# # Attack rechecker mode rule section
# #
# resource "wallarm_rule_attack_rechecker" "disable_rechecker" {
#   enabled =  false

#   action {
#     point = {
#       instance = 6
#     }
#   }
# }


# #
# # Set response headers rule section
# #
# resource "wallarm_rule_set_response_header" "resp_headers" {
#   mode = "append"

#   action {
#     point = {
#       instance = 6
#     }
#   }

#   headers = {
#     Server = "Wallarm WAF"
#     Blocked = "Blocked by Wallarm"
#   }

# }

# #
# # Attack rechecker rewrite rule section
# #
# resource "wallarm_rule_attack_rechecker_rewrite" "default_rewrite" {
#   rules =  ["my.awesome-application.com", "my.example.com"]
#   point = [["header", "HOST"]]
# }

# # Test
# resource "wallarm_node" "terraform" {
#   client_id = 6039
#   hostname = "Terraform Tests"
# }


# resource "wallarm_rule_mode" "test" {

# 	# attack_type = ["any", "sqli", "rce", "crlf", "nosqli", "xxe", "ptrav", "xss", "scanner", "redir", "ldapi"]
# #   attack_type = ["any"]
#   mode =  "monitoring"
#   client_id = 6039

#   action {
# 		point = {
# 		  instance = 1
# 		}
# 	}

#   action {
# 		point = {
# 		  instance = 1
# 		}
# 	}

# 	action {
# 		type = "iequal"
# 		point = {
# 		  action_name = "masking"
# 		}
# 	}

#   action {
# 		type = "iequal"
# 		point = {
# 		  action_name = "masking"
# 		}
# 	}

# 	action {
# 		type = "absent"
# 		point = {
# 		  action_ext = ""
# 		}
# 	}

# 	action {
# 		type = "absent"
# 		point = {
# 		  path = 0
# 		}
# 	}

# 	action {
# 		type = "iequal"
# 		point = {
# 		  scheme = "hTTps"
# 		}
# 	}

# 	action {
# 		type = "iequal"
# 		point = {
# 		  method = "GET"
# 		}
# 	}

# 	action {
# 		type = "equal"
# 		point = {
# 		  proto = "1.1"
# 		}
# 	}

# 	action {
# 		type = "regex"
# 		point = {
# 		  uri = "/api/token[0-9A-Za-z]+"
# 		}
# 	}

# 	action {
# 		type = "iequal"
# 		value = "ExaMpLe.com"
# 		point = {
# 		  header = "HoST"
# 		}
# 	}

# 	# point = [["post"],["json_doc"],["array",0],["hash","password"]]
# }

# # resource "wallarm_rule_mode" "tiredful_api_mode" {
# #   mode =  "monitoring"
# # }



# resource "wallarm_rule_mode" "test" {

# 	mode = "block"

# 	action {
# 		point = {
# 		  instance = 9
# 		}
# 	}

# 	# Intentionally create a duplicate which is supposed to be removed by Set
# 	action {
# 		point = {
# 		  instance = 9
# 		}
# 	}

# 	action {
# 		type = "iequal"
# 		point = {
# 		  action_name = "wmode"
# 		}
# 	}

# 	action {
# 		type = "absent"
# 		point = {
# 		  action_ext = ""
# 		}
# 	}

# 	action {
# 		value = "api"
# 		type = "equal"
# 		point = {
# 		  path = 1
# 		}
# 	}

# 	action {
# 		value = "login"
# 		type = "iequal"
# 		point = {
# 		  path = 3
# 		}
# 	}

# 	action {
# 		type = "iequal"
# 		point = {
# 		  method = "PUT"
# 		}
# 	}

# 	action {
# 		type = "equal"
# 		point = {
# 		  scheme = "http"
# 		}
# 	}

# 	action {
# 		type = "equal"
# 		point = {
# 		  proto = "1.0"
# 		}
# 	}

# 	action {
# 		type = "regex"
# 		point = {
# 		  uri = "/console/username[0-9A-Za-z]+"
# 		}
# 	}

# 	action {
# 		type = "iequal"
# 		value = "https://docs.wallarm.com/admin-en/installation-nginx-en/"
# 		point = {
# 		  header = "referer"
# 		}
# 	}
# }



# resource "wallarm_rule_masking" "masking_json" {

#   action {
# 		value = "login"
# 		type = "iequal"
# 		point = {
# 		  path = 3
# 		}
# 	}

#   point = [["header","HOST"],["pollution"]]
# }

# resource "wallarm_rule_regex" "scanner_rule" {
#   regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
#   experimental = true
#   attack_type =  "scanner"
#   action {
#     point = {
#       instance = 5
#     }
#   }
#   point = [["header", "X-AUTHENTICATION"]]
# }

# resource "wallarm_rule_ignore_regex" "ignore_regex" {
#   regex_id = wallarm_rule_regex.scanner_rule.regex_id
#   action {
#     point = {
#       instance = 5
#     }
#   }
#   point = [["header", "X-LOGIN"]]
#   depends_on = [wallarm_rule_regex.scanner_rule]
# }

# resource "wallarm_integration_insightconnect" "insight_integration" {
#   name = "New Terraform InsightConnect Integration"
#   api_url = "https://us.api.insight.rapid7.com/connect/v1/workflows/d1763a97-e41b-1020-a651-26c1427657081/events/execute"
#   api_token = "c038033e-550a-0260-aa00-a102e5b356a7"
#   active = true

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
#     active = true
#   }
# }

# locals {
#   deploy = 1
# }

# resource "wallarm_rule_masking" "mask_headers" {

#   for_each = {for value in csvdecode(file("./RuleMasking.csv")):
#     value.name => value.type
#     if value.type == "headers" && local.deploy == 1
#   }

#   action {
# 		point = {
# 		  method = "POST|GET|PATCH"
# 		}
#     type = "regex"
# 	}

#   point = [["header", each.value]]
# }
