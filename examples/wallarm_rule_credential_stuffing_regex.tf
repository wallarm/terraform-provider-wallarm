resource "wallarm_rule_credential_stuffing_regex" "regex1" {
  regex = "*abc*"
  login_regex = "user*"
  case_sensitive = false
}

resource "wallarm_rule_credential_stuffing_regex" "regex2" {
  client_id = 123

	action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  regex = "*abc*"
  login_regex = "user*"
  case_sensitive = true
  cred_stuff_type = "custom"
}
