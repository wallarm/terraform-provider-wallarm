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
		// All API-defaulted int fields below use Optional+Default(<API default>).
		// Schema default matches the API default so re-plan is clean, AND
		// removing the line from HCL plans `current → default` symmetrically
		// (see references/schema-decisions.md §A row 2).
		// All four are mutable in-place via Update — confirmed by per-field
		// PUT mutability probe. max_value_size_kb has documented range 1..100;
		// other fields API-enforced.
		"max_depth": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  10,
		},
		"max_value_size_kb": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      10,
			ValidateFunc: validation.IntBetween(1, 100),
		},
		"max_doc_size_kb": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  100,
		},
		"max_aliases": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  5,
		},
		"max_doc_per_batch": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  10,
		},
		// API default is true. Optional+Default (NOT Computed) so removing the
		// field from HCL plans as "back to default" — symmetric with adding
		// `= false` planning as "true → false". Optional+Computed would leave
		// state stuck at the last user value when the HCL line is removed.
		"introspection": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"debug_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
	}
	sh := lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields)

	return &schema.Resource{
		CreateContext: resourceWallarmGraphqlDetectionCreate,
		ReadContext:   resourceWallarmGraphqlDetectionRead,
		UpdateContext: resourcerule.Update(apiClient, resourcerule.WithMode, resourcerule.WithMaxDepth, resourcerule.WithMaxValueSizeKb, resourcerule.WithMaxDocSizeKb, resourcerule.WithMaxAliases, resourcerule.WithMaxDocPerBatch, resourcerule.WithIntrospection, resourcerule.WithDebugEnabled),
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
