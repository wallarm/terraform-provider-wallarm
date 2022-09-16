resource "wallarm_rule_binary_data" "binary" {
  action {
    type = "iequal"
    value = "binary.wallarm.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"],["form_urlencoded","query"]]
}
