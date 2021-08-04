package wallarm

import (
	"fmt"
	"log"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmSetResponseHeader() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmSetResponseHeaderCreate,
		Read:   resourceWallarmSetResponseHeaderRead,
		Update: resourceWallarmSetResponseHeaderUpdate,
		Delete: resourceWallarmSetResponseHeaderDelete,

		Schema: map[string]*schema.Schema{

			"rule_id": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
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
				ValidateFunc: validation.StringInSlice([]string{"append", "replace"}, false),
			},

			"headers": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
											if v < 0 || v > 2048 {
												errs = append(errs, fmt.Errorf("%q must be between 0 and 2048 inclusive, got: %d", key, v))
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

func resourceWallarmSetResponseHeaderCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)
	mode := d.Get("mode").(string)
	headers := d.Get("headers").(map[string]interface{})

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	var ruleIDs []int
	for k, v := range headers {
		vp := &wallarm.ActionCreate{
			Type:      "set_response_header",
			Clientid:  clientID,
			Action:    &action,
			Mode:      mode,
			Name:      k,
			Values:    []string{v.(string)},
			Validated: false,
			Comment:   comment,
		}
		actionResp, err := client.HintCreate(vp)
		if err != nil {
			return err
		}

		actionID := actionResp.Body.ActionID
		d.Set("action_id", actionID)

		ruleIDs = append(ruleIDs, actionResp.Body.ID)

		resID := fmt.Sprintf("%d/%d/%d", clientID, actionID, actionResp.Body.ID)
		d.SetId(resID)

	}

	d.Set("rule_id", ruleIDs)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmSetResponseHeaderRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)
	mode := d.Get("mode").(string)
	headers := d.Get("headers").(map[string]interface{})

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	var ruleIDInterface []interface{}
	if v, ok := d.GetOk("rule_id"); ok {
		ruleIDInterface = v.([]interface{})
	} else {
		return nil
	}
	ruleIDs := expandInterfaceToIntList(ruleIDInterface)

	hint := &wallarm.HintRead{
		Limit:     1000,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ActionID: []int{actionID},
		},
	}
	actionHints, err := client.HintRead(hint)
	if err != nil {
		return err
	}

	// This is mandatory to fill in the default values in order to compare them deeply.
	// Assign new values to the old struct slice.
	fillInDefaultValues(&action)

	var notFoundRules []int
	var updatedRuleIDs []int
	for _, rule := range *actionHints.Body {

		// Check out by ID. The specific rule should be found.
		if wallarm.Contains(ruleIDs, rule.ID) {
			updatedRuleIDs = append(updatedRuleIDs, rule.ID)
			continue
		}

		for name, value := range headers {

			expectedRule := wallarm.ActionBody{
				ActionID: actionID,
				Type:     "set_response_header",
				Action:   action,
				Mode:     mode,
				Name:     name,
				Values:   []interface{}{value},
			}

			actualRule := &wallarm.ActionBody{
				ActionID: rule.ActionID,
				Type:     rule.Type,
				Action:   rule.Action,
				Mode:     rule.Mode,
				Name:     rule.Name,
				Values:   rule.Values,
			}

			if cmp.Equal(expectedRule, *actualRule) && equalWithoutOrder(action, rule.Action) {
				updatedRuleIDs = append(updatedRuleIDs, rule.ID)
				continue
			}
		}

		notFoundRules = append(notFoundRules, rule.ID)
	}

	if err := d.Set("rule_id", updatedRuleIDs); err != nil {
		return err
	}

	d.Set("client_id", clientID)

	if len(updatedRuleIDs) == 0 {
		log.Printf("[WARN] these rule IDs: %v have been found under the action ID: %d. But it isn't in the Terraform Plan.", notFoundRules, actionID)
		d.SetId("")
	}

	return nil
}

func resourceWallarmSetResponseHeaderDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)

	var ruleIDInterface []interface{}
	if v, ok := d.GetOk("rule_id"); ok {
		ruleIDInterface = v.([]interface{})
	} else {
		return nil
	}
	ruleIDs := expandInterfaceToIntList(ruleIDInterface)

	for _, ruleID := range ruleIDs {
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

func resourceWallarmSetResponseHeaderUpdate(d *schema.ResourceData, m interface{}) error {
	if err := resourceWallarmSetResponseHeaderDelete(d, m); err != nil {
		return err
	}
	return resourceWallarmSetResponseHeaderCreate(d, m)
}
