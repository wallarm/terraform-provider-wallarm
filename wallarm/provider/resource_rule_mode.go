package wallarm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmMode() *schema.Resource {
	fields := map[string]*schema.Schema{
		"mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"default", "off", "monitoring", "block", "safe_blocking"}, false),
			ForceNew:     true,
		},

		"action": resourcerule.ScopeActionSchema(),
	}
	return &schema.Resource{
		CreateContext: resourceWallarmModeCreate,
		ReadContext:   resourceWallarmModeRead,
		UpdateContext: resourcerule.Update(apiClient),
		DeleteContext: resourceWallarmModeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmModeImport,
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmModeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.IsNewResource() {
		actionID, rule, exists, err := existingHintForAction(d, m, "wallarm_mode")
		if err != nil {
			return diag.FromErr(err)
		}
		if exists {
			clientID, err := retrieveClientID(d, m)
			if err != nil {
				return diag.FromErr(err)
			}
			existingID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, rule.ID, rule.Mode)
			return diag.FromErr(ImportAsExistsError("wallarm_rule_mode", existingID))
		}
	}
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	actionsFromState := d.Get("action").(*schema.Set)
	mode := d.Get("mode").(string)

	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}
	wm := &wallarm.ActionCreate{
		Type:                "wallarm_mode",
		Clientid:            clientID,
		Action:              &action,
		Mode:                mode,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
	}

	actionResp, err := client.HintCreate(wm)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionResp.Body.ActionID, actionResp.Body.ID, actionResp.Body.Mode)
	d.SetId(resID)

	return resourceWallarmModeRead(ctx, d, m)
}

func resourceWallarmModeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m), resourcerule.ReadOptionWithAction))
}

func resourceWallarmModeDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceWallarmModeImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 4)
	if len(idAttr) == 4 {
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
		mode := idAttr[3]
		d.Set("action_id", actionID)
		d.Set("rule_id", ruleID)
		d.Set("rule_type", "wallarm_mode")

		existingID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, ruleID, mode)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}/{wallarm-mode}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
