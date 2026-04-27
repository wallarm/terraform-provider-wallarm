package wallarm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
)

func resourceWallarmBinaryData() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": resourcerule.ScopeActionSchema(),
		"point":  defaultPointSchema,
	}

	return &schema.Resource{
		CreateContext: resourceWallarmBinaryDataCreate,
		ReadContext:   resourceWallarmBinaryDataRead,
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("binary_data"),
		},
		UpdateContext: resourcerule.Update(apiClient),
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmBinaryDataCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)

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
	}

	actionResp, err := client.HintCreate(wm)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmBinaryDataRead(ctx, d, m)
}

func resourceWallarmBinaryDataRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m), resourcerule.ReadOptionWithPoint))
}
