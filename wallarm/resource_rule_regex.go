package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resource_rule"
	"github.com/wallarm/wallarm-go"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmRegex() *schema.Resource {
	fields := map[string]*schema.Schema{
		"regex_id": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"attack_type": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"action": defaultResourceRuleActionSchema,

		"regex": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"point": {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
		},

		"experimental": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
			ForceNew: true,
		},
	}
	return &schema.Resource{
		Create: resourceWallarmRegexCreate,
		Read:   resourceWallarmRegexRead,
		Delete: resourceWallarmRegexDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmRegexImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmRegexCreate(d *schema.ResourceData, m interface{}) error {
	experimental := d.Get("experimental").(bool)
	var actionType string
	if experimental {
		actionType = experimentalRegex
	} else {
		actionType = "regex"
	}

	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	regex := d.Get("regex").(string)
	attackType := d.Get("attack_type").(string)

	ps := d.Get("point").([]interface{})
	d.Set("point", ps)

	points, err := expandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resource_rule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	rx := &wallarm.ActionCreate{
		Type:                actionType,
		AttackType:          attackType,
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Regex:               regex,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
		Mitigation:          fields.Mitigation,
	}
	regexResp, err := client.HintCreate(rx)
	if err != nil {
		return err
	}

	d.Set("regex_id", regexResp.Body.RegexID.(float64))
	d.Set("rule_id", regexResp.Body.ID)
	d.Set("action_id", regexResp.Body.ActionID)
	d.Set("rule_type", regexResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d/%s", clientID, regexResp.Body.ActionID, regexResp.Body.ID, regexResp.Body.Type)
	d.SetId(resID)

	return resourceWallarmRegexRead(d, m)
}

func resourceWallarmRegexRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)
	ruleID := d.Get("rule_id").(int)
	regex := d.Get("regex").(string)
	regexID := d.Get("regex_id").(float64)
	attackType := d.Get("attack_type").(string)

	experimental := d.Get("experimental").(bool)
	var actionType string
	if experimental {
		actionType = "experimental_regex"
	} else {
		actionType = "regex"
	}

	var ps []interface{}
	if v, ok := d.GetOk("point"); ok {
		ps = v.([]interface{})
	} else {
		return nil
	}
	var points []interface{}
	for _, point := range ps {
		p := point.([]interface{})
		points = append(points, p...)
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resource_rule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	actsSlice := make([]interface{}, 0, len(action))
	for _, a := range action {
		acts, err := actionDetailsToMap(a)
		if err != nil {
			return err
		}
		actsSlice = append(actsSlice, acts)
	}

	actionsSet := schema.NewSet(hashResponseActionDetails, actsSlice)

	hint := &wallarm.HintRead{
		Limit:     1000,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ID:       []int{ruleID},
		},
	}
	actionHints, err := client.HintRead(hint)
	if err != nil {
		return err
	}

	// This is mandatory to fill in the default values in order to compare them deeply.
	// Assign new values to the old struct slice.
	fillInDefaultValues(&action)

	expectedRule := wallarm.ActionBody{
		ActionID:   actionID,
		Type:       actionType,
		AttackType: attackType,
		RegexID:    regexID,
		Regex:      regex,
		Point:      points,
	}

	var updatedRule *wallarm.ActionBody
	for _, rule := range *actionHints.Body {
		if ruleID == rule.ID {
			updatedRule = &rule
			break
		}

		// The response has a different structure so we have to align them
		// to uniform view then to compare.
		alignedPoints := alignPointScheme(rule.Point)

		actualRule := &wallarm.ActionBody{
			ActionID:   rule.ActionID,
			Type:       rule.Type,
			AttackType: rule.AttackType,
			RegexID:    rule.RegexID,
			Regex:      rule.Regex,
			Point:      alignedPoints,
		}

		if cmp.Equal(expectedRule, *actualRule) && equalWithoutOrder(action, rule.Action) {
			updatedRule = &rule
			break
		}
	}

	if updatedRule == nil {
		d.SetId("")
		return nil
	}

	d.Set("rule_id", updatedRule.ID)
	d.Set("client_id", clientID)
	d.Set("active", updatedRule.Active)
	d.Set("title", updatedRule.Title)
	d.Set("mitigation", updatedRule.Mitigation)
	d.Set("set", updatedRule.Set)

	if actionsSet.Len() != 0 {
		d.Set("action", &actionsSet)
	} else {
		log.Printf("[WARN] action was empty so it either doesn't exist or it is a default branch which has no conditions. Actions: %v", &actionsSet)
	}

	return nil
}

func resourceWallarmRegexDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	ruleID := d.Get("rule_id").(int)
	h := &wallarm.HintDelete{
		Filter: &wallarm.HintDeleteFilter{
			Clientid: []int{clientID},
			ID:       ruleID,
		},
	}

	if err := client.HintDelete(h); err != nil {
		return err
	}

	return nil
}

func resourceWallarmRegexImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(wallarm.API)
	idAttr := strings.SplitN(d.Id(), "/", 4)
	if len(idAttr) == 4 {
		clientID, err := strconv.Atoi(idAttr[0])
		if err != nil {
			return nil, err
		}
		actionID, err := strconv.Atoi(idAttr[1])
		if err != nil {
			return nil, err
		}
		ruleID, err := strconv.Atoi(idAttr[2])
		if err != nil {
			return nil, err
		}
		hintType := idAttr[3]
		d.Set("action_id", actionID)
		d.Set("rule_id", ruleID)
		d.Set("rule_type", hintType)

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{hintType},
			},
		}
		actionHints, err := client.HintRead(hint)
		if err != nil {
			return nil, err
		}
		actionsSet := schema.Set{
			F: hashResponseActionDetails,
		}
		if len(*actionHints.Body) != 0 && len((*actionHints.Body)[0].Action) != 0 {
			for _, a := range (*actionHints.Body)[0].Action {
				acts, err := actionDetailsToMap(a)
				if err != nil {
					return nil, err
				}
				actionsSet.Add(acts)
			}
			d.Set("action", &actionsSet)
			d.Set("regex_id", (*actionHints.Body)[0].RegexID)
			d.Set("regex", (*actionHints.Body)[0].Regex)
			d.Set("attack_type", (*actionHints.Body)[0].AttackType)

			pointInterface := (*actionHints.Body)[0].Point
			point := wrapPointElements(pointInterface)
			d.Set("point", point)
		}

		if hintType == "experimental_regex" {
			d.Set("experimental", true)
		} else {
			d.Set("experimental", false)
		}

		existingID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, ruleID, hintType)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}/{regex/experimental_regex}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
