resource "wallarm_rule_credential_stuffing_point" "point1" {
  point = [["header", "HOST"]]
  login_point = [["header", "SESSION-ID"]]
}

resource "wallarm_rule_credential_stuffing_point" "point2" {
  client_id = 123

	action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  point = [["header", "HOST"]]
  login_point = ["header", "SESSION-ID"]
}
