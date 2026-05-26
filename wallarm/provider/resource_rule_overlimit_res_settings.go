package wallarm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmOverlimitResSettings() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": resourcerule.ScopeActionSchema(),

		"overlimit_time": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(0, 2_147_483_647),
		},

		// `mode` actualised against API ground truth (probed 2026-05-01) —
		// API default `monitoring`. Optional+Default("monitoring") for
		// symmetric remove-restores-default semantics; mutable via WithMode
		// (per references/schema-decisions.md §A row 2).
		"mode": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "monitoring",
			ValidateFunc: validation.StringInSlice([]string{"off", "monitoring", "blocking"}, false),
		},
	}
	return &schema.Resource{
		CreateContext: resourceWallarmOverlimitResSettingsCreate,
		ReadContext:   resourceWallarmOverlimitResSettingsRead,
		UpdateContext: resourcerule.Update(apiClient, resourcerule.WithMode, resourcerule.WithOverlimitTime),
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("overlimit_res_settings"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmOverlimitResSettingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if diags := guardExistingHint(d, m, "overlimit_res_settings", "wallarm_rule_overlimit_res_settings", nil); diags.HasError() {
		return diags
	}

	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	mode := d.Get("mode").(string)

	// `overlimit_time` is Required (range 0..MaxInt). wallarm-go v0.12.1
	// changed OverlimitTime to *int so callers can transmit a literal 0 that
	// the API range allows; non-pointer int+omitempty silently dropped 0.
	actionBody := &wallarm.ActionCreate{
		Type:                "overlimit_res_settings",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		VariativityDisabled: true,
		Comment:             fields.Comment,
		Mode:                mode,
		OverlimitTime:       lo.ToPtr(d.Get("overlimit_time").(int)),
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
	}

	actionResp, err := client.HintCreate(actionBody)
	if err != nil {
		return diag.FromErr(err)
	}
	actionID := actionResp.Body.ActionID

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionID)
	d.Set("rule_type", actionResp.Body.Type)
	d.Set("client_id", clientID)
	resID := fmt.Sprintf("%d/%d/%d", clientID, actionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmOverlimitResSettingsRead(ctx, d, m)
}

func resourceWallarmOverlimitResSettingsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m), resourcerule.ReadOptionWithAction))
}
