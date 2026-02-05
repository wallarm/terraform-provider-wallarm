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
func resourceWallarmGraphqlDetection() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": defaultResourceRuleActionSchema,
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"monitoring", "block"}, false),
			ForceNew:     true,
		},
		"max_depth": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_value_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_doc_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_alias_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_doc_per_batch": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"introspection": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"debug_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
	}
	sh := lo.Assign(fields, commonResourceRuleFields)

	return &schema.Resource{
		Create: resourceWallarmGraphqlDetectionCreate,
		Read:   resourceWallarmGraphqlDetectionRead,
		Update: resourceWallarmGraphqlDetectionUpdate,
		Delete: resourceWallarmGraphqlDetectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmGraphqlDetectionImport,
		},
		Schema: sh,
	}
}

func resourceWallarmGraphqlDetectionCreate(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmCreate(d, m.(wallarm.API), retrieveClientID(d),
		"graphql_detection", "", resourceWallarmGraphqlDetectionRead)
}

func resourceWallarmGraphqlDetectionRead(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API))
}

func resourceWallarmGraphqlDetectionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"graphql_detection"},
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

func resourceWallarmGraphqlDetectionUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return err
}

func resourceWallarmGraphqlDetectionImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "graphql_detection")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
