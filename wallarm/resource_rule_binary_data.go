package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/samber/lo"
)

func resourceWallarmBinaryData() *schema.Resource {
	fields := map[string]*schema.Schema{
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
	}

	return &schema.Resource{
		Create: resourceWallarmBinaryDataCreate,
		Read:   resourceWallarmBinaryDataRead,
		Delete: resourceWallarmBinaryDataDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmBinaryDataImport,
		},
		Update: resourceWallarmBinaryDataUpdate,
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmBinaryDataCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)

	ps := d.Get("point").([]interface{})
	d.Set("point", ps)

	points, err := expandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	wm := &wallarm.ActionCreate{
		Type:                "binary_data",
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
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
		d.SetId("")
		return err
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmBinaryDataRead(d, m)
}

func resourceWallarmBinaryDataRead(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API), common.ReadOptionWithPoint)
}

func resourceWallarmBinaryDataDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"binary_data"},
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

		if err = client.HintDelete(h); err != nil {
			return err
		}
	}
	return nil
}

func resourceWallarmBinaryDataImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "binary_data")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)
	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}

func resourceWallarmBinaryDataUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	log.Printf("DEBUGGG before update resourceWallarmBinaryDataUpdate, action_id: %v\n", d.Get("rule_id"))
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{VariativityDisabled: lo.ToPtr(true)})
	log.Printf("DEBUGGG after update resourceWallarmBinaryDataUpdate, action_id: %v, sucess=%t\n", d.Get("rule_id"), err == nil)

	return err
}
