package wallarm

// Contains wraps methods (for string and int) to check if List contains the element.
func Contains(a interface{}, x interface{}) bool {
	switch x.(type) {
	case int:
		group := a.([]int)
		return intInList(group, x.(int))
	case string:
		group := a.([]string)
		return strInList(group, x.(string))
	default:
		return false
	}
}

func strInList(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func intInList(a []int, x int) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
