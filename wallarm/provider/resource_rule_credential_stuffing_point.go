package wallarm

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmCredentialStuffingPoint() *schema.Resource {
	fields := map[string]*schema.Schema{
		"point":       defaultPointSchema,
		"login_point": defaultPointSchema,
		"cred_stuff_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "default",
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"custom", "default"}, false),
		},
		"action": resourcerule.ScopeActionSchema(),
	}
	return &schema.Resource{
		CreateContext: resourceWallarmCredentialStuffingPointCreate,
		ReadContext:   resourceWallarmCredentialStuffingPointRead,
		UpdateContext: resourceWallarmCredentialStuffingPointUpdate,
		DeleteContext: resourceWallarmCredentialStuffingPointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmCredentialStuffingPointImport,
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmCredentialStuffingPointCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	credStuffType := d.Get("cred_stuff_type").(string)

	iPoint := d.Get("point").([]interface{})
	point, err := resourcerule.ExpandPointsToTwoDimensionalArray(iPoint)
	if err != nil {
		return diag.FromErr(err)
	}

	iLoginPoint := d.Get("login_point").([]interface{})
	loginPoint, err := resourcerule.ExpandPointsToTwoDimensionalArray(iLoginPoint)
	if err != nil {
		return diag.FromErr(err)
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.HintCreate(&wallarm.ActionCreate{
		Type:                "credentials_point",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Point:               point,
		LoginPoint:          loginPoint,
		CredStuffType:       credStuffType,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	resID := fmt.Sprintf("%d/%d/%d", resp.Body.Clientid, resp.Body.ActionID, resp.Body.ID)
	d.SetId(resID)
	d.Set("client_id", resp.Body.Clientid)
	d.Set("action_id", resp.Body.ActionID)
	d.Set("rule_id", resp.Body.ID)

	return resourceWallarmCredentialStuffingPointRead(context.TODO(), d, m)
}

func resourceWallarmCredentialStuffingPointRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	ruleID := d.Get("rule_id").(int)

	rule, err := findCredentialStuffingRule(client, clientID, ruleID)
	if !d.IsNewResource() {
		if _, ok := err.(*ruleNotFoundError); ok {
			log.Printf("[WARN] Rule %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
	}
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("cred_stuff_type", rule.CredStuffType)
	if err := d.Set("point", resourcerule.WrapPointElements(rule.Point)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting point: %w", err))
	}
	if err := d.Set("login_point", resourcerule.WrapPointElements(rule.LoginPoint)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting login_point: %w", err))
	}
	d.Set("rule_type", rule.Type)
	d.Set("action_id", rule.ActionID)
	d.Set("active", rule.Active)
	d.Set("title", rule.Title)
	d.Set("mitigation", rule.Mitigation)
	d.Set("set", rule.Set)
	d.Set("variativity_disabled", rule.VariativityDisabled)
	d.Set("comment", rule.Comment)
	actionsSet := schema.Set{F: resourcerule.HashResponseActionDetails}
	for _, a := range rule.Action {
		acts, err := resourcerule.ActionDetailsToMap(a)
		if err != nil {
			return diag.FromErr(err)
		}
		actionsSet.Add(acts)
	}
	if err := d.Set("action", &actionsSet); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action: %w", err))
	}

	return nil
}

func resourceWallarmCredentialStuffingPointDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	ruleID := d.Get("rule_id").(int)

	err = client.HintDelete(&wallarm.HintDelete{
		Filter: &wallarm.HintDeleteFilter{
			Clientid: []int{clientID},
			ID:       ruleID,
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWallarmCredentialStuffingPointUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return diag.FromErr(err)
}

func resourceWallarmCredentialStuffingPointImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := apiClient(m)
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	clientID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return nil, err
	}
	actionID, err := strconv.Atoi(idParts[1])
	if err != nil {
		return nil, err
	}
	ruleID, err := strconv.Atoi(idParts[2])
	if err != nil {
		return nil, err
	}

	rule, err := findCredentialStuffingRule(client, clientID, ruleID)
	if err != nil {
		return nil, err
	}

	d.Set("client_id", clientID)
	d.Set("rule_id", ruleID)
	d.Set("action_id", actionID)
	d.Set("cred_stuff_type", rule.Type)

	actionsSet := schema.Set{
		F: resourcerule.HashResponseActionDetails,
	}
	for _, a := range rule.Action {
		acts, err := resourcerule.ActionDetailsToMap(a)
		if err != nil {
			return nil, err
		}
		actionsSet.Add(acts)
	}
	if err := d.Set("action", &actionsSet); err != nil {
		return nil, fmt.Errorf("error setting action: %w", err)
	}

	if err := d.Set("point", resourcerule.WrapPointElements(rule.Point)); err != nil {
		return nil, fmt.Errorf("error setting point: %w", err)
	}
	if err := d.Set("login_point", resourcerule.WrapPointElements(rule.LoginPoint)); err != nil {
		return nil, fmt.Errorf("error setting login_point: %w", err)
	}

	existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
	d.SetId(existingID)

	return []*schema.ResourceData{d}, nil
}
