resource "wallarm_rule_credential_stuffing_mode" "mode1" {

}

resource "wallarm_rule_credential_stuffing_mode" "mode2" {
  client_id = 123

	action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  mode = "custom"
}
