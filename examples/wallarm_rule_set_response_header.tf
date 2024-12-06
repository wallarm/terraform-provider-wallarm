resource "wallarm_rule_set_response_header" "resp_headers" {
  mode = "replace"

  action {
    point = {
      "instance" = 6
    }
  }

  name = "Server"
  values = ["Wallarm", "Yes, you are blocked"]
}

resource "wallarm_rule_set_response_header" "resp_headers_waf" {
  mode = "append"

  action {
    point = {
      "instance" = 6
    }
  }

  name = "WAF"
  values = ["Wallarm Blocked"]
}
