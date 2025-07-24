package wallarm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmParserState() *schema.Resource {
	fields := map[string]*schema.Schema{
		"parser": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"base64", "cookie", "form_urlencoded", "gzip", "grpc", "json_doc", "multipart", "percent", "protobuf", "htmljs", "viewstate", "xml", "jwt", "gql"}, false),
			ForceNew:     true,
		},

		"state": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			ForceNew:     true,
		},

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
		Create: resourceWallarmParserStateCreate,
		Read:   resourceWallarmParserStateRead,
		Delete: resourceWallarmParserStateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmParserStateImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmParserStateCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	parser := d.Get("parser").(string)
	state := d.Get("state").(string)

	ps := d.Get("point").([]interface{})
	d.Set("point", ps)

	points, err := expandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := common.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	wm := &wallarm.ActionCreate{
		Type:                "parser_state",
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Validated:           false,
		Comment:             fields.Comment,
		Parser:              parser,
		State:               state,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
		Mitigation:          fields.Mitigation,
	}

	actionResp, err := client.HintCreate(wm)
	if err != nil {
		return err
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmParserStateRead(d, m)
}

func resourceWallarmParserStateRead(d *schema.ResourceData, m interface{}) error {
	return common.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API), common.ReadOptionWithPoint)
}

func resourceWallarmParserStateDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"parser_state"},
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
		if err := client.RuleDelete(actionID); err != nil {
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

		if err := client.HintDelete(h); err != nil {
			return err
		}
	}
	return nil
}

func resourceWallarmParserStateImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "parser_state")

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
				Type:     []string{"parser_state"},
			},
		}
		actionHints, err := client.HintRead(hint)
		if err != nil {
			return nil, err
		}
		actionsSet := schema.Set{
			F: hashResponseActionDetails,
		}
		if len((*actionHints.Body)) != 0 && len((*actionHints.Body)[0].Action) != 0 {
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
		d.Set("parser", (*actionHints.Body)[0].Parser)
		d.Set("state", (*actionHints.Body)[0].State)

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	if err := resourceWallarmParserStateRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
