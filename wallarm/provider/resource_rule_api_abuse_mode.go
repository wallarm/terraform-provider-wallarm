package wallarm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"

	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

const ruleTypeAPIAbuseMode = "api_abuse_mode"

func resourceWallarmAPIAbuseMode() *schema.Resource {
	fields := map[string]*schema.Schema{
		"mode": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "enabled",
			ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			Description:  "API abuse mode. One of: enabled, disabled. Default: enabled. Changing this destroys and recreates the rule.",
		},
		"action": resourcerule.ScopeActionSchema(),
	}
	return &schema.Resource{
		CreateContext: resourceWallarmAPIAbuseModeCreate,
		ReadContext:   resourceWallarmAPIAbuseModeRead,
		UpdateContext: resourcerule.Update(apiClient, resourcerule.WithMode),
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import(ruleTypeAPIAbuseMode),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmAPIAbuseModeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if diags := guardExistingHint(d, m, ruleTypeAPIAbuseMode, "wallarm_rule_api_abuse_mode", nil); diags.HasError() {
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

	resp, err := client.HintCreate(&wallarm.ActionCreate{
		Type:                ruleTypeAPIAbuseMode,
		Clientid:            clientID,
		Action:              &action,
		Mode:                d.Get("mode").(string),
		Active:              fields.Active,
		Comment:             fields.Comment,
		Title:               fields.Title,
		Validated:           false,
		VariativityDisabled: true,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("rule_id", resp.Body.ID)
	d.Set("action_id", resp.Body.ActionID)
	d.Set("rule_type", ruleTypeAPIAbuseMode)
	d.SetId(fmt.Sprintf("%d/%d/%d", clientID, resp.Body.ActionID, resp.Body.ID))
	return resourceWallarmAPIAbuseModeRead(ctx, d, m)
}

func resourceWallarmAPIAbuseModeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m)))
}
