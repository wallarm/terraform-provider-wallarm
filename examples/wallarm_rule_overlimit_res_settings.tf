resource "wallarm_overlimit_res_settings_rule" "example_overlimit_res_settings" {
  comment    = "Example overlimit res settings rule"

  action = {
    type  = "equal"
    value = "example_value"
    point = {
      header       = ["X-Example-Header"]
      method       = "GET"
      path         = 10
      action_name  = "example_action"
      action_ext   = "example_extension"
      query        = "example_query"
      proto        = "HTTP/1.1"
      scheme       = "https"
      uri          = "/example_uri"
      instance     = 1
    }
  }

  overlimit_time = 2000
  mode  = "monitoring"
}