package wallarm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmRateLimit() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": defaultResourceLimitActionSchema,

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
			Optional:     true,
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
	}
	return &schema.Resource{
		Create: resourceWallarmRateLimitCreate,
		Read:   resourceWallarmRateLimitRead,
		Delete: resourceWallarmRateLimitDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmRateLimitImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmRateLimitCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
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
		Type:       "rate_limit",
		Clientid:   clientID,
		Action:     &action,
		Validated:  false,
		Comment:    fields.Comment,
		Point:      point,
		Delay:      delay,
		Burst:      burst,
		Rate:       rate,
		RspStatus:  rspStatus,
		TimeUnit:   timeUnit,
		Set:        fields.Set,
		Active:     fields.Active,
		Title:      fields.Title,
		Mitigation: fields.Mitigation,
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
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API), common.ReadOptionWithAction)
}

func resourceWallarmRateLimitDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
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

func resourceWallarmRateLimitImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "rate_limit")

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{"rate_limit"},
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
		d.Set("delay", (*actionHints.Body)[0].Delay)
		d.Set("rate", (*actionHints.Body)[0].Rate)
		d.Set("burst", (*actionHints.Body)[0].Burst)
		d.Set("rsp_status", (*actionHints.Body)[0].RspStatus)
		d.Set("time_unit", (*actionHints.Body)[0].TimeUnit)
		d.Set("suffix", (*actionHints.Body)[0].Suffix)

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
