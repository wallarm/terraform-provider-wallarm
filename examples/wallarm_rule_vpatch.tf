resource "wallarm_rule_vpatch" "default" {
  attack_type =  ["sqli"]
  point = [["get_all"]]
}

resource "wallarm_rule_vpatch" "vpatch" {
  attack_type =  ["redir"]
  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
  action {
    type = "equal"
    value = "api"
    point = {
      path = 0
    }
  }
  action {
    type = "regex"
    value = "logon"
    point = {
      path = 1
    }
  }
  action {
    type = "equal"
    point = {
      method = "POST"
    }
  }
  action {
    point = {
      instance = "1"
    }
  }
  action {
    type = "regex"
    point = {
      scheme = "https"
    }
  }
  point = [["post"],["xml"],["hash","user"]]
}

resource "wallarm_rule_vpatch" "splunk" {
  attack_type =  ["sqli", "nosqli"]
  action {
    type = "iequal"
    value = "splunk.wallarm-demo.com:88"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
}

resource "wallarm_rule_vpatch" "tiredful_api" {
  attack_type =  ["any"]
  action {
    point = {
      instance = "9"
    }
  }
  action {
    type = "absent"
    point = {
      path = 0
    }
  }

  action {
    type = "equal"
    point = {
      action_name = "formmail"
    }
  }

  action {
    type = "equal"
    point = {
      action_ext = "cgi"
    }
  }

  point = [["uri"]]
}

resource "wallarm_rule_vpatch" "env_sample" {
  attack_type =  ["any"]

  action {
    type = "equal"
    point = {
      action_name = ".env.sample"
    }
  }

  action {
    type = "equal"
    point = {
      action_ext = "php"
    }
  }

  point = [["uri"]]
}