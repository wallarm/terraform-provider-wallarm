package wallarm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmSetResponseHeader() *schema.Resource {
	fields := map[string]*schema.Schema{
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"append", "replace"}, false),
		},

		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"values": {
			Type:     schema.TypeSet,
			Required: true,
			ForceNew: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"action": resourcerule.ScopeActionSchema(),
	}
	return &schema.Resource{
		CreateContext: resourceWallarmSetResponseHeaderCreate,
		ReadContext:   resourceWallarmSetResponseHeaderRead,
		UpdateContext: resourceWallarmSetResponseHeaderUpdate,
		DeleteContext: resourceWallarmSetResponseHeaderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmSetResponseHeaderImport,
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmSetResponseHeaderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	mode := d.Get("mode").(string)
	name := d.Get("name").(string)
	valuesInterface := d.Get("values").(*schema.Set).List()

	values := make([]string, 0, len(valuesInterface))
	for _, item := range valuesInterface {
		str, _ := item.(string)
		values = append(values, str)
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	vp := &wallarm.ActionCreate{
		Type:                "set_response_header",
		Clientid:            clientID,
		Action:              &action,
		Mode:                mode,
		Name:                name,
		Values:              values,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
	}
	actionResp, err := client.HintCreate(vp)

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_id", actionResp.Body.ID)
	d.Set("client_id", clientID)
	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmSetResponseHeaderRead(ctx, d, m)
}

func resourceWallarmSetResponseHeaderRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.ResourceRuleWallarmRead(d, clientID, apiClient(m),
		common.ReadOptionWithMode, common.ReadOptionWithName, common.ReadOptionWithValues))
}

func resourceWallarmSetResponseHeaderDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceWallarmSetResponseHeaderUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return diag.FromErr(err)
}

func resourceWallarmSetResponseHeaderImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "set_response_header")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
