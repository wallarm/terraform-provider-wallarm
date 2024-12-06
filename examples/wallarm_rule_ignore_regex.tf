resource "wallarm_rule_ignore_regex" "ingore_regex" {
  regex_id =  100365
  
  point = [["header", "X-LOGIN"]]
}
