package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/wallarm-go"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// nolint:dupl
func resourceWallarmBolaCounter() *schema.Resource {
	fields := map[string]*schema.Schema{
		"counter": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"action": defaultResourceRuleActionSchema,
	}

	return &schema.Resource{
		Create: resourceWallarmBolaCounterCreate,
		Read:   resourceWallarmBolaCounterRead,
		Delete: resourceWallarmBolaCounterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmBolaCounterImport,
		},

		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmBolaCounterCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	actionsFromState := d.Get("action").(*schema.Set)

	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}
	wm := &wallarm.ActionCreate{
		Type:                "bola_counter",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
		Mitigation:          fields.Mitigation,
	}

	actionResp, err := client.HintCreate(wm)
	if err != nil {
		return err
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmBolaCounterRead(d, m)
}

func resourceWallarmBolaCounterRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)
	ruleID := d.Get("rule_id").(int)
	actionsFromState := d.Get("action").(*schema.Set)

	action, err := expandSetToActionDetailsList(actionsFromState)
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
			Type:     []string{"bola_counter"},
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
		ActionID: actionID,
		Type:     "bola_counter",
		Action:   action,
	}

	notFoundRules := make([]int, 0)
	var updatedRuleID int
	var updatedCounter string
	for _, rule := range *actionHints.Body {
		actualRule := &wallarm.ActionBody{
			ActionID: rule.ActionID,
			Type:     rule.Type,
			Action:   rule.Action,
		}

		if ruleID == rule.ID || cmp.Equal(expectedRule, *actualRule) && equalWithoutOrder(action, rule.Action) {
			updatedRuleID = rule.ID
			updatedCounter = rule.Counter
			break
		}

		notFoundRules = append(notFoundRules, rule.ID)
	}

	d.Set("rule_id", updatedRuleID)
	d.Set("counter", updatedCounter)
	d.Set("client_id", clientID)

	if actionsSet.Len() != 0 {
		d.Set("action", &actionsSet)
	} else {
		log.Printf("[WARN] action was empty so it either doesn't exist or it is a default branch which has no conditions. Actions: %v", &actionsSet)
	}

	if updatedRuleID == 0 {
		log.Printf("[WARN] these rule IDs: %v have been found under the action ID: %d. But it isn't in the Terraform Plan.", notFoundRules, actionID)
		d.SetId("")
	}

	return nil
}

func resourceWallarmBolaCounterDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"bola_counter"},
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

func resourceWallarmBolaCounterImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(wallarm.API)
	idAttr := strings.SplitN(d.Id(), "/", 3)
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
		d.Set("action_id", actionID)
		d.Set("rule_id", ruleID)
		d.Set("rule_type", "bola_counter")

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{"bola_counter"},
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

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	if err := resourceWallarmBolaCounterRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
