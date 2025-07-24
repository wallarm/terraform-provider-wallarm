package wallarm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmIgnoreRegex() *schema.Resource {
	fields := map[string]*schema.Schema{
		"regex_id": {
			Type:     schema.TypeInt,
			Required: true,
			ForceNew: true,
		},

		"action": defaultResourceRuleActionSchema,

		"point": {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
	return &schema.Resource{
		Create: resourceWallarmIgnoreRegexCreate,
		Read:   resourceWallarmIgnoreRegexRead,
		Delete: resourceWallarmIgnoreRegexDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmIgnoreRegexImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmIgnoreRegexCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	regexID := d.Get("regex_id").(int)

	ps := d.Get("point").([]interface{})
	d.Set("point", ps)

	points, err := expandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := common.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	vp := &wallarm.ActionCreate{
		Type:                "disable_regex",
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		RegexID:             regexID,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
		Mitigation:          fields.Mitigation,
	}

	actionResp, err := client.HintCreate(vp)
	if err != nil {
		return err
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)
	d.Set("regex_id", actionResp.Body.RegexID)

	resID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionResp.Body.ActionID, actionResp.Body.ID, actionResp.Body.Type)
	d.SetId(resID)

	return resourceWallarmIgnoreRegexRead(d, m)
}

func resourceWallarmIgnoreRegexRead(d *schema.ResourceData, m interface{}) error {
	return common.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API),
		common.ReadOptionWithPoint, common.ReadOptionWithRegexID)
}

func resourceWallarmIgnoreRegexDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"disable_regex"},
			Clientid: []int{clientID},
			ID:       []int{actionID},
		},
		Limit:  1000,
		Offset: 0,
	}
	respRules, err := client.RuleRead(rule)
	if err != nil {
		return err
	}

	if len(respRules.Body) == 1 && respRules.Body[0].Hints == 1 && respRules.Body[0].GroupedHintsCount == 1 {
		if err := client.RuleDelete(actionID); err != nil {
			return err
		}
	} else {
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
	}

	return nil
}

func resourceWallarmIgnoreRegexImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(wallarm.API)
	idAttr := strings.SplitN(d.Id(), "/", 3)
	if len(idAttr) == 3 {
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
		d.Set("action_id", actionID)
		d.Set("rule_id", ruleID)
		d.Set("rule_type", "disable_regex")

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{"disable_regex"},
			},
		}
		actionHints, err := client.HintRead(hint)
		if err != nil {
			return nil, err
		}
		actionsSet := schema.Set{
			F: hashResponseActionDetails,
		}
		if len((*actionHints.Body)) != 0 && len((*actionHints.Body)[0].Action) != 0 {
			for _, a := range (*actionHints.Body)[0].Action {
				acts, err := actionDetailsToMap(a)
				if err != nil {
					return nil, err
				}
				actionsSet.Add(acts)
			}
			d.Set("action", &actionsSet)
		}

		pointInterface := (*actionHints.Body)[0].Point
		point := wrapPointElements(pointInterface)
		d.Set("point", point)

		d.Set("regex_id", (*actionHints.Body)[0].RegexID)

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	if err := resourceWallarmIgnoreRegexRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
