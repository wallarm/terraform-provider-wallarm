resource "wallarm_rule_bruteforce_counter" "root_counter" {
	action {
		type = "iequal"
		value = "/"
		point = {
			path = 0
		}
	}
}
