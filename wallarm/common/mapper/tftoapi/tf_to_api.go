package tftoapi

import (
	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/wallarm-go"
)

func EnumeratedParameters(enumeratedParameters []interface{}) *wallarm.EnumeratedParameters {
	if len(enumeratedParameters) == 0 {
		return nil
	}
	enumeratedParameterObj := enumeratedParameters[0].(map[string]interface{})
	mode := enumeratedParameterObj["mode"].(string)
	switch mode {
	case "regexp":
		return mapEnumeratedParameterRegexpToAPI(enumeratedParameterObj)
	case "exact":
		return mapEnumeratedParameterExactToAPI(enumeratedParameterObj)
	default:
		return mapEnumeratedParameterRegexpToAPI(enumeratedParameterObj)
	}
}

func mapEnumeratedParameterRegexpToAPI(enumeratedParameter map[string]interface{}) *wallarm.EnumeratedParameters {
	return &wallarm.EnumeratedParameters{
		Mode:                 "regexp",
		NameRegexps:          common.ConvertToStringSlice(enumeratedParameter["name_regexps"].([]interface{})),
		ValueRegexp:          common.ConvertToStringSlice(enumeratedParameter["value_regexps"].([]interface{})),
		PlainParameters:      lo.ToPtr(enumeratedParameter["plain_parameters"].(bool)),
		AdditionalParameters: lo.ToPtr(enumeratedParameter["additional_parameters"].(bool)),
	}
}

func mapEnumeratedParameterExactToAPI(enumeratedParameter map[string]interface{}) *wallarm.EnumeratedParameters {
	result := &wallarm.EnumeratedParameters{
		Mode: "exact",
	}

	pointsList, ok := enumeratedParameter["points"].([]interface{})
	if !ok || len(pointsList) == 0 {
		return result
	}

	points := make([]*wallarm.Points, 0, len(pointsList))
	for _, item := range pointsList {
		pointsObj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		point, _ := pointsObj["point"].([]interface{})
		sensitive, _ := pointsObj["sensitive"].(bool)

		points = append(points, &wallarm.Points{
			Point:     point,
			Sensitive: sensitive,
		})
	}

	result.Points = points
	return result
}

func Reaction(reaction []interface{}) *wallarm.Reaction {
	if len(reaction) == 0 {
		return nil
	}

	reactionObj := reaction[0].(map[string]interface{})

	var blockBySession, blockByIP, graylistByIP *int
	if bbs := reactionObj["block_by_session"].(int); bbs != 0 {
		blockBySession = &bbs
	}
	if bbip := reactionObj["block_by_ip"].(int); bbip != 0 {
		blockByIP = &bbip
	}
	if gbip := reactionObj["graylist_by_ip"].(int); gbip != 0 {
		graylistByIP = &gbip
	}

	return &wallarm.Reaction{
		BlockBySession: blockBySession,
		BlockByIP:      blockByIP,
		GraylistByIP:   graylistByIP,
	}
}

func Threshold(threshold []interface{}) *wallarm.Threshold {
	if len(threshold) == 0 {
		return nil
	}

	thresholdObj := threshold[0].(map[string]interface{})

	return &wallarm.Threshold{
		Count:  thresholdObj["count"].(int),
		Period: thresholdObj["period"].(int),
	}
}

func AdvancedConditions(advancedConditions []interface{}) []wallarm.AdvancedCondition {
	if len(advancedConditions) == 0 {
		return nil
	}

	response := make([]wallarm.AdvancedCondition, 0, len(advancedConditions))
	for _, advancedCondition := range advancedConditions {
		advancedConditionObj := advancedCondition.(map[string]interface{})
		response = append(response, wallarm.AdvancedCondition{
			Field:    advancedConditionObj["field"].(string),
			Value:    common.ConvertToStringSlice(advancedConditionObj["value"].([]interface{})),
			Operator: advancedConditionObj["operator"].(string),
		})
	}

	return response
}

func ArbitraryConditionsReq(arbitraryConditions []interface{}) []wallarm.ArbitraryConditionReq {
	if len(arbitraryConditions) == 0 {
		return nil
	}

	response := make([]wallarm.ArbitraryConditionReq, 0, len(arbitraryConditions))
	for _, arbitraryCondition := range arbitraryConditions {
		arbitraryConditionObj := arbitraryCondition.(map[string]interface{})
		response = append(response, wallarm.ArbitraryConditionReq{
			Point:    mapPointToAPI(arbitraryConditionObj["point"].([]interface{})),
			Value:    common.ConvertToStringSlice(arbitraryConditionObj["value"].([]interface{})),
			Operator: arbitraryConditionObj["operator"].(string),
		})
	}

	return response
}

// nolint
func mapPointToAPI(point []interface{}) wallarm.TwoDimensionalSlice {
	response := make(wallarm.TwoDimensionalSlice, 0, len(point))
	if len(point) == 0 {
		return response
	}

	for _, p1 := range point {
		p1ToResp := make([]interface{}, 0, len(p1.([]interface{})))
		p1ToResp = append(p1ToResp, p1.([]interface{})...)
		response = append(response, p1ToResp)
	}

	return response
}
