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
		"action": defaultResourceRuleActionSchema,
	}
	return &schema.Resource{
		Create: resourceWallarmCredentialStuffingPointCreate,
		Read:   resourceWallarmCredentialStuffingPointRead,
		Update: resourceWallarmCredentialStuffingPointUpdate,
		Delete: resourceWallarmCredentialStuffingPointDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmCredentialStuffingPointImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmCredentialStuffingPointCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	credStuffType := d.Get("cred_stuff_type").(string)

	iPoint := d.Get("point").([]interface{})
	point, err := expandPointsToTwoDimensionalArray(iPoint)
	if err != nil {
		return err
	}

	iLoginPoint := d.Get("login_point").([]interface{})
	loginPoint, err := expandPointsToTwoDimensionalArray(iLoginPoint)
	if err != nil {
		return err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
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

	return resourceWallarmCredentialStuffingPointRead(d, m)
}

func resourceWallarmCredentialStuffingPointRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	ruleID := d.Get("rule_id").(int)

	rule, err := findRule(client, clientID, ruleID)
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

	d.Set("cred_stuff_type", rule.CredStuffType)
	if err := d.Set("point", wrapPointElements(rule.Point)); err != nil {
		return fmt.Errorf("error setting point: %w", err)
	}
	if err := d.Set("login_point", wrapPointElements(rule.LoginPoint)); err != nil {
		return fmt.Errorf("error setting login_point: %w", err)
	}
	d.Set("rule_type", rule.Type)
	d.Set("action_id", rule.ActionID)
	d.Set("active", rule.Active)
	d.Set("title", rule.Title)
	d.Set("mitigation", rule.Mitigation)
	d.Set("set", rule.Set)
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

func resourceWallarmCredentialStuffingPointDelete(d *schema.ResourceData, m interface{}) error {
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

func resourceWallarmCredentialStuffingPointUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return err
}

func resourceWallarmCredentialStuffingPointImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(wallarm.API)
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

	_, err = findRule(client, clientID, ruleID)
	if err != nil {
		return nil, err
	}

	ruleType := "credentials_point"

	d.Set("client_id", clientID)
	d.Set("rule_id", ruleID)
	d.Set("action_id", actionID)
	d.Set("type", ruleType)

	hint := &wallarm.HintRead{
		Limit:     1000,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ID:       []int{ruleID},
			Type:     []string{ruleType},
		},
	}
	actionHints, err := client.HintRead(hint)
	if err != nil {
		return nil, err
	}
	actionsSet := schema.Set{
		F: hashResponseActionDetails,
	}
	if len((*actionHints.Body)) != 0 && len((*actionHints.Body)[0].Action) != 0 {
		for _, a := range (*actionHints.Body)[0].Action {
			acts, err := actionDetailsToMap(a)
			if err != nil {
				return nil, err
			}
			actionsSet.Add(acts)
		}
		if err := d.Set("action", &actionsSet); err != nil {
			return nil, fmt.Errorf("error setting action: %w", err)
		}
	}

	pointInterface := (*actionHints.Body)[0].Point
	point := wrapPointElements(pointInterface)
	if err := d.Set("point", point); err != nil {
		return nil, fmt.Errorf("error setting point: %w", err)
	}
	pointInterface = (*actionHints.Body)[0].LoginPoint
	point = wrapPointElements(pointInterface)
	if err := d.Set("login_point", point); err != nil {
		return nil, fmt.Errorf("error setting login_point: %w", err)
	}

	existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
	d.SetId(existingID)

	return []*schema.ResourceData{d}, nil
}
