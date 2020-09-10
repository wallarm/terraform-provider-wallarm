
resource "wallarm_rule_mode" "wp_mode" {
  mode =  "block"

  action {
    point = {
      instance = 6
    }
  }

  action {
    type = "iequal"
    value = "monitor"
    point = {
      path = 0
    }
  }
}

resource "wallarm_rule_mode" "tiredful_api_mode" {
  mode =  "monitoring"

  action {
    point = {
      instance = "9"
    }
  }

  action {
    type = "equal"
    point = {
      action_name = "formmail"
    }
  }
}


resource "wallarm_rule_mode" "ad_mode" {
  mode =  "default"

  action {
    type = "equal"
    value = "api"
    point = {
      path = 0
    }
  }

  action {
    type = "equal"
    value = "active-directory"
    point = {
      path = 1
    }
  }
}

resource "wallarm_rule_mode" "dvwa_mode" {
  mode =  "block"

  action {
    type = "equal"
    value = "dvwa.wallarm-demo.com"
    point = {
      header = "HOST"
    }
  }

  action {
    type = "equal"
    point = {
      method = "GET"
    }
  }
}