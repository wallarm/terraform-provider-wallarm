package apitotf

import (
	"fmt"

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
		result["points"] = mapPointsToTF(enumeratedParameters.Points)
	default:
		result["name_regexps"] = enumeratedParameters.NameRegexps
		result["value_regexps"] = enumeratedParameters.ValueRegexp
		result["plain_parameters"] = enumeratedParameters.PlainParameters
		result["additional_parameters"] = enumeratedParameters.AdditionalParameters
	}

	return []interface{}{result}
}

func mapPointsToTF(points *wallarm.Points) []interface{} {
	if points == nil {
		return nil
	}

	point := make([]interface{}, 0, len(points.Point))
	for _, p := range points.Point {
		point = append(point, p)
	}

	return []interface{}{
		map[string]interface{}{
			"point":     point,
			"sensitive": points.Sensitive,
		},
	}
}

func AdvancedConditions(advancedConditions []wallarm.AdvancedCondition) []interface{} {
	if advancedConditions == nil {
		return nil
	}

	result := make([]interface{}, 0, len(advancedConditions))
	for _, advancedCondition := range advancedConditions {
		result = append(result, map[string]interface{}{
			"field":    advancedCondition.Field,
			"operator": advancedCondition.Operator,
			"value":    advancedCondition.Value,
		})
	}

	return result
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

func SliceAnyToSliceString(in []any) []string {
	if in == nil {
		return nil
	}

	result := make([]string, 0, len(in))
	for _, el := range in {
		result = append(result, fmt.Sprintf("%s", el))
	}

	return result
}
