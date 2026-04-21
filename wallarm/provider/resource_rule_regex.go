package wallarm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

		"action": resourcerule.ScopeActionSchema(),

		"regex": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"point": defaultPointSchema,

		"experimental": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
			ForceNew: true,
		},
	}
	return &schema.Resource{
		CreateContext: resourceWallarmRegexCreate,
		ReadContext:   resourceWallarmRegexRead,
		UpdateContext: resourcerule.Update(apiClient),
		DeleteContext: resourceWallarmRegexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmRegexImport,
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmRegexCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	experimental := d.Get("experimental").(bool)
	var actionType string
	if experimental {
		actionType = experimentalRegex
	} else {
		actionType = "regex"
	}

	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	regex := d.Get("regex").(string)
	attackType := d.Get("attack_type").(string)

	ps := d.Get("point").([]interface{})
	if err := d.Set("point", ps); err != nil {
		return diag.FromErr(fmt.Errorf("error setting point: %w", err))
	}

	points, err := resourcerule.ExpandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return diag.FromErr(err)
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(err)
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
	}
	regexResp, err := client.HintCreate(rx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("regex_id", regexResp.Body.RegexID.(float64))
	d.Set("rule_id", regexResp.Body.ID)
	d.Set("action_id", regexResp.Body.ActionID)
	d.Set("rule_type", regexResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d/%s", clientID, regexResp.Body.ActionID, regexResp.Body.ID, regexResp.Body.Type)
	d.SetId(resID)

	return resourceWallarmRegexRead(ctx, d, m)
}

func resourceWallarmRegexRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m)))
}

func resourceWallarmRegexDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ruleID := d.Get("rule_id").(int)
	h := &wallarm.HintDelete{
		Filter: &wallarm.HintDeleteFilter{
			Clientid: []int{clientID},
			ID:       []int{ruleID},
		},
	}

	if err := client.HintDelete(h); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWallarmRegexImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
