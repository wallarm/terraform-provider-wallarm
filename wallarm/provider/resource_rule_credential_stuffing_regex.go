package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
		"action": defaultResourceRuleActionSchema,
	}
	return &schema.Resource{
		Create: resourceWallarmCredentialStuffingRegexCreate,
		Read:   resourceWallarmCredentialStuffingRegexRead,
		Update: resourceWallarmCredentialStuffingRegexUpdate,
		Delete: resourceWallarmCredentialStuffingRegexDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmCredentialStuffingRegexImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmCredentialStuffingRegexCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	regex := d.Get("regex").(string)
	credStuffType := d.Get("cred_stuff_type").(string)
	caseSensitive := d.Get("case_sensitive").(bool)
	loginRegex := d.Get("login_regex").(string)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
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
		Mitigation:          fields.Mitigation,
	})
	if err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%d/%d", resp.Body.Clientid, resp.Body.ActionID, resp.Body.ID)
	d.SetId(resID)
	d.Set("client_id", resp.Body.Clientid)
	d.Set("action_id", resp.Body.ActionID)
	d.Set("rule_id", resp.Body.ID)

	return resourceWallarmCredentialStuffingRegexRead(d, m)
}

func resourceWallarmCredentialStuffingRegexRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
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
		return err
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
	d.Set("variativity_disabled", true)
	if rule.Comment == "" {
		d.Set("comment", "Managed by Terraform")
	} else {
		d.Set("comment", rule.Comment)
	}
	actionsSet := schema.Set{F: hashResponseActionDetails}
	for _, a := range rule.Action {
		acts, err := actionDetailsToMap(a)
		if err != nil {
			return err
		}
		actionsSet.Add(acts)
	}
	if err := d.Set("action", &actionsSet); err != nil {
		return fmt.Errorf("error setting action: %w", err)
	}

	return nil
}

func resourceWallarmCredentialStuffingRegexDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	ruleID := d.Get("rule_id").(int)

	err := client.HintDelete(&wallarm.HintDelete{
		Filter: &wallarm.HintDeleteFilter{
			Clientid: []int{clientID},
			ID:       ruleID,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func resourceWallarmCredentialStuffingRegexUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return err
}

func resourceWallarmCredentialStuffingRegexImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(wallarm.API)
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

	_, err = findCredentialStuffingRule(client, clientID, ruleID)
	if err != nil {
		return nil, err
	}

	d.Set("client_id", clientID)
	d.Set("rule_id", ruleID)

	return []*schema.ResourceData{d}, nil
}
