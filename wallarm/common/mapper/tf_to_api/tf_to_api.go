package tf_to_api

import (
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
		PlainParameters:      enumeratedParameter["plain_parameters"].(bool),
		AdditionalParameters: enumeratedParameter["additional_parameters"].(bool),
	}
}

func mapEnumeratedParameterExactToAPI(enumeratedParameter map[string]interface{}) *wallarm.EnumeratedParameters {
	return nil
	//return &wallarm.EnumeratedParameters{
	//	Mode:                 "exact",
	//	Points:
	//}
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
	response := make([]wallarm.AdvancedCondition, 0, len(advancedConditions))
	if len(advancedConditions) == 0 {
		return response
	}

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
	response := make([]wallarm.ArbitraryConditionReq, 0, len(arbitraryConditions))
	if len(arbitraryConditions) == 0 {
		return response
	}

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

func mapPointToAPI(point []interface{}) wallarm.TwoDimensionalSlice {
	response := make(wallarm.TwoDimensionalSlice, 0, len(point))
	if len(point) == 0 {
		return response
	}

	for _, p1 := range point {
		p1ToResp := make([]interface{}, 0, len(p1.([]interface{})))
		for _, p2 := range p1.([]interface{}) {
			p1ToResp = append(p1ToResp, p2)
		}
		response = append(response, p1ToResp)
	}

	return response
}
