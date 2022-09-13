package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmMode() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmModeCreate,
		Read:   resourceWallarmModeRead,
		Delete: resourceWallarmModeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmModeImport,
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

			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"default", "off", "monitoring", "block"}, false),
				ForceNew:     true,
			},

			"action": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"equal", "iequal", "regex", "absent"}, false),
							ForceNew:     true,
						},

						"value": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},

						"point": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"method": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{"GET", "HEAD", "POST",
											"PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}, false),
									},

									"path": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
											v := val.(int)
											if v < 0 || v > 60 {
												errs = append(errs, fmt.Errorf("%q must be between 0 and 60 inclusive, got: %d", key, v))
											}
											return
										},
									},

									"action_name": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},

									"action_ext": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},

									"proto": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"1.0", "1.1", "2.0", "3.0"}, false),
									},

									"scheme": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"http", "https"}, true),
									},

									"uri": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},

									"instance": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
											v := val.(int)
											if v < -1 {
												errs = append(errs, fmt.Errorf("%q must be be greater than -1 inclusive, got: %d", key, v))
											}
											return
										},
									},
								},
							},
						},
					},
				},
			},
		},
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
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)
	actionsFromState := d.Get("action").(*schema.Set)
	mode := d.Get("mode").(string)

	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}
	wm := &wallarm.ActionCreate{
		Type:      "wallarm_mode",
		Clientid:  clientID,
		Action:    &action,
		Mode:      mode,
		Validated: false,
		Comment:   comment,
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

	var notFoundRules []int
	var updatedRuleID int
	for _, rule := range *actionHints.Body {
		if ruleID == rule.ID {
			updatedRuleID = rule.ID
			continue
		}

		actualRule := &wallarm.ActionBody{
			ActionID: rule.ActionID,
			Type:     rule.Type,
			Action:   rule.Action,
		}

		if cmp.Equal(expectedRule, *actualRule) && equalWithoutOrder(action, rule.Action) {
			updatedRuleID = rule.ID
			continue
		}

		notFoundRules = append(notFoundRules, rule.ID)
	}

	if err := d.Set("rule_id", updatedRuleID); err != nil {
		return err
	}

	d.Set("client_id", clientID)

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

func resourceWallarmModeDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintsCount: wallarm.TwoDimensionalSlice{{1, nil}},
			HintType:   []string{"wallarm_mode"},
			Clientid:   []int{clientID},
			ID:         []int{actionID},
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
		var actsSlice []map[string]interface{}
		if len((*actionHints.Body)) != 0 && len((*actionHints.Body)[0].Action) != 0 {
			for _, a := range (*actionHints.Body)[0].Action {
				acts, err := actionDetailsToMap(a)
				if err != nil {
					return nil, err
				}
				actsSlice = append(actsSlice, acts)
				actionsSet.Add(acts)
			}
			if err := d.Set("action", &actionsSet); err != nil {
				return nil, err
			}
		}

		existingID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, ruleID, mode)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}/{wallarm-mode}\"", d.Id())
	}

	resourceWallarmModeRead(d, m)

	return []*schema.ResourceData{d}, nil
}
