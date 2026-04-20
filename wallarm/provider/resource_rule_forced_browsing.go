package wallarm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
)

// nolint:dupl
func resourceWallarmForcedBrowsing() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action":    resourcerule.ScopeActionSchema(),
		"threshold": thresholdSchema,
		"reaction":  reactionSchema,
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"monitoring", "block"}, false),
			ForceNew:     true,
		},
		"advanced_conditions":  advancedConditionsSchema,
		"arbitrary_conditions": arbitraryConditionsSchema,
	}
	sh := lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields)

	return &schema.Resource{
		CreateContext: resourceWallarmForcedBrowsingCreate,
		ReadContext:   resourceWallarmForcedBrowsingRead,
		UpdateContext: resourcerule.ResourceRuleWallarmUpdate(apiClient),
		DeleteContext: resourceWallarmForcedBrowsingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.ResourceRuleWallarmImport("forced_browsing"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        sh,
	}
}

func resourceWallarmForcedBrowsingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourcerule.ResourceRuleWallarmCreate(ctx, d, apiClient(m), clientID,
		"forced_browsing", "dirbust", resourceWallarmForcedBrowsingRead, m)
}

func resourceWallarmForcedBrowsingRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.ResourceRuleWallarmRead(d, clientID, apiClient(m),
		resourcerule.ReadOptionWithMode,
		resourcerule.ReadOptionWithAction,
		resourcerule.ReadOptionWithThreshold,
		resourcerule.ReadOptionWithReaction,
		resourcerule.ReadOptionWithEnumeratedParameters,
		resourcerule.ReadOptionWithArbitraryConditions,
	))
}

func resourceWallarmForcedBrowsingDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
