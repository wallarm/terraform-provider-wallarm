package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmDisableAttackType() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmDisableAttackTypeCreate,
		Read:   resourceWallarmDisableAttackTypeRead,
		Delete: resourceWallarmDisableAttackTypeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmDisableAttackTypeImport,
		},

		Schema: map[string]*schema.Schema{

			"rule_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"action_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"rule_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"client_id": defaultClientIDWithValidationSchema,

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"attack_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: `Possible values: "any", "sqli", "rce", "crlf", "nosqli", "ptrav",
				"xxe", "ptrav", "xss", "scanner", "redir", "ldapi", "any", "redir", "mass_assignment", "ssrf"`,
			},

			"action": defaultResourceRuleActionSchema,

			"point": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func resourceWallarmDisableAttackTypeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)
	attackType := d.Get("attack_type").(string)

	ps := d.Get("point").([]interface{})
	if err := d.Set("point", ps); err != nil {
		return err
	}

	points, err := expandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	wm := &wallarm.ActionCreate{
		Type:                "disable_attack_type",
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Validated:           false,
		Comment:             comment,
		AttackType:          attackType,
		VariativityDisabled: true,
	}

	actionResp, err := client.HintCreate(wm)
	if err != nil {
		return err
	}

	if err = d.Set("rule_id", actionResp.Body.ID); err != nil {
		return err
	}
	if err = d.Set("action_id", actionResp.Body.ActionID); err != nil {
		return err
	}
	if err = d.Set("rule_type", actionResp.Body.Type); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmDisableAttackTypeRead(d, m)
}

func resourceWallarmDisableAttackTypeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)
	ruleID := d.Get("rule_id").(int)

	ps := d.Get("point").([]interface{})
	var points []interface{}
	for _, point := range ps {
		p := point.([]interface{})
		points = append(points, p...)
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	var actsSlice []interface{}
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
			Type:     []string{"disable_attack_type"},
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
		Type:     "disable_attack_type",
		Point:    points,
	}

	var notFoundRules []int
	var updatedRuleID int
	for _, rule := range *actionHints.Body {

		if ruleID == rule.ID {
			updatedRuleID = rule.ID
			continue
		}

		// The response has a different structure so we have to align them
		// to uniform view then to compare.
		alignedPoints := alignPointScheme(rule.Point)

		actualRule := &wallarm.ActionBody{
			ActionID: rule.ActionID,
			Type:     rule.Type,
			Point:    alignedPoints,
		}

		if cmp.Equal(expectedRule, *actualRule) && equalWithoutOrder(action, rule.Action) {
			updatedRuleID = rule.ID
			continue
		}

		notFoundRules = append(notFoundRules, rule.ID)
	}

	if err = d.Set("rule_id", updatedRuleID); err != nil {
		return err
	}

	if err = d.Set("client_id", clientID); err != nil {
		return err
	}

	if actionsSet.Len() != 0 {
		if err = d.Set("action", &actionsSet); err != nil {
			return err
		}
	} else {
		log.Printf("[WARN] action was empty so it either doesn't exist or it is a default branch which has no conditions. Actions: %v", &actionsSet)
	}

	if updatedRuleID == 0 {
		log.Printf("[WARN] these rule IDs: %v have been found under the action ID: %d. But it isn't in the Terraform Plan.", notFoundRules, actionID)
		d.SetId("")
	}

	return nil
}

func resourceWallarmDisableAttackTypeDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"disable_attack_type"},
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

func resourceWallarmDisableAttackTypeImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
		if err = d.Set("action_id", actionID); err != nil {
			return nil, err
		}
		if err = d.Set("rule_id", ruleID); err != nil {
			return nil, err
		}
		if err = d.Set("rule_type", "disable_attack_type"); err != nil {
			return nil, err
		}

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{"disable_attack_type"},
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
			if err := d.Set("action", &actionsSet); err != nil {
				return nil, err
			}
		}

		if err = d.Set("attack_type", (*actionHints.Body)[0].AttackType); err != nil {
			return nil, err
		}
		pointInterface := (*actionHints.Body)[0].Point
		point := wrapPointElements(pointInterface)
		if err = d.Set("point", point); err != nil {
			return nil, err
		}

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	if err := resourceWallarmDisableAttackTypeRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
