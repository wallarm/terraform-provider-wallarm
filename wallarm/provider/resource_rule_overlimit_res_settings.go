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
			ForceNew:     true,
			Required:     true,
			ValidateFunc: validation.IntBetween(0, 2_147_483_647),
		},

		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"off", "monitoring", "blocking"}, false),
		},
	}
	return &schema.Resource{
		CreateContext: resourceWallarmOverlimitResSettingsCreate,
		ReadContext:   resourceWallarmOverlimitResSettingsRead,
		UpdateContext: resourcerule.Update(apiClient),
		DeleteContext: resourceWallarmOverlimitResSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("overlimit_res_settings"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmOverlimitResSettingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	overlimitTime := d.Get("overlimit_time").(int)
	mode := d.Get("mode").(string)

	actionBody := &wallarm.ActionCreate{
		Type:                "overlimit_res_settings",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		VariativityDisabled: true,
		Comment:             fields.Comment,
		Mode:                mode,
		OverlimitTime:       overlimitTime,
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

func resourceWallarmOverlimitResSettingsDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
