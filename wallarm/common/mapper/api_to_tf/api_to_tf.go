package api_to_tf

import (
	"github.com/wallarm/wallarm-go"
)

func Threshold(threshold *wallarm.Threshold) []interface{} {
	if threshold == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"count":  threshold.Count,
			"period": threshold.Period,
		},
	}
}

func Reaction(reaction *wallarm.Reaction) []interface{} {
	if reaction == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"block_by_session": reaction.BlockBySession,
			"block_by_ip":      reaction.BlockByIP,
			"graylist_by_ip":   reaction.GraylistByIP,
		},
	}
}

func EnumeratedParameters(enumeratedParameters *wallarm.EnumeratedParameters) []interface{} {
	if enumeratedParameters == nil {
		return nil
	}

	result := map[string]interface{}{
		"mode": enumeratedParameters.Mode,
	}
	switch enumeratedParameters.Mode {
	case "exact":
		// TODO
	default:
		result["name_regexps"] = enumeratedParameters.NameRegexps
		result["value_regexps"] = enumeratedParameters.ValueRegexp
		result["plain_parameters"] = enumeratedParameters.PlainParameters
		result["additional_parameters"] = enumeratedParameters.AdditionalParameters
	}

	return []interface{}{result}

}

func ArbitraryConditions(arbitraryConditions []wallarm.ArbitraryConditionResp) []interface{} {
	if arbitraryConditions == nil {
		return nil
	}

	result := make([]interface{}, 0, len(arbitraryConditions))
	for _, arbitraryCondition := range arbitraryConditions {
		result = append(result, map[string]interface{}{
			"point":    []interface{}{arbitraryCondition.Point},
			"operator": arbitraryCondition.Operator,
			"value":    arbitraryCondition.Value,
		})
	}

	return result
}
