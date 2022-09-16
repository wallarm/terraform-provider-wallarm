resource "wallarm_rule_parser_state" "enable_base64_parser" {
  action {
    type = "iequal"
    value = "parsers.wallarm.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"]]
  parser = "base64"
  state = "enabled"
}

resource "wallarm_rule_parser_state" "disable_gzip_parser" {
  point = [["header","HOST"],["pollution"]]
  parser = "gzip"
  state = "disabled"
}
