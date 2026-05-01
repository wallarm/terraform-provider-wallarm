resource "wallarm_rule_forced_browsing" "dirbust_block" {
  mode = "block"

  action {
    type  = "iequal"
    value = "api.example.com"
    point = {
      header = "HOST"
    }
  }

  threshold {
    count  = 30
    period = 60
  }

  reaction {
    block_by_ip = 600
  }

  # Optional: only count 404 responses (typical dirbust signal).
  advanced_conditions {
    field    = "status_code"
    operator = "eq"
    value    = ["404"]
  }
}
