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

func resourceWallarmCredentialStuffingRegex() *schema.Resource {
	fields := map[string]*schema.Schema{
		"regex": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"cred_stuff_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "default",
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"custom", "default"}, false),
		},
		"case_sensitive": {
			Type:     schema.TypeBool,
			Required: true,
			ForceNew: true,
		},
		"login_regex": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(1, 4096),
		},
		"action": resourcerule.ScopeActionSchema(),
	}
	return &schema.Resource{
		CreateContext: resourceWallarmCredentialStuffingRegexCreate,
		ReadContext:   resourceWallarmCredentialStuffingRegexRead,
		UpdateContext: resourcerule.Update(apiClient),
		DeleteContext: resourceWallarmCredentialStuffingRegexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmCredentialStuffingRegexImport,
		},
		CustomizeDiff: resourcerule.ActionScopeCustomizeDiff,
		Schema:        lo.Assign(fields, commonResourceRuleFields, resourcerule.ActionScopeFields),
	}
}

func resourceWallarmCredentialStuffingRegexCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	regex := d.Get("regex").(string)
	credStuffType := d.Get("cred_stuff_type").(string)
	caseSensitive := d.Get("case_sensitive").(bool)
	loginRegex := d.Get("login_regex").(string)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.HintCreate(&wallarm.ActionCreate{
		Type:                "credentials_regex",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Regex:               regex,
		LoginRegex:          loginRegex,
		CredStuffType:       credStuffType,
		CaseSensitive:       &caseSensitive,
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

	// Invalidate so the following Read re-fetches from the v4 API and picks up the new rule.
	m.(*ProviderMeta).CredentialStuffingCache.Invalidate()
	return resourceWallarmCredentialStuffingRegexRead(ctx, d, m)
}

func resourceWallarmCredentialStuffingRegexRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	ruleID := d.Get("rule_id").(int)

	rule, err := m.(*ProviderMeta).CredentialStuffingCache.GetOrFetch(client, clientID, ruleID)
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

	d.Set("regex", rule.Regex)
	d.Set("cred_stuff_type", rule.CredStuffType)
	d.Set("case_sensitive", rule.CaseSensitive)
	d.Set("login_regex", rule.LoginRegex)
	d.Set("rule_type", rule.Type)
	d.Set("action_id", rule.ActionID)
	d.Set("active", rule.Active)
	d.Set("title", rule.Title)
	d.Set("mitigation", rule.Mitigation)
	d.Set("set", rule.Set)
	d.Set("variativity_disabled", rule.VariativityDisabled)
	d.Set("comment", rule.Comment)
	actionsSet := schema.Set{F: resourcerule.HashActionDetails}
	for _, a := range rule.Action {
		acts, err := resourcerule.ActionDetailsToMap(a)
		if err != nil {
			return diag.FromErr(err)
		}
		resourcerule.TransformAPIActionToSchema(acts)
		actionsSet.Add(acts)
	}
	if err := d.Set("action", &actionsSet); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action: %w", err))
	}

	return nil
}

func resourceWallarmCredentialStuffingRegexDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	ruleID := d.Get("rule_id").(int)

	resp, err := client.HintDelete(&wallarm.HintDelete{
		Filter: &wallarm.HintDeleteFilter{
			Clientid: []int{clientID},
			ID:       []int{ruleID},
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	resourcerule.LogIfHintDeleteNoOp(resp, ruleID)

	m.(*ProviderMeta).CredentialStuffingCache.Invalidate()
	return nil
}

func resourceWallarmCredentialStuffingRegexImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := apiClient(m)
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	clientID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return nil, err
	}
	ruleID, err := strconv.Atoi(idParts[2])
	if err != nil {
		return nil, err
	}

	_, err = m.(*ProviderMeta).CredentialStuffingCache.GetOrFetch(client, clientID, ruleID)
	if err != nil {
		return nil, err
	}

	d.Set("client_id", clientID)
	d.Set("rule_id", ruleID)

	return []*schema.ResourceData{d}, nil
}
