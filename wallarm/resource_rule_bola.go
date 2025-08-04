package wallarm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/samber/lo"
)

// nolint:dupl
func resourceWallarmBola() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action":    defaultResourceRuleActionSchema,
		"threshold": thresholdSchema,
		"reaction":  reactionSchema,
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"monitoring", "block"}, false),
			ForceNew:     true,
		},
		"enumerated_parameters": enumeratedParametersSchema,
		"advanced_conditions":   advancedConditionsSchema,
		"arbitrary_conditions":  arbitraryConditionsSchema,
	}
	sh := lo.Assign(fields, commonResourceRuleFields)

	return &schema.Resource{
		Create: resourceWallarmBolaCreate,
		Read:   resourceWallarmBolaRead,
		Delete: resourceWallarmBolaDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmBolaImport,
		},
		Schema: sh,
	}
}

func resourceWallarmBolaCreate(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmCreate(d, m.(wallarm.API), retrieveClientID(d),
		"bola", "bola", resourceWallarmBolaRead)
}

func resourceWallarmBolaRead(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API),
		common.ReadOptionWithMode,
		common.ReadOptionWithAction,
		common.ReadOptionWithThreshold,
		common.ReadOptionWithReaction,
		common.ReadOptionWithEnumeratedParameters,
		common.ReadOptionWithArbitraryConditions,
	)
}

func resourceWallarmBolaDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"bola"},
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
		if err = client.RuleDelete(actionID); err != nil {
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

		if err = client.HintDelete(h); err != nil {
			return err
		}
	}
	return nil
}

func resourceWallarmBolaImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "bola")

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{"bola"},
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
		}

		pointInterface := (*actionHints.Body)[0].Point
		point := wrapPointElements(pointInterface)
		d.Set("point", point)

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
