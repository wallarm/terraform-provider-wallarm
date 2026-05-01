resource "wallarm_rule_enum" "param_enum_block" {
  mode = "block"

  action {
    type  = "iequal"
    value = "api.example.com"
    point = {
      header = "HOST"
    }
  }

  threshold {
    count  = 50
    period = 60
  }

  reaction {
    block_by_ip = 600
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = [""]
    additional_parameters = false
    plain_parameters      = true
  }
}

resource "wallarm_rule_enum" "exact_referer_monitoring" {
  mode = "monitoring"

  action {
    type  = "iequal"
    value = "monitor.example.com"
    point = {
      header = "HOST"
    }
  }

  threshold {
    count  = 100
    period = 60
  }

  reaction {
    graylist_by_ip = 600
  }

  enumerated_parameters {
    mode = "exact"
    points {
      point     = ["header", "REFERER"]
      sensitive = false
    }
    points {
      point     = ["get", "id"]
      sensitive = true
    }
  }
}
