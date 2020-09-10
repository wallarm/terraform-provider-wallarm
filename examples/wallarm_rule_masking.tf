resource "wallarm_rule_masking" "dvwa_sensitive" {

  action {
    point = {
      instance = 5
    }
  }

  point = [["header", "X-KEY"]]
}

resource "wallarm_rule_masking" "masking_header" {

  action {
    type = "absent"
    point = {
      path = 0
    }
  }

  action {
    type = "equal"
    point = {
      action_name = "masking"
    }
  }

  action {
    type = "absent"
    point = {
      action_ext = ""
    }
  }
  point = [["header", "X-KEY"]]
}

resource "wallarm_rule_masking" "masking_json" {

  action {
    type = "absent"
    point = {
      path = 0
    }
  }

  action {
    type = "equal"
    point = {
      action_name = "masking"
    }
  }

  action {
    type = "absent"
    point = {
      action_ext = ""
    }
  }
  point = [["post"],["json_doc"],["hash", "field"]]
}