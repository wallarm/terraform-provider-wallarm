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

func resourceWallarmIgnoreRegex() *schema.Resource {
	fields := map[string]*schema.Schema{
		"regex_id": {
			Type:     schema.TypeInt,
			Required: true,
			ForceNew: true,
		},

		"action": resourcerule.ScopeActionSchema(),

		"point": defaultPointSchema,
	}
	return &schema.Resource{
		CreateContext: resourceWallarmIgnoreRegexCreate,
		ReadContext:   resourceWallarmIgnoreRegexRead,
		UpdateContext: resourceWallarmIgnoreRegexUpdate,
		DeleteContext: resourceWallarmIgnoreRegexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmIgnoreRegexImport,
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmIgnoreRegexCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	regexID := d.Get("regex_id").(int)

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

	vp := &wallarm.ActionCreate{
		Type:                "disable_regex",
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		RegexID:             regexID,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
	}

	actionResp, err := client.HintCreate(vp)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)
	d.Set("regex_id", actionResp.Body.RegexID)

	resID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionResp.Body.ActionID, actionResp.Body.ID, actionResp.Body.Type)
	d.SetId(resID)

	return resourceWallarmIgnoreRegexRead(ctx, d, m)
}

func resourceWallarmIgnoreRegexRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.ResourceRuleWallarmRead(d, clientID, apiClient(m),
		resourcerule.ReadOptionWithPoint, resourcerule.ReadOptionWithRegexID))
}

func resourceWallarmIgnoreRegexDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceWallarmIgnoreRegexUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return diag.FromErr(err)
}

func resourceWallarmIgnoreRegexImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 4)
	if len(idAttr) < 3 {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}[/{type}]\"", d.Id())
	}

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

	d.Set("client_id", clientID)
	d.Set("action_id", actionID)
	d.Set("rule_id", ruleID)
	d.Set("rule_type", "disable_regex")

	ruleType := "disable_regex"
	if len(idAttr) == 4 {
		ruleType = idAttr[3]
	}
	d.SetId(fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, ruleID, ruleType))

	return []*schema.ResourceData{d}, nil
}
