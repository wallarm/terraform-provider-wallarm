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
func resourceWallarmGraphqlDetection() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": resourcerule.ScopeActionSchema(),
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"off", "default", "monitoring", "block"}, false),
			ForceNew:     true,
		},
		"max_depth": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_value_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_doc_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_alias_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"max_doc_per_batch": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"introspection": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"debug_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
	}
	sh := lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields)

	return &schema.Resource{
		CreateContext: resourceWallarmGraphqlDetectionCreate,
		ReadContext:   resourceWallarmGraphqlDetectionRead,
		UpdateContext: resourcerule.ResourceRuleWallarmUpdate(apiClient),
		DeleteContext: resourceWallarmGraphqlDetectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.ResourceRuleWallarmImport("graphql_detection"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        sh,
	}
}

func resourceWallarmGraphqlDetectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourcerule.ResourceRuleWallarmCreate(ctx, d, apiClient(m), clientID,
		"graphql_detection", "", resourceWallarmGraphqlDetectionRead, m)
}

func resourceWallarmGraphqlDetectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.ResourceRuleWallarmRead(d, clientID, apiClient(m)))
}

func resourceWallarmGraphqlDetectionDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
