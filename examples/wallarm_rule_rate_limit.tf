resource "wallarm_rate_limit" "example" {
  comment    = "Example rate limit rule"

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

  point = ["example_point_1", "example_point_2"]

  delay      = 100
  burst      = 200
  rate       = 300
  rsp_status = 404
  time_unit  = "rps"
}