resource "wallarm_rule_set_response_header" "resp_headers" {
  mode = "replace"

  action {
    point = {
      instance = 6
    }
  }

  headers = {
    Server = "Wallarm"
    Blocked = "Yes, you are"
  }

}

resource "wallarm_rule_set_response_header" "resp_headers" {
  mode = "append"

  action {
    point = {
      instance = "6"
    }
  }

  headers = {
    Server = "Wallarm WAF"
    Blocked = "Wallarm Blocked"
  }

}
