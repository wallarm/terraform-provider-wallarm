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

func resourceWallarmDisableAttackType() *schema.Resource {
	fields := map[string]*schema.Schema{
		// Allowlist sourced from docs/resources/rule_disable_attack_type.md,
		// the schema's pre-v2.3.9 Description (covers mass_assignment / ssrf /
		// redir / any), and CLAUDE.md hits-attack-types (covers invalid_xml).
		// Union ensures any user value the API accepts also passes plan.
		"attack_type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				"any", "sqli", "xss", "rce", "ptrav", "crlf", "nosqli", "xxe",
				"ldapi", "scanner", "redir", "ssti", "ssi", "mail_injection",
				"mass_assignment", "ssrf", "invalid_xml", "vpatch",
			}, false),
			Description: `Possible values: "any", "sqli", "xss", "rce", "ptrav", "crlf",
				"nosqli", "xxe", "ldapi", "scanner", "redir", "ssti", "ssi",
				"mail_injection", "mass_assignment", "ssrf", "invalid_xml", "vpatch"`,
		},

		"action": resourcerule.ScopeActionSchema(),

		"point": defaultPointSchema,
	}
	return &schema.Resource{
		CreateContext: resourceWallarmDisableAttackTypeCreate,
		ReadContext:   resourceWallarmDisableAttackTypeRead,
		UpdateContext: resourcerule.Update(apiClient, resourcerule.WithAttackType),
		DeleteContext: resourcerule.Delete(apiClient),
		Importer: &schema.ResourceImporter{
			StateContext: resourcerule.Import("disable_attack_type"),
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmDisableAttackTypeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	attackType := d.Get("attack_type").(string)

	ps := d.Get("point").([]interface{})
	if err := d.Set("point", ps); err != nil {
		return diag.FromErr(fmt.Errorf("error setting point: %w", err))
	}

	points, err := resourcerule.ExpandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return diag.FromErr(err)
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	wm := &wallarm.ActionCreate{
		Type:                "disable_attack_type",
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Validated:           false,
		Comment:             fields.Comment,
		AttackType:          attackType,
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

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmDisableAttackTypeRead(ctx, d, m)
}

func resourceWallarmDisableAttackTypeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcerule.Read(d, clientID, apiClient(m), resourcerule.ReadOptionWithPoint))
}

// nolint:dupl
