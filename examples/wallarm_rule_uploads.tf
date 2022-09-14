resource "wallarm_rule_uploads" "allow_docs" {
  action {
    type = "iequal"
    value = "uploads.wallarm.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"],["form_urlencoded","query"]]
  file_type = "docs"
}
