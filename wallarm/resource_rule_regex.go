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

func resourceWallarmRegex() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmRegexCreate,
		Read:   resourceWallarmRegexRead,
		Delete: resourceWallarmRegexDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmRegexImport,
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

			"regex_id": {
				Type:     schema.TypeFloat,
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

			"attack_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{"any", "sqli", "rce", "crlf", "nosqli", "ptrav",
					"xxe", "ptrav", "xss", "scanner", "redir", "ldapi", "vpatch"}, false),
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

									"query": {
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
		},
	}
}

func resourceWallarmRegexCreate(d *schema.ResourceData, m interface{}) error {
	experimental := d.Get("experimental").(bool)
	var actionType string
	if experimental {
		actionType = "experimental_regex"
	} else {
		actionType = "regex"
	}

	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)
	regex := d.Get("regex").(string)
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

	rx := &wallarm.ActionCreate{
		Type:                actionType,
		AttackType:          attackType,
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Regex:               regex,
		Validated:           false,
		Comment:             comment,
		VariativityDisabled: true,
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
	clientID := retrieveClientID(d, client)
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
			ActionID:   rule.ActionID,
			Type:       rule.Type,
			AttackType: rule.AttackType,
			RegexID:    rule.RegexID,
			Regex:      rule.Regex,
			Point:      alignedPoints,
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

func resourceWallarmRegexDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)

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

	resourceWallarmRegexRead(d, m)

	return []*schema.ResourceData{d}, nil
}
