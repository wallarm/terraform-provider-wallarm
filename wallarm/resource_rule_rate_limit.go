package wallarm

import (
	"fmt"
	"log"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmRateLimit() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmRateLimitCreate,
		Read:   resourceWallarmRateLimitRead,
		Delete: resourceWallarmRateLimitDelete,

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
										Computed: true,
									},

									"action_ext": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Computed: true,
									},

									"query": {
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
												errs = append(errs, fmt.Errorf("%q must be greater than -1 inclusive, got: %d", key, v))
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
					Elem: &schema.Schema{Type: schema.TypeString},
				},
			},

			"delay": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 1000),
			},

			"burst": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 1000),
			},

			"rate": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 1000),
			},

			"rsp_status": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(400, 599),
			},

			"time_unit": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"rps", "rpm"}, false),
			},
		},
	}
}

func resourceWallarmRateLimitCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	iPoint := d.Get("point").([]interface{})
	point, err := expandPointsToTwoDimensionalArray(iPoint)
	if err != nil {
		return err
	}
	delay := d.Get("delay").(int)
	burst := d.Get("burst").(int)
	rate := d.Get("rate").(int)
	rspStatus := d.Get("rsp_status").(int)
	timeUnit := d.Get("time_unit").(string)

	actionBody := &wallarm.ActionCreate{
		Type:      "rate_limit",
		Clientid:  clientID,
		Action:    &action,
		Validated: false,
		Comment:   comment,
		Point:     point,
		Delay:     delay,
		Burst:     burst,
		Rate:      rate,
		RspStatus: rspStatus,
		TimeUnit:  timeUnit,
	}

	actionResp, err := client.HintCreate(actionBody)
	if err != nil {
		return err
	}
	actionID := actionResp.Body.ActionID

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionID)
	d.Set("rule_type", actionResp.Body.Type)
	d.Set("client_id", clientID)
	d.Set("point", actionResp.Body.Point)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionID, actionResp.Body.ID)
	d.SetId(resID)

	return nil
}

func resourceWallarmRateLimitRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	actionID := d.Get("action_id").(int)
	ruleID := d.Get("rule_id").(int)

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

	// This is mandatory to fill in the default values in order to compare them deeply.
	// Assign new values to the old struct slice.
	fillInDefaultValues(&action)

	expectedRule := wallarm.ActionBody{
		ActionID: actionID,
		Type:     "rate_limit",
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

	if updatedRuleID == 0 {
		log.Printf("[WARN] these rule IDs: %v have been found under the action ID: %d. But it isn't in the Terraform Plan.", notFoundRules, actionID)
		d.SetId("")
	}

	return nil
}

func resourceWallarmRateLimitDelete(d *schema.ResourceData, m interface{}) error {
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
