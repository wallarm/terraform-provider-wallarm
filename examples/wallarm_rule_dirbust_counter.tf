resource "wallarm_rule_dirbust_counter" "login_counter" {
	action {
    	type = "iequal"
    	point = {
      		action_name = "login"
    	}
  	}
}
