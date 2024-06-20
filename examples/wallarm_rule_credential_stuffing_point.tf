resource "wallarm_rule_credential_stuffing_point" "point1" {
  point = [["HEADER", "HOST"]]
  login_point = ["HEADER", "SESSION-ID"]
}

resource "wallarm_rule_credential_stuffing_point" "point2" {
  client_id = 123

	action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  point = [["HEADER", "HOST"]]
  login_point = ["HEADER", "SESSION-ID"]
}
