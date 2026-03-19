# ═══════════════════════════════════════════════════════════════════════════════
# Custom rules examples — copy the object for your resource type, fill in values,
# and add it to the `custom_rules` list in terraform.tfvars.
#
# Common fields (available for all resource types):
#   name          = "unique_rule_name"          # Required. Used as config filename and resource key.
#   resource_type = "wallarm_rule_xxx"          # Required. Determines which Terraform resource is created.
#   comment       = "Managed by Terraform"      # Optional. Rule description in Wallarm console.
#   point         = [["uri"]]                   # Optional. Detection point (list of lists).
#   path          = "/api/v1/endpoint"          # Optional. Auto-expanded into action conditions.
#   domain        = "example.com"               # Optional. Matched via HOST header (case-insensitive).
#   instance      = "101"                       # Optional. Application pool ID.
#   method        = "POST"                      # Optional. HTTP method (case-insensitive).
#   scheme        = "https"                     # Optional. URL scheme.
#   proto         = "1.1"                       # Optional. HTTP protocol version.
#   query         = "key=value"                 # Optional. Query string match.
# ═══════════════════════════════════════════════════════════════════════════════

# custom_rules = [
#   # ─── wallarm_mode ──────────────────────────────────────────────────────────────────
#   # Set Wallarm filtration mode for matching requests.
#   # mode: "monitoring", "safe_blocking", "block", "off", "default"
#   {
#     name          = "block_admin"
#     comment       = "Wallarm Mode rule"
#     resource_type = "wallarm_rule_mode"
#     mode          = "block"
#     path          = "/api/v1/*/*/*/*/*/*/action.*"
#     query         = [
#       { key = "key1", value = "value_1" },
#       { key = "key2", value = "value_2" }
#     ]
#     domain        = "super-example.com"
#     scheme        = "https"
#     proto         = "1.1"
#     method        = "POST"
#     headers = [
#       { name = "WLRM-TEST", value = "aaabbbccc" }, # type defaults to equal
#       { name = "Content-Type", value = "application/json" },
#     ]
#   },

#   #─── binary_data ────────────────────────────────────────────────────────────
#   # Mark a request part as containing binary data (skip parsing).
#   {
#     name          = "binary_upload"
#     resource_type = "wallarm_rule_binary_data"
#     point         = [["post"], ["multipart_all"], ["file"]]
#     path          = "/**/*.*"
#     domain        = "wallarm.net"
#     instance      = "1"
#   },

#   # ─── masking (sensitive_data) ───────────────────────────────────────────────
#   # Mask sensitive data in a request part so it's not stored.
#   {
#     name          = "post_body_masking"
#     resource_type = "wallarm_rule_masking"
#     point         = [["post"]]
#     path          = "/api/v1/endpoint"
#     domain        = "example.com"
#     instance      = "1"
#   },

  # # ─── disable_attack_type ────────────────────────────────────────────────────
  # # Disable detection of specific attack types for matching requests.
  # # One rule is created per attack_type in the list.
  # {
  #   name          = "disable_sqli_xss"
  #   resource_type = "wallarm_rule_disable_attack_type"
  #   attack_types  = ["sqli", "xss"]
  #   point         = [["post"], ["form_urlencoded", "query"]]
  #   path          = "/api/health"
  #   domain        = "example.com"
  # },

  # # ─── disable_stamp ─────────────────────────────────────────────────────────
  # # Disable specific detection stamps (signature IDs).
  # # One rule is created per stamp in the list.
  # {
  #   name          = "disable_stamps_login"
  #   resource_type = "wallarm_rule_disable_stamp"
  #   stamps        = [1001, 1002, 1003]
  #   point         = [["uri"]]
  #   path          = "/login"
  #   domain        = "example.com"
  # },

  # # ─── vpatch ────────────────────────────────────────────────────────────────
  # # Create a virtual patch — block specific attack types without changing app code.
  # # One rule is created per attack_type in the list.
  # {
  #   name          = "vpatch_rce"
  #   resource_type = "wallarm_rule_vpatch"
  #   attack_types  = ["rce", "ssti"]
  #   point         = [["get_all"]]
  #   path          = "/admin/exec"
  #   domain        = "example.com"
  # },

  # # ─── uploads ───────────────────────────────────────────────────────────────
  # # Configure file upload handling for a specific file type.
  # {
  #   name          = "allow_docs_uploads"
  #   resource_type = "wallarm_rule_uploads"
  #   file_type     = "docs"
  #   point         = [["post"]]
  #   path          = "/api/documents"
  #   domain        = "example.com"
  # },

  # # ─── parser_state ──────────────────────────────────────────────────────────
  # # Enable or disable a specific parser for matching requests.
  # {
  #   name          = "disable_json_parser"
  #   resource_type = "wallarm_rule_parser_state"
  #   parser        = "json_doc"
  #   state         = "disabled"
  #   point         = [["post"]]
  #   path          = "/api/raw"
  #   domain        = "example.com"
  # },

  # # ─── regex ─────────────────────────────────────────────────────────────────
  # # Add a custom regex-based detection rule.
  # {
  #   name          = "detect_custom_injection"
  #   resource_type = "wallarm_rule_regex"
  #   attack_type   = "sqli"
  #   regex         = ".*regex_pattern.*"
  #   experimental  = false
  #   point         = [["uri"]]
  #   path          = "/api/search"
  #   domain        = "example.com"
  # },

  # # ─── experimental_regex ─────────────────────────────────────────────────────────────────
  # # Add a custom regex-based detection rule.
  # {
  #   name          = "detect_xxx_injection"
  #   resource_type = "wallarm_rule_regex"
  #   attack_type   = "sqli"
  #   regex         = ".*xxx.*"
  #   experimental  = true
  #   point         = [["uri"]]
  #   path          = "/api/test/action.php"
  #   domain        = "example.com"
  # },

  # # ─── ignore_regex (disable_regex) ──────────────────────────────────────────
  # # Disable a specific regex-based detection rule by its ID.
  # {
  #   name          = "ignore_regex_42"
  #   resource_type = "wallarm_rule_ignore_regex"
  #   regex_rule    = "detect_custom_injection"
  #   point         = [["uri"]]
  #   path          = "/api/webhook"
  #   domain        = "example.com"
  # },

  # {
  #   name          = "detect_ssn"
  #   resource_type = "wallarm_rule_regex"
  #   attack_type   = "vpatch"
  #   regex         = "b-\\d{3}-\\d{2}-\\d{4}-b"
  #   point         = [["post"]]
  #   path          = "/api/submit"
  #   domain        = "example.com"
  # },
  # {
  #   name          = "ignore_ssn_debug"
  #   resource_type = "wallarm_rule_ignore_regex"
  #   regex_rule    = "detect_ssn"        # ← references the name above
  #   point         = [["post"], ["multipart_all"]]
  #   path          = "/api/debug"
  #   domain        = "example.com"
  # },


  # # ─── file_upload_size_limit ────────────────────────────────────────────────
  # # Set a file upload size limit for matching requests.
  # {
  #   name          = "limit_upload_10mb"
  #   resource_type = "wallarm_rule_file_upload_size_limit"
  #   mode          = "block"
  #   size          = 10
  #   size_unit     = "mb"
  #   point         = [["post"], ["json_doc"], ["hash", "field"]]
  #   path          = "/api/upload"
  #   domain        = "example.com"
  # },

  # # ─── rate_limit ────────────────────────────────────────────────────────────
  # # Apply rate limiting to matching requests.
  # {
  #   name          = "rate_limit_api"
  #   resource_type = "wallarm_rule_rate_limit"
  #   delay         = 0
  #   burst         = 5
  #   rate          = 100
  #   rsp_status    = 503
  #   time_unit     = "rps"
  #   point         = [["uri"]]
  #   path          = "/api/v1"
  #   domain        = "example.com"
  # },

  # # ─── credential_stuffing_point ─────────────────────────────────────────────
  # # Define credential stuffing detection points.
  # {
  #   name            = "cred_stuff_login"
  #   resource_type   = "wallarm_rule_credential_stuffing_point"
  #   point           = [["post"], ["json_doc"], ["hash", "password"]]
  #   login_point     = [["post"], ["json_doc"], ["hash", "login"]]
  #   cred_stuff_type = "default"
  #   path            = "/auth/login"
  #   domain          = "example.com"
  # },

  # # ─── credential_stuffing_regex ─────────────────────────────────────────────
  # # Define credential stuffing detection via regex patterns.
  # {
  #   name            = "cred_stuff_regex_login"
  #   resource_type   = "wallarm_rule_credential_stuffing_regex"
  #   regex           = "^(password(\\d|confirm)|pwd|client(|\\.|-|_|)secret)$"
  #   login_regex     = "(user|usr)(|_|-|-.)(name|login)(|[\\d])$"
  #   case_sensitive  = false
  #   cred_stuff_type = "default"
  #   path            = "/auth/login"
  #   domain          = "example.com"
  # },

  # # ─── set_response_header ───────────────────────────────────────────────────
  # # Add or modify response headers for matching requests.
  # {
  #   name          = "add_security_headers"
  #   resource_type = "wallarm_rule_set_response_header"
  #   header_name   = "X-Content-Type-Options"
  #   header_mode   = "replace"
  #   header_values = ["nosniff"]
  #   path          = "/"
  #   domain        = "example.com"
  # },

  # # ─── overlimit_res_settings ────────────────────────────────────────────────
  # # Configure overlimit resource processing settings.
  # {
  #   name           = "overlimit_api"
  #   resource_type  = "wallarm_rule_overlimit_res_settings"
  #   overlimit_time = 3000
  #   mode           = "blocking"
  #   path           = "/api"
  #   domain         = "example.com"
  # },

  # # ─── graphql_detection ─────────────────────────────────────────────────────
  # # Configure GraphQL-specific detection and limits.
  # {
  #   name              = "graphql_api"
  #   resource_type     = "wallarm_rule_graphql_detection"
  #   mode              = "block"
  #   max_depth         = 10
  #   max_value_size_kb = 100
  #   max_doc_size_kb   = 1024
  #   max_alias_size_kb = 50
  #   max_doc_per_batch = 5
  #   introspection     = false
  #   debug_enabled     = false
  #   path              = "/graphql"
  #   domain            = "example.com"
  # },

  # # ─── brute ─────────────────────────────────────────────────────────────────
  # # Configure brute-force protection with threshold and reaction.
  # {
  #   name          = "brute_login"
  #   resource_type = "wallarm_rule_brute"
  #   mode          = "block"
  #   path          = "/auth/login"
  #   domain        = "example.com"
  #   threshold = {
  #     period = 300
  #     count  = 10
  #   }
  #   reaction = {
  #     block_by_ip      = 600
  #   }
  #   enumerated_parameters = {
  #     mode             = "regexp"
  #     name_regexps     = ["^password$"]
  #     value_regexps    = [""]
  #     additional_parameters = false
  #     plain_parameters      = false
  #   }
  # },

  # # ─── bola ──────────────────────────────────────────────────────────────────
  # # Configure BOLA (Broken Object-Level Authorization) protection.
  # {
  #   name          = "bola_users_api"
  #   resource_type = "wallarm_rule_bola"
  #   mode          = "block"
  #   path          = "/api/users"
  #   domain        = "example.com"
  #   threshold = {
  #     period = 300
  #     count  = 20
  #   }
  #   reaction = {
  #     block_by_session = 600
  #   }
  #   enumerated_parameters = {
  #     mode         = "exact"
  #     points       = [{ point = ["get","test"], sensitive = true }]
  #   }
  # },

  # # ─── enum ──────────────────────────────────────────────────────────────────
  # # Configure account enumeration protection.
  # {
  #   name          = "enum_registration"
  #   resource_type = "wallarm_rule_enum"
  #   mode          = "block"
  #   path          = "/api/register"
  #   domain        = "example.com"
  #   threshold = {
  #     period = 300
  #     count  = 15
  #   }
  #   reaction = {
  #     block_by_session = 600
  #   }
  #   enumerated_parameters = {
  #     mode             = "regexp"
  #     name_regexps     = ["^email$"]
  #     value_regexps    = [""]
  #     additional_parameters = false
  #     plain_parameters      = false
  #   }
  # },

  # # ─── rate_limit_enum ───────────────────────────────────────────────────────
  # # Configure rate-limit-based enumeration protection.
  # {
  #   name          = "rate_limit_enum_search"
  #   resource_type = "wallarm_rule_rate_limit_enum"
  #   mode          = "block"
  #   path          = "/api/search"
  #   domain        = "example.com"
  #   threshold = {
  #     period = 60
  #     count  = 30
  #   }
  #   reaction = {
  #     block_by_session = 600
  #   }
  # },

  # # ─── forced_browsing ──────────────────────────────────────────────────────
  # # Configure forced browsing (directory traversal) protection.
  # {
  #   name          = "forced_browsing_root"
  #   resource_type = "wallarm_rule_forced_browsing"
  #   mode          = "block"
  #   path          = "/"
  #   domain        = "example.com"
  #   threshold = {
  #     period = 300
  #     count  = 50
  #   }
  #   reaction = {
  #     block_by_ip = 600
  #   }
  # },

  # # ─── trigger_counters───────────────────────────────────────────────────────
  # # ─── bruteforce_counter ────────────────────────────────────────────────────
  # # Mark an endpoint as a brute-force counter point.
  # {
  #   name          = "bruteforce_counter_login"
  #   resource_type = "wallarm_rule_bruteforce_counter"
  #   path          = "/auth/login"
  #   domain        = "example.com"
  # },

  # # ─── dirbust_counter ──────────────────────────────────────────────────────
  # # Mark an endpoint as a directory busting counter point.
  # {
  #   name          = "dirbust_counter_static"
  #   resource_type = "wallarm_rule_dirbust_counter"
  #   path          = "/static"
  #   domain        = "example.com"
  # },

  # # ─── bola_counter ──────────────────────────────────────────────────────────
  # # Mark an endpoint as a BOLA counter point.
  # {
  #   name          = "bola_counter_users"
  #   resource_type = "wallarm_rule_bola_counter"
  #   path          = "/api/users"
  #   domain        = "example.com"
  # },
# ]
