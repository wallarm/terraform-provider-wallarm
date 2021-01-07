resource "wallarm_rule_attack_rechecker_rewrite" "default_rewrite" {
  rules =  ["my.awesome-application.com"]
  point = [["header", "HOST"]]
}