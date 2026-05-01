resource "wallarm_rule_rate_limit_enum" "api_dos_block" {
  mode = "block"

  action {
    type  = "iequal"
    value = "api.example.com"
    point = {
      header = "HOST"
    }
  }

  threshold {
    count  = 100
    period = 60
  }

  reaction {
    block_by_ip = 600
  }
}
