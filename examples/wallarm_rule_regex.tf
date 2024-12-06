resource "wallarm_rule_regex" "regex" {
  regex = "[^0-9a-f]|^.{33,}$|^.{0,31}$"
  experimental = true
  attack_type = "redir"

  action {
    type = "iequal"
    value = "tiredful-api.wallarm-demo.com"
    point = {
      "header" = "HOST"
    }
  }
  point = [["header", "X-AUTHENTICATION"]]
}