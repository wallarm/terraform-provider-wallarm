package common

import "fmt"

func ConvertToStringSlice(input []interface{}) []string {
	result := make([]string, 0, len(input))
	for _, v := range input {
		if v == nil {
			continue
		}
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprintf("%v", v)
		}
		result = append(result, s)
	}
	return result
}
