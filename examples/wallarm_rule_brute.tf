resource "wallarm_rule_brute" "login_brute_block" {
  mode = "block"

  action {
    type  = "iequal"
    value = "auth.example.com"
    point = {
      header = "HOST"
    }
  }

  threshold {
    count  = 5
    period = 30
  }

  reaction {
    block_by_ip      = 600
    block_by_session = 600
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["user(name)?", "email"]
    value_regexps         = [""]
    additional_parameters = false
    plain_parameters      = true
  }
}

resource "wallarm_rule_brute" "login_brute_monitoring" {
  mode = "monitoring"

  action {
    type  = "iequal"
    value = "monitor.example.com"
    point = {
      header = "HOST"
    }
  }

  threshold {
    count  = 10
    period = 60
  }

  reaction {
    graylist_by_ip = 600
  }

  enumerated_parameters {
    mode = "exact"
    points {
      point     = ["post", "form_urlencoded", "username"]
      sensitive = false
    }
    points {
      point     = ["post", "form_urlencoded", "password"]
      sensitive = true
    }
  }
}
