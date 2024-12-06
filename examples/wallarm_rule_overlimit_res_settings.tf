resource "wallarm_rule_overlimit_res_settings" "example_overlimit_res_settings" {
  action {
    point = {
      "path" = 0
    }
    type = "absent"
  }
  action {
    point = {
      "action_name" = "upload"
    }
    type = "equal"
  }
  action {
    point = {
      "action_ext" = ""
    }
    type = "absent"
  }
  mode = "blocking"
  overlimit_time = 2000
}