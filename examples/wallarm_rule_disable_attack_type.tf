resource "wallarm_rule_disable_attack_type" "ignore_attack" {
  action {
    type = "iequal"
    value = "attack-types.wallarm.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"],["form_urlencoded","query"]]
  attack_type = "sqli"
}
