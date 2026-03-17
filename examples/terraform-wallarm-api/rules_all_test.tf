# # ## Total 23 rules + 3 counters + 2 creds_tuff

# resource "wallarm_rule_binary_data" "multipart_binary" {
#   comment = "Binary data"
#   point = [["post"], ["multipart_all"], ["file"]]
# }
# resource "wallarm_rule_binary_data" "binary" {
#   comment = "Binary data"
#   action {
#     type = "iequal"
#     value = "binary.wallarm.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["post"], ["form_urlencoded", "query"]]
# }
# resource "wallarm_rule_bola" "wallarm_rule_bola_regexp" {
#   mode = "block"
#   comment = "Bola enum"

#   action {
#     type = "iequal"
#     value = "wbola1.wallarm.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   reaction {
#     block_by_ip = 3600
#   }

#   threshold {
#     count = 5
#     period = 30
#   }
#   enumerated_parameters {
#     mode                  = "regexp"
#     name_regexps          = ["foo", "bar"]
#     value_regexps         = [""]
#   }
# }

# resource "wallarm_rule_brute" "wallarm_rule_brute_regexp" {
#   mode = "block"
#   action {
#     type = "iequal"
#     value = "wbrute1.wallarm.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   reaction {
#     block_by_ip = 3600
#   }

#   threshold {
#     count = 5
#     period = 30
#   }

#   enumerated_parameters {
#     mode                  = "regexp"
#     name_regexps          = ["foo", "bar"]
#     value_regexps         = ["baz"]
#     additional_parameters = false
#     plain_parameters      = false
#   }
# }
# resource "wallarm_rule_disable_attack_type" "disable_sqli" {
#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["get_all"]]
#   attack_type = "sqli"
#   title = "TEST SQLI"
#   set = "disable_sqli"
#   active = true
# }
# resource "wallarm_rule_disable_attack_type" "disable_xss" {
#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["get_all"]]
#   attack_type = "xss"
#   title = "TEST XSS"
#   set = "disable_xss"
# }
# resource "wallarm_rule_file_upload_size_limit" "file_upload_size_limit_1" {
#   mode = "block"

#   action {
#     type = "iequal"
#     value = "file-upload-size.wallarm.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   size = 1
#   size_unit = "mb"

#   point = [["post"]]
# }
# resource "wallarm_rule_forced_browsing" "forced_browsing_test" {
#   mode = "block"

#   action {
#     type = "iequal"
#     value = "wdirbust1.wallarm.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   reaction {
#     block_by_ip = 3600
#   }

#   threshold {
#     count = 5
#     period = 30
#   }
# }
# resource "wallarm_rule_graphql_detection" "graphql_detection_1" {
#   mode = "block"

#   action {
#     type = "iequal"
#     value = "gql.wallarm.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   max_depth = 10
#   max_value_size_kb = 10
#   max_doc_size_kb = 100
#   max_alias_size_kb = 5
#   max_doc_per_batch = 10
#   introspection = true
#   debug_enabled = true
# }
# resource "wallarm_rule_regex" "regex_curltool" {
#   regex = ".*curltool.*"
#   experimental = false
#   attack_type =  "vpatch"

#   action {
#     type = "iequal"
#     value = "front.example.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   title = "TEST REGEX"
#   set = "test"

#   point = [["uri"]]
# }
# resource "wallarm_rule_ignore_regex" "ignore_regex" {
#   regex_id = wallarm_rule_regex.regex_curltool.regex_id

#   action {
#     point = {
#       instance = 5
#     }
#   }

#   point = [["header", "X-AUTHENTICATION"]]
#   depends_on = [wallarm_rule_regex.regex_curltool]
#   title = "TEST IGNORE REGEX"
#   set = "test"
# }
# resource "wallarm_rule_regex" "scanner_rule" {
#   regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
#   experimental = true
#   attack_type = "scanner"
#   action {
#     point = {
#       instance = 5
#     }
#   }
#   point = [["header", "X-AUTHENTICATION"]]
#   title = "TEST EXP REGEX"
#   set = "test"
# }
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

#   action {
#     type = "equal"
#     value = "admin"
#     point = {
#       query = "user"
#     }
#   }

#   point = [["post"], ["json_doc"], ["hash", "field"]]

#   title = "TEST MASK"
#   set = "test"
# }
# resource "wallarm_rule_mode" "tiredful_api_mode" {
#   mode =  "monitoring"

#   action {
#     point = {
#       instance = 9
#     }
#   }

#   action {
#     type = "equal"
#     point = {
#       scheme = "https"
#     }
#   }

#   action {
#     type = "equal"
#     value = "admin"
#     point = {
#       query = "user"
#     }
#   }
#   title = "TEST MODE"
#   set = "test"
# }
# resource "wallarm_rule_mode" "tiredful_api_mode_ext" {
#   mode =  "monitoring"

#   action {
#     type = "equal"
#     value = "admin"
#     point = {
#       query = "user"
#     }
#   }
#   title = "TEST MODE 1"
#   set = "test"
# }
# resource "wallarm_rule_overlimit_res_settings" "example_overlimit_res_settings" {
#   action {
#     point = {
#       path = 0
#     }
#     type = "absent"
#   }
#   action {
#     point = {
#       action_name = "upload"
#     }
#     type = "equal"
#   }
#   action {
#     point = {
#       action_ext = ""
#     }
#     type = "absent"
#   }
#   mode = "blocking"
#   overlimit_time = 2000
#   title = "TEST OVERLIMIT"
#   set = "test"
# }
# resource "wallarm_rule_parser_state" "disable_htmljs_parsing" {
#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["get_all"]]
#   parser = "htmljs"
#   state = "disabled"
#   title = "TEST PARSER"
#   set = "test"
# }
# resource "wallarm_rule_parser_state" "disable_b64_gql" {
#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["post"], ["gql"], ["gql_query", "test"]]
#   parser = "base64"
#   state = "disabled"
# }
# resource "wallarm_rule_rate_limit" "rate_limit_api" {
#   action {
#     type = "equal"
#     value = "api"
#     point = {
#       path = 0
#     }
#   }
#   action {
#     point = {
#       instance = 1
#     }
#   }

#   point = [["post"], ["json_doc"], ["hash", "email"]]

#   delay      = 100
#   burst      = 200
#   rate       = 300
#   rsp_status = 404
#   time_unit  = "rps"
#   title = "TEST RATE"
#   set = "test"
# }
# resource "wallarm_rule_rate_limit_enum" "rule_rate_limit_enum_" {
#   mode = "block"

#   action {
#     type = "iequal"
#     value = "wenum1.wallarm.com"
#     point = {
#       header = "HOST"
#     }
#   }

#   reaction {
#     block_by_ip = 3600
#   }

#   threshold {
#     count = 5
#     period = 30
#   }
# }
# resource "wallarm_rule_set_response_header" "set_response_header_01143deef5893a6aa3128b13dfababfc" {
#   mode = "append"

#   action {
#     point = {
#       instance = 3
#     }
#   }

#   name = "Server"
#   values = ["Wallarm solution", "Blocked by Wallarm"]
#   title = "TEST RSP HEADER"
#   set = "test"
# }
# resource "wallarm_rule_uploads" "allow_markup_in_body" {
#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point = [["post"]]
#   file_type = "html"
#   title = "TEST UPLOADS"
#   set = "test"
# }
# resource "wallarm_rule_vpatch" "splunk" {
#   attack_type = "sqli"

#   action {
#     type = "iequal"
#     value = "app.example.com"

#     point = {
#       header = "HOST"
#     }

#   }

#   title = "TEST VPATCH"
#   set = "test"

#   point = [["get_all"]]
# }

# resource "wallarm_rule_credential_stuffing_point" "credential_stuffing_point" {
#   point = [["post"], ["json_doc"], ["hash", "password"]]
#   login_point = [["post"], ["json_doc"], ["hash", "login"]]
#   cred_stuff_type = "default"
#   comment = "Test cred stuff"

#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
# }
# resource "wallarm_rule_credential_stuffing_regex" "credential_stuffing_regex" {
#   regex = "abc.*"
#   login_regex = "def.*"
#   case_sensitive = true
#   cred_stuff_type = "custom"

#   action {
#     type = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
# }

# ##### Counters

# # # resource "wallarm_rule_bola_counter" "bola_counter_test" {
# # #   comment = "This is a comment for a test bola counter"

# # #   action {
# # #     type = "absent"
# # #     point = {
# # #       path = 0
# # #     }
# # #   }

# # # 	action {
# # # 		type = "iequal"
# # #     point = {
# # # 			action_name = "login"
# # #     }
# # #   }
# # # 	action {
# # # 		type = "equal"
# # #     point = {
# # # 			action_ext = "aspx"
# # #     }
# # #   }
# # # }

# # # resource "wallarm_rule_bruteforce_counter" "brute_counter_test" {
# # #   action {
# # #     type = "equal"
# # #     value = "api"
# # #     point = {
# # #       path = 0
# # #     }
# # #   }
# # # }
# # # resource "wallarm_rule_dirbust_counter" "dirbust_counter_test" {
# # # 	comment = "This is a comment for a test Forced browsing counter"

# # #   action {
# # #     type = "iequal"
# # #     value = "example.com"
# # #     point = {
# # #       header = "HOST"
# # #     }
# # #   }
# # # }

