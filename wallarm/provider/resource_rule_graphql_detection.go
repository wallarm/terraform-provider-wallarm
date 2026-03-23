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
		UpdateContext: resourceWallarmGraphqlDetectionUpdate,
		DeleteContext: resourceWallarmGraphqlDetectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmGraphqlDetectionImport,
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
			ID:       ruleID,
		},
	}
	if err := client.HintDelete(h); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceWallarmGraphqlDetectionUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return diag.FromErr(err)
}

func resourceWallarmGraphqlDetectionImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "graphql_detection")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
