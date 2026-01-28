package wallarm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmRegex() *schema.Resource {
	fields := map[string]*schema.Schema{
		"regex_id": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"attack_type": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"action": defaultResourceRuleActionSchema,

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
	}
	return &schema.Resource{
		Create: resourceWallarmRegexCreate,
		Read:   resourceWallarmRegexRead,
		Delete: resourceWallarmRegexDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmRegexImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmRegexCreate(d *schema.ResourceData, m interface{}) error {
	experimental := d.Get("experimental").(bool)
	var actionType string
	if experimental {
		actionType = experimentalRegex
	} else {
		actionType = "regex"
	}

	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	regex := d.Get("regex").(string)
	attackType := d.Get("attack_type").(string)

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

	rx := &wallarm.ActionCreate{
		Type:                actionType,
		AttackType:          attackType,
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Regex:               regex,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
		Mitigation:          fields.Mitigation,
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
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API))
}

func resourceWallarmRegexDelete(d *schema.ResourceData, m interface{}) error {
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

func resourceWallarmRegexImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "rate_limit")

		existingID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, ruleID, hintType)
		d.SetId(existingID)

		if hintType == "experimental_regex" {
			d.Set("experimental", true)
		} else {
			d.Set("experimental", false)
		}
	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}/{regex/experimental_regex}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
