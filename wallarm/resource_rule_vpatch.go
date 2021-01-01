package wallarm

import (
	"fmt"
	"log"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// TODO: Add an importer:
// Importer: &schema.ResourceImporter{
// 	State: resourceWallarmVpatchImport,
// },

func resourceWallarmVpatch() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmVpatchCreate,
		Read:   resourceWallarmVpatchRead,
		Update: resourceWallarmVpatchUpdate,
		Delete: resourceWallarmVpatchDelete,

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

			"attack_type": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Description: `Possible values: "any", "sqli", "rce", "crlf", "nosqli", "ptrav",
				"xxe", "ptrav", "xss", "scanner", "redir", "ldapi"`,
				Elem: &schema.Schema{Type: schema.TypeString},
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
										Computed: true,
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
										Computed: true,
									},

									"action_ext": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Computed: true,
									},

									"proto": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Computed: true,
									},

									"scheme": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice([]string{"http", "https"}, true),
									},

									"uri": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Computed: true,
									},

									"instance": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
											v := val.(int)
											if v < -1 {
												errs = append(errs, fmt.Errorf("%q must be between greater then -1 inclusive, got: %d", key, v))
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
		},
	}
}

func resourceWallarmVpatchCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	attackType := d.Get("attack_type").([]interface{})
	var attacks []string
	for _, attack := range attackType {
		attacks = append(attacks, attack.(string))
	}

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

	ruleIDs := make([]int, len(attacks))
	attackTypes := make([]string, len(attacks))
	for i, attack := range attacks {
		vp := &wallarm.ActionCreate{
			Type:       "vpatch",
			AttackType: attack,
			Clientid:   clientID,
			Action:     &action,
			Point:      points,
			Validated:  false,
		}

		actionResp, err := client.HintCreate(vp)
		if err != nil {
			return err
		}

		if err := d.Set("action_id", actionResp.Body.ActionID); err != nil {
			return err
		}

		if err := d.Set("rule_type", actionResp.Body.Type); err != nil {
			return err
		}

		ruleIDs[i] = actionResp.Body.ID
		attackTypes[i] = actionResp.Body.AttackType

		d.SetId(actionResp.Body.Type)
	}

	if err := d.Set("rule_id", ruleIDs); err != nil {
		return err
	}

	if err := d.Set("attack_type", attackTypes); err != nil {
		return err
	}

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmVpatchRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)

	var attackType []interface{}
	if v, ok := d.GetOk("attack_type"); ok {
		attackType = v.([]interface{})
	} else {
		return nil
	}
	var attacks []string
	for _, attack := range attackType {
		attacks = append(attacks, attack.(string))
	}

	var ruleIDInterface []interface{}
	if v, ok := d.GetOk("rule_id"); ok {
		ruleIDInterface = v.([]interface{})
	} else {
		return nil
	}
	ruleIDs := expandInterfaceToIntList(ruleIDInterface)

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
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

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

	// This is mandatory to fill in the default values in order to compare structs deeply.
	// Assign new values to the old struct slice.
	fillInDefaultValues(&action)

	var expectedRules []wallarm.ActionBody
	for _, attack := range attacks {
		r := wallarm.ActionBody{
			ActionID:   actionID,
			Type:       "vpatch",
			Point:      points,
			AttackType: attack,
		}
		expectedRules = append(expectedRules, r)
	}

	var notFoundRules []int
	var updatedRuleIDs []int
out:
	for _, rule := range *actionHints.Body {

		// Check straight right by ID. The specific rule should be found.
		if wallarm.Contains(ruleIDs, rule.ID) {
			updatedRuleIDs = append(updatedRuleIDs, rule.ID)
			continue
		}

		// The response has a different structure so we have to align them
		// to uniform view then to compare.
		alignedPoints := alignPointScheme(rule.Point)

		actualRule := &wallarm.ActionBody{
			ActionID:   rule.ActionID,
			Type:       rule.Type,
			Point:      alignedPoints,
			AttackType: rule.AttackType,
		}

		for _, expectedRule := range expectedRules {
			if cmp.Equal(expectedRule, *actualRule) && equalWithoutOrder(action, rule.Action) {
				updatedRuleIDs = append(updatedRuleIDs, rule.ID)
				continue out
			}
		}

		notFoundRules = append(notFoundRules, rule.ID)
	}

	if err := d.Set("rule_id", updatedRuleIDs); err != nil {
		return err
	}

	d.Set("client_id", clientID)

	if len(updatedRuleIDs) == 0 {
		log.Printf("[WARN] these rule IDs: %v have been found under the action ID: %d. But they aren't in the Terraform Plan.", notFoundRules, actionID)
		d.SetId("")
	}

	return nil
}

func resourceWallarmVpatchDelete(d *schema.ResourceData, m interface{}) error {
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

func resourceWallarmVpatchUpdate(d *schema.ResourceData, m interface{}) error {
	if err := resourceWallarmVpatchDelete(d, m); err != nil {
		return err
	}
	return resourceWallarmVpatchCreate(d, m)
}
