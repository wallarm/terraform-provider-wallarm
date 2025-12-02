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

func resourceWallarmSetResponseHeader() *schema.Resource {
	fields := map[string]*schema.Schema{
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"append", "replace"}, false),
		},

		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"values": {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"action": defaultResourceRuleActionSchema,
	}
	return &schema.Resource{
		Create: resourceWallarmSetResponseHeaderCreate,
		Read:   resourceWallarmSetResponseHeaderRead,
		Update: resourceWallarmSetResponseHeaderUpdate,
		Delete: resourceWallarmSetResponseHeaderDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmSetResponseHeaderImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmSetResponseHeaderCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	mode := d.Get("mode").(string)
	name := d.Get("name").(string)
	valuesInterface := d.Get("values").([]interface{})

	values := make([]string, 0, len(valuesInterface))
	for _, item := range valuesInterface {
		str, _ := item.(string)
		values = append(values, str)
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	vp := &wallarm.ActionCreate{
		Type:                "set_response_header",
		Clientid:            clientID,
		Action:              &action,
		Mode:                mode,
		Name:                name,
		Values:              values,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
		Mitigation:          fields.Mitigation,
	}
	actionResp, err := client.HintCreate(vp)

	if err != nil {
		return err
	}

	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_id", actionResp.Body.ID)
	d.Set("client_id", clientID)
	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmSetResponseHeaderRead(d, m)
}

func resourceWallarmSetResponseHeaderRead(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API),
		common.ReadOptionWithMode, common.ReadOptionWithName, common.ReadOptionWithValues)
}

func resourceWallarmSetResponseHeaderDelete(d *schema.ResourceData, m interface{}) error {
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

func resourceWallarmSetResponseHeaderUpdate(d *schema.ResourceData, m interface{}) error {
	if err := resourceWallarmSetResponseHeaderDelete(d, m); err != nil {
		return err
	}
	return resourceWallarmSetResponseHeaderCreate(d, m)
}

func resourceWallarmSetResponseHeaderImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "set_response_header")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
