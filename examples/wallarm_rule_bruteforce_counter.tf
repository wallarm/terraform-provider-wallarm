resource "wallarm_rule_bruteforce_counter" "root_counter" {
	counter = "b:root"
	action {
		type = "iequal"
		value = "/"
		point = {
			path = 0
		}
	}
}