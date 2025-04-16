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

func resourceWallarmBruteForceCounter() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmBruteForceCounterCreate,
		Read:   resourceWallarmBruteForceCounterRead,
		Delete: resourceWallarmBruteForceCounterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmBruteForceCounterImport,
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

			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Client ID to perform changes",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v <= 0 {
						errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
					}
					return
				},
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"counter": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"action": defaultResourceRuleActionSchema,
		},
	}
}

func resourceWallarmBruteForceCounterCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)
	actionsFromState := d.Get("action").(*schema.Set)

	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}
	wm := &wallarm.ActionCreate{
		Type:                "brute_counter",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		Comment:             comment,
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

	return resourceWallarmBruteForceCounterRead(d, m)
}

func resourceWallarmBruteForceCounterRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)
	ruleID := d.Get("rule_id").(int)
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
			ActionID: []int{actionID},
			Type:     []string{"brute_counter"},
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
		Type:     "brute_counter",
		Action:   action,
	}

	var notFoundRules []int
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

	if err = d.Set("rule_id", updatedRuleID); err != nil {
		return err
	}

	if err = d.Set("counter", updatedCounter); err != nil {
		return err
	}

	if err = d.Set("client_id", clientID); err != nil {
		return err
	}

	if actionsSet.Len() != 0 {
		if err := d.Set("action", &actionsSet); err != nil {
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

func resourceWallarmBruteForceCounterDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"brute_counter"},
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

func resourceWallarmBruteForceCounterImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
		if err = d.Set("action_id", actionID); err != nil {
			return nil, err
		}
		if err = d.Set("rule_id", ruleID); err != nil {
			return nil, err
		}
		if err = d.Set("rule_type", "brute_counter"); err != nil {
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
				Type:     []string{"brute_counter"},
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

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	if err := resourceWallarmBruteForceCounterRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
