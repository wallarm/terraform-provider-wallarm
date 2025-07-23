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
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmMode() *schema.Resource {
	fields := map[string]*schema.Schema{
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"default", "off", "monitoring", "block", "safe_blocking"}, false),
			ForceNew:     true,
		},

		"action": defaultResourceRuleActionSchema,
	}
	return &schema.Resource{
		Create: resourceWallarmModeCreate,
		Read:   resourceWallarmModeRead,
		Delete: resourceWallarmModeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmModeImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmModeCreate(d *schema.ResourceData, m interface{}) error {
	if d.IsNewResource() {
		existingID, exists, err := existsAction(d, m, "wallarm_mode")
		if err != nil {
			return err
		}
		if exists {
			return ImportAsExistsError("wallarm_rule_mode", existingID)
		}
	}
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	actionsFromState := d.Get("action").(*schema.Set)
	mode := d.Get("mode").(string)

	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}
	wm := &wallarm.ActionCreate{
		Type:                "wallarm_mode",
		Clientid:            clientID,
		Action:              &action,
		Mode:                mode,
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

	resID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionResp.Body.ActionID, actionResp.Body.ID, actionResp.Body.Mode)
	d.SetId(resID)

	return resourceWallarmModeRead(d, m)
}

func resourceWallarmModeRead(d *schema.ResourceData, m interface{}) error {
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
			ActionID: []int{actionID},
			Type:     []string{"wallarm_mode"},
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
		Type:     "wallarm_mode",
		Action:   action,
	}

	var updatedRule *wallarm.ActionBody
	for _, rule := range *actionHints.Body {
		if ruleID == rule.ID {
			updatedRule = &rule
			break
		}

		actualRule := &wallarm.ActionBody{
			ActionID: rule.ActionID,
			Type:     rule.Type,
			Action:   rule.Action,
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

func resourceWallarmModeDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"wallarm_mode"},
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
		if err := client.HintDelete(h); err != nil {
			return err
		}
	}

	return nil
}

func resourceWallarmModeImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
		mode := idAttr[3]
		d.Set("action_id", actionID)
		d.Set("rule_id", ruleID)
		d.Set("mode", mode)
		d.Set("rule_type", "wallarm_mode")

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{"wallarm_mode"},
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

		existingID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, ruleID, mode)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}/{wallarm-mode}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
