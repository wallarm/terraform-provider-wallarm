resource "wallarm_rule_attack_rechecker" "disable_rechecker" {
  enabled = false

  action {
    point = {
      instance = 6
    }
  }

}