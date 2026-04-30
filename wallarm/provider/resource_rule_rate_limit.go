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

func resourceWallarmRateLimit() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": resourcerule.ScopeActionSchema(),

		"point": defaultPointSchema,

		"delay": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 1000),
		},

		"burst": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(0, 1000),
		},

		"rate": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(0, 1000),
		},

		"rsp_status": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(400, 599),
		},

		"time_unit": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"rps", "rpm"}, false),
		},
	}
	return &schema.Resource{
		CreateContext: resourceWallarmRateLimitCreate,
		ReadContext:   resourceWallarmRateLimitRead,
		UpdateContext: resourcerule.Update(apiClient, resourcerule.WithDelay, resourcerule.WithBurst, resourcerule.WithRate, resourcerule.WithRspStatus, resourcerule.WithTimeUnit),
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("rate_limit"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmRateLimitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	iPoint := d.Get("point").([]interface{})
	point, err := resourcerule.ExpandPointsToTwoDimensionalArray(iPoint)
	if err != nil {
		return diag.FromErr(err)
	}
	rspStatus := d.Get("rsp_status").(int)
	timeUnit := d.Get("time_unit").(string)

	// Required ints (rate, burst): always send a pointer. d.Get returns the
	// user-supplied value (the schema's Required gate prevents omission).
	// Optional ints (delay): use the GetRawConfig-aware helper so a literal
	// 0 reaches the API but an omitted field doesn't override the API
	// default. wallarm-go v0.12.1 changed these fields to *int specifically
	// so callers can transmit 0 — non-pointer int+omitempty silently dropped
	// it, and the API rejected with "can't be blank".
	actionBody := &wallarm.ActionCreate{
		Type:      "rate_limit",
		Clientid:  clientID,
		Action:    &action,
		Validated: false,
		Comment:   fields.Comment,
		Point:     point,
		Delay:     resourcerule.GetIntPointerIfConfigured(d, "delay"),
		Burst:     lo.ToPtr(d.Get("burst").(int)),
		Rate:      lo.ToPtr(d.Get("rate").(int)),
		RspStatus: rspStatus,
		TimeUnit:  timeUnit,
		Set:       fields.Set,
		Active:    fields.Active,
		Title:     fields.Title,
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
	if err := d.Set("point", resourcerule.WrapPointElements(actionResp.Body.Point)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting point: %w", err))
	}

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmRateLimitRead(ctx, d, m)
}

func resourceWallarmRateLimitRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m), resourcerule.ReadOptionWithAction))
}
