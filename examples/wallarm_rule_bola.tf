resource "wallarm_rule_bola" "user_id_bola_block" {
  mode = "block"

  action {
    type  = "iequal"
    value = "api.example.com"
    point = {
      header = "HOST"
    }
  }

  threshold {
    count  = 20
    period = 60
  }

  reaction {
    block_by_ip = 600
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["user_?id", "account_?id"]
    value_regexps         = [""]
    additional_parameters = false
    plain_parameters      = true
  }

  # Optional: only count requests that returned 200.
  advanced_conditions {
    field    = "status_code"
    operator = "eq"
    value    = ["200"]
  }
}
