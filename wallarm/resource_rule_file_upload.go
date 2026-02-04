package wallarm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/samber/lo"
)

// nolint:dupl
func resourceWallarmFileUploadSizeLimit() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": defaultResourceRuleActionSchema,
		"point": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
		},
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"monitoring", "block", "off", "default"}, false),
			ForceNew:     true,
		},
		"size": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"size_unit": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"b", "kb", "mb", "gb", "tb"}, false),
			ForceNew:     true,
		},
	}
	sh := lo.Assign(fields, commonResourceRuleFields)

	return &schema.Resource{
		Create: resourceWallarmFileUploadSizeLimitCreate,
		Read:   resourceWallarmFileUploadSizeLimitRead,
		Update: resourceWallarmFileUploadSizeLimitUpdate,
		Delete: resourceWallarmFileUploadSizeLimitDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmFileUploadSizeLimitImport,
		},
		Schema: sh,
	}
}

func resourceWallarmFileUploadSizeLimitCreate(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmCreate(d, m.(wallarm.API), retrieveClientID(d),
		"file_upload_size_limit", "", resourceWallarmFileUploadSizeLimitRead)
}

func resourceWallarmFileUploadSizeLimitRead(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API))
}

func resourceWallarmFileUploadSizeLimitDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"file_upload_size_limit"},
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

func resourceWallarmFileUploadSizeLimitUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{VariativityDisabled: lo.ToPtr(true)})
	return err
}

func resourceWallarmFileUploadSizeLimitImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "file_upload_size_limit")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
