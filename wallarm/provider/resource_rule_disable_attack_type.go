package wallarm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmDisableAttackType() *schema.Resource {
	fields := map[string]*schema.Schema{
		"attack_type": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			Description: `Possible values: "any", "sqli", "rce", "crlf", "nosqli", "ptrav",
				"xxe", "xss", "scanner", "redir", "ldapi", "mass_assignment", "ssrf"`,
		},

		"action": resourcerule.ScopeActionSchema(),

		"point": defaultPointSchema,
	}
	return &schema.Resource{
		CreateContext: resourceWallarmDisableAttackTypeCreate,
		ReadContext:   resourceWallarmDisableAttackTypeRead,
		UpdateContext: resourcerule.ResourceRuleWallarmUpdate(apiClient),
		DeleteContext: resourceWallarmDisableAttackTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.ResourceRuleWallarmImport("disable_attack_type"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmDisableAttackTypeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
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

	wm := &wallarm.ActionCreate{
		Type:                "disable_attack_type",
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Validated:           false,
		Comment:             fields.Comment,
		AttackType:          attackType,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
	}

	actionResp, err := client.HintCreate(wm)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmDisableAttackTypeRead(ctx, d, m)
}

func resourceWallarmDisableAttackTypeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.ResourceRuleWallarmRead(d, clientID, apiClient(m), resourcerule.ReadOptionWithPoint))
}

func resourceWallarmDisableAttackTypeDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

// nolint:dupl
