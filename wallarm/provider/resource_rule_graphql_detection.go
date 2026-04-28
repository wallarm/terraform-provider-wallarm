package wallarm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"

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
		},
		"max_depth": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"max_value_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"max_doc_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"max_alias_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"max_doc_per_batch": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"introspection": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"debug_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
	sh := lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields)

	return &schema.Resource{
		CreateContext: resourceWallarmGraphqlDetectionCreate,
		ReadContext:   resourceWallarmGraphqlDetectionRead,
		UpdateContext: resourcerule.Update(apiClient, resourcerule.WithMode, resourcerule.WithMaxDepth, resourcerule.WithMaxValueSizeKb, resourcerule.WithMaxDocSizeKb, resourcerule.WithMaxDocPerBatch, resourcerule.WithIntrospection, resourcerule.WithDebugEnabled),
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("graphql_detection"),
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
	return resourcerule.Create(ctx, d, apiClient(m), clientID,
		"graphql_detection", "", resourceWallarmGraphqlDetectionRead, m)
}

func resourceWallarmGraphqlDetectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m)))
}
