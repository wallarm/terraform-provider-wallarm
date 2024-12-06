resource "wallarm_rule_rate_limit" "rate_limit" {
  comment = "Example rate limit rule"

  action {
    type = "iequal"
    value = "example.com"
    point = {
      "header" = "HOST"
    }
  }
  action {
    type = "equal"
    value = "api"
    point = {
      "path" = 0
    }
  }
  action {
    type = "equal"
    value = "logon"
    point = {
      "path" = 1
    }
  }
  action {
    type = "equal"
    point = {
      "method" = "POST"
    }
  }
  action {
    point = {
      "instance" = 1
    }
  }
  action {
    type = "equal"
    point = {
      "scheme" = "https"
    }
  }

  point = [["post"], ["json_doc"], ["hash", "enter"]]

  delay      = 100
  burst      = 200
  rate       = 300
  rsp_status = 404
  time_unit  = "rps"
}