resource "wallarm_rule_dirbust_counter" "root_counter" {
	counter = "d:login"
	
	action {
    	type = "iequal"
    	point = {
      		action_name = "login"
    	}
  	}
}