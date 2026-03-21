# ─── HCL Rule Examples (all 25 types + GQL point test) ────────────────────────
# Uncomment to create. Uses action_* scope fields for path-to-action expansion.

# # 1. binary_data
# resource "wallarm_rule_binary_data" "test" {
#   action_path   = "/api/v1/upload"
#   action_domain = "example.com"
#   point         = [["post"]]
# }

# # 2. masking (sensitive_data)
# resource "wallarm_rule_masking" "test" {
#   action_path     = "/api/v1/auth/login"
#   action_domain   = "secure.example.com"
#   action_instance = "17"
#   point           = [["post"], ["json_doc"], ["hash", "password"]]
# }

# # 3. disable_attack_type
# resource "wallarm_rule_disable_attack_type" "test" {
#   attack_type   = "sqli"
#   action_path   = "/api/v1/search"
#   action_domain = "app.example.com"
#   action_method = "POST"
#   point         = [["post"], ["form_urlencoded", "query"]]
# }

# # 4. disable_stamp
# resource "wallarm_rule_disable_stamp" "test" {
#   stamp         = 1001
#   action_path   = "/api/v2/users/*"
#   action_domain = "api.example.com"
#   point         = [["get", "id"]]
# }

# # 5. vpatch
# resource "wallarm_rule_vpatch" "test" {
#   attack_type   = "xss"
#   action_path   = "/api/**/comments"
#   action_domain = "forum.example.com"
#   point         = [["uri"]]
# }

# # 6. uploads
# resource "wallarm_rule_uploads" "test" {
#   file_type     = "docs"
#   action_path   = "/api/v1/documents"
#   action_domain = "files.example.com"
#   action_scheme = "https"
#   point         = [["post"]]
# }

# # 7. ignore_regex (disable_regex)
# resource "wallarm_rule_ignore_regex" "test" {
#   regex_id      = wallarm_rule_regex.test.regex_id
#   action_path   = "/api/v1/webhook"
#   action_domain = "hooks.example.com"
#   point         = [["post"], ["json_doc"]]
# }

# # 8. regex
# resource "wallarm_rule_regex" "test" {
#   attack_type   = "sqli"
#   regex         = ".*select.*from.*"
#   experimental  = false
#   action_path   = "/api/v1/*/*/*/*/*/*/query"
#   action_domain = "db.example.com"
#   action_method = "POST"
#   point         = [["post"], ["json_doc"], ["hash", "sql"]]
# }

# # 9. parser_state
# resource "wallarm_rule_parser_state" "test" {
#   parser        = "json_doc"
#   state         = "disabled"
#   action_path   = "/api/v1/raw"
#   action_domain = "data.example.com"
#   point         = [["post"]]
# }

# # 10. file_upload_size_limit
# resource "wallarm_rule_file_upload_size_limit" "test" {
#   mode          = "block"
#   size          = 10
#   size_unit     = "mb"
#   action_path   = "/api/v1/upload/*"
#   action_domain = "cdn.example.com"
#   point         = [["post"]]
# }

# # 11. rate_limit
# resource "wallarm_rule_rate_limit" "test" {
#   delay         = 100
#   burst         = 50
#   rate          = 200
#   rsp_status    = 429
#   time_unit     = "rps"
#   action_path   = "/api/v1/public"
#   action_domain = "gateway.example.com"
#   action_proto  = "1.1"
#   point         = [["get", "api_key"]]
# }

# # 12. credential_stuffing_point
# resource "wallarm_rule_credential_stuffing_point" "test" {
#   cred_stuff_type = "default"
#   login_point     = [["post"], ["json_doc"], ["hash", "email"]]
#   action_path     = "/auth/login"
#   action_domain   = "id.example.com"
#   action_method   = "POST"
#   action_scheme   = "https"
#   point           = [["post"], ["json_doc"], ["hash", "password"]]
# }

# # 13. credential_stuffing_regex
# resource "wallarm_rule_credential_stuffing_regex" "test" {
#   regex           = "\\w+@\\w+"
#   login_regex     = "\\w+@\\w+"
#   case_sensitive  = false
#   cred_stuff_type = "default"
#   action_path     = "/auth/register"
#   action_domain   = "id.example.com"
# }

# # 14. mode
# resource "wallarm_rule_mode" "test" {
#   mode          = "block"
#   action_path   = "/admin/**/*"
#   action_domain = "panel.example.com"
# }

# # 15. set_response_header
# resource "wallarm_rule_set_response_header" "test" {
#   name          = "X-Frame-Options"
#   mode          = "append"
#   values        = ["DENY"]
#   action_path   = "/"
#   action_domain = "web.example.com"
# }

# # 16. overlimit_res_settings
# resource "wallarm_rule_overlimit_res_settings" "test" {
#   mode           = "blocking"
#   overlimit_time = 5000
#   action_path    = "/api/v1/export.*"
#   action_domain  = "reports.example.com"
# }

# # 17. graphql_detection
# resource "wallarm_rule_graphql_detection" "test" {
#   mode              = "block"
#   max_depth         = 10
#   max_value_size_kb = 64
#   max_doc_size_kb   = 128
#   max_alias_size_kb = 32
#   max_doc_per_batch = 5
#   introspection     = false
#   debug_enabled     = false
#   action_path       = "/graphql"
#   action_domain     = "api.example.com"
#   action_method     = "POST"
# }

# # 18. brute
# resource "wallarm_rule_brute" "test" {
#   mode          = "block"
#   action_path   = "/auth/login"
#   action_domain = "auth.example.com"
#
#   threshold {
#     period = 300
#     count  = 10
#   }
#   reaction {
#     block_by_ip = 600
#   }
#   enumerated_parameters {
#     mode                  = "regexp"
#     name_regexps          = ["^password$"]
#     value_regexps         = [""]
#     additional_parameters = false
#     plain_parameters      = false
#   }
# }

# # 19. bruteforce_counter
# resource "wallarm_rule_bruteforce_counter" "test" {
#   action_path   = "/api/v1/tokens"
#   action_domain = "auth.example.com"
#   action_method = "POST"
# }

# # 20. dirbust_counter
# resource "wallarm_rule_dirbust_counter" "test" {
#   action_path     = "/static/**/*.*"
#   action_instance = "42"
# }

# # 21. bola
# resource "wallarm_rule_bola" "test" {
#   mode          = "block"
#   action_path   = "/api/v1/*/profile"
#   action_domain = "users.example.com"
#
#   threshold {
#     period = 600
#     count  = 20
#   }
#   reaction {
#     block_by_session = 3600
#   }
#   enumerated_parameters {
#     mode         = "regexp"
#     name_regexps = ["^id$"]
#   }
# }

# # 22. bola_counter
# resource "wallarm_rule_bola_counter" "test" {
#   action_path   = "/api/v2/*/orders"
#   action_domain = "shop.example.com"
# }

# # 23. enum
# resource "wallarm_rule_enum" "test" {
#   mode          = "block"
#   action_path   = "/api/v1/coupons"
#   action_domain = "promo.example.com"
#
#   threshold {
#     period = 120
#     count  = 5
#   }
#   reaction {
#     block_by_session = 600
#   }
#   enumerated_parameters {
#     mode = "exact"
#     points {
#       point     = ["post", "json_doc", "hash", "username"]
#       sensitive = false
#     }
#     points {
#       point     = ["post", "json_doc", "hash", "password"]
#       sensitive = true
#     }
#   }
# }

# # 24. rate_limit_enum
# resource "wallarm_rule_rate_limit_enum" "test" {
#   mode          = "block"
#   action_path   = "/api/v1/verify"
#   action_domain = "otp.example.com"
#
#   threshold {
#     period = 60
#     count  = 3
#   }
#   reaction {
#     block_by_ip = 1800
#   }
# }

# # 25. forced_browsing
# resource "wallarm_rule_forced_browsing" "test" {
#   mode          = "monitoring"
#   action_path   = "/admin/*"
#   action_domain = "internal.example.com"
#
#   action_header {
#     name  = "X-Forwarded-For"
#     value = ".*"
#     type  = "regex"
#   }
#
#   threshold {
#     period = 300
#     count  = 15
#   }
#   reaction {
#     graylist_by_ip = 7200
#   }
# }

# # 26. GQL point test (standard action block format)
# resource "wallarm_rule_parser_state" "disable_b64_gql" {
#   action {
#     type  = "iequal"
#     value = "example.com"
#     point = {
#       header = "HOST"
#     }
#   }
#   point  = [["post"], ["gql"], ["gql_query", "test"]]
#   parser = "base64"
#   state  = "disabled"
# }
