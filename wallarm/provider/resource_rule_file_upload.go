package wallarm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
		"action": defaultResourceRuleActionSchema,
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
	sh := lo.Assign(fields, commonResourceRuleFields)

	return &schema.Resource{
		CreateContext: resourceWallarmFileUploadSizeLimitCreate,
		ReadContext:   resourceWallarmFileUploadSizeLimitRead,
		UpdateContext: resourceWallarmFileUploadSizeLimitUpdate,
		DeleteContext: resourceWallarmFileUploadSizeLimitDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmFileUploadSizeLimitImport,
		},
		Schema: sh,
	}
}

func resourceWallarmFileUploadSizeLimitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourcerule.ResourceRuleWallarmCreate(ctx, d, apiClient(m), clientID,
		"file_upload_size_limit", "", resourceWallarmFileUploadSizeLimitRead, m)
}

func resourceWallarmFileUploadSizeLimitRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.ResourceRuleWallarmRead(d, clientID, apiClient(m)))
}

func resourceWallarmFileUploadSizeLimitDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"file_upload_size_limit"},
			Clientid: []int{clientID},
			ID:       []int{actionID},
		},
		Limit:  DefaultAPIListLimit,
		Offset: 0,
	}
	respRules, err := client.RuleRead(rule)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(respRules.Body) == 1 && respRules.Body[0].Hints == 1 && respRules.Body[0].GroupedHintsCount == 1 {
		if err = client.ActionDelete(actionID); err != nil {
			return diag.FromErr(err)
		}
	} else {
		ruleID := d.Get("rule_id").(int)
		h := &wallarm.HintDelete{
			Filter: &wallarm.HintDeleteFilter{
				Clientid: []int{clientID},
				ID:       ruleID,
			},
		}

		if err = client.HintDelete(h); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceWallarmFileUploadSizeLimitUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return diag.FromErr(err)
}

func resourceWallarmFileUploadSizeLimitImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 3)
	if len(idAttr) == 3 {
		clientID, err := strconv.Atoi(idAttr[0])
		if err != nil {
			return nil, err
		}
		actionID, err := strconv.Atoi(idAttr[1])
		if err != nil {
			return nil, err
		}
		ruleID, err := strconv.Atoi(idAttr[2])
		if err != nil {
			return nil, err
		}
		d.Set("action_id", actionID)
		d.Set("rule_id", ruleID)
		d.Set("rule_type", "file_upload_size_limit")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
