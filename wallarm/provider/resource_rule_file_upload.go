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
func resourceWallarmFileUploadSizeLimit() *schema.Resource {
	fields := map[string]*schema.Schema{
		"action": resourcerule.ScopeActionSchema(),
		"point":  defaultPointSchema,
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"monitoring", "block", "off", "default"}, false),
			ForceNew:     true,
		},
		"size": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"size_unit": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"b", "kb", "mb", "gb", "tb"}, false),
			ForceNew:     true,
		},
	}
	sh := lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields)

	return &schema.Resource{
		CreateContext: resourceWallarmFileUploadSizeLimitCreate,
		ReadContext:   resourceWallarmFileUploadSizeLimitRead,
		UpdateContext: resourcerule.Update(apiClient),
		DeleteContext: resourceWallarmFileUploadSizeLimitDelete,
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

func resourceWallarmFileUploadSizeLimitDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
