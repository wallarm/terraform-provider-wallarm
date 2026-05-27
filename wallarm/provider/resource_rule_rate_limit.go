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

		// Schema actualised against API ground truth (probed 2026-05-01).
		// `delay`, `burst`, `time_unit` are Optional API-side; the API has
		// its own defaults (`time_unit=rps`). `rate` and `rsp_status` are
		// the only required-by-API fields. Computed lets state preserve
		// API-echoed values when the user omits the field in HCL.
		"delay": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(0, 1000),
		},

		"burst": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(0, 1000),
		},

		"rate": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(0, 1000),
		},

		"rsp_status": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(400, 599),
		},

		// Optional+Default("rps") — stable API default, mutable via WithTimeUnit;
		// removing the line plans `current → "rps"` symmetrically (per
		// references/schema-decisions.md §A row 2).
		"time_unit": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "rps",
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
	// Required ints (rate, rsp_status): always send. Optional ints (delay,
	// burst): use the GetRawConfig-aware helper so a literal 0 reaches the
	// API but an omitted field doesn't override the API default — relies on
	// wallarm-go v0.12.1's `*int+omitempty` for these fields. `time_unit` is
	// `Optional+Default("rps")`, so `d.Get` always returns a non-empty value
	// (user's, or the schema default) — a plain `d.Get(...).(string)` here.
	actionBody := &wallarm.ActionCreate{
		Type:      "rate_limit",
		Clientid:  clientID,
		Action:    &action,
		Validated: false,
		Comment:   fields.Comment,
		Point:     point,
		Delay:     resourcerule.GetPointerIfConfigured[int](d, "delay"),
		Burst:     resourcerule.GetPointerIfConfigured[int](d, "burst"),
		Rate:      lo.ToPtr(d.Get("rate").(int)),
		RspStatus: d.Get("rsp_status").(int),
		TimeUnit:  d.Get("time_unit").(string),
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
