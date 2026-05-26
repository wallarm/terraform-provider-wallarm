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
func resourceWallarmFileUploadSizeLimit() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": resourcerule.ScopeActionSchema(),
		// Override of `defaultPointSchema` (which is Required) — the API
		// treats `point` as Optional for this rule type. When omitted, the
		// API's own default scope is applied. ForceNew preserved because
		// changing the upload point still requires recreating the rule.
		"point": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
		},
		// Schema actualised against API ground truth (probed 2026-05-01).
		// `mode` is Optional+Default("monitoring") — stable API default,
		// mutable via WithMode; symmetric remove-restores-default per
		// references/schema-decisions.md §A row 2.
		// `size_unit` stays Optional+Computed+ForceNew: API default "b" but
		// ForceNew + Default would be the import trap (anti-pattern 3).
		"mode": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "monitoring",
			ValidateFunc: validation.StringInSlice([]string{"monitoring", "block", "off", "default"}, false),
		},
		"size": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntAtLeast(1),
		},
		"size_unit": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{"b", "kb", "mb", "gb", "tb"}, false),
			ForceNew:     true,
		},
	}
	sh := lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields)

	return &schema.Resource{
		CreateContext: resourceWallarmFileUploadSizeLimitCreate,
		ReadContext:   resourceWallarmFileUploadSizeLimitRead,
		UpdateContext: resourcerule.Update(apiClient, resourcerule.WithMode, resourcerule.WithSize),
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("file_upload_size_limit"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        sh,
	}
}

func resourceWallarmFileUploadSizeLimitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourcerule.Create(ctx, d, apiClient(m), clientID,
		"file_upload_size_limit", "", resourceWallarmFileUploadSizeLimitRead, m)
}

func resourceWallarmFileUploadSizeLimitRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m)))
}
