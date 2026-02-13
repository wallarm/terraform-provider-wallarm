resource "wallarm_rule_disable_stamp" "example" {
  action {
    type  = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
  stamp = 1234
}
