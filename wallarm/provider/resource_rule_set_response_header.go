package wallarm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/samber/lo"
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
		UpdateContext: resourcerule.Update(apiClient),
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("set_response_header"),
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
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m),
		resourcerule.ReadOptionWithMode, resourcerule.ReadOptionWithName, resourcerule.ReadOptionWithValues))
}
