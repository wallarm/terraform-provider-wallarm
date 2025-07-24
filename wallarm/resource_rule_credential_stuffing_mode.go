package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmCredentialStuffingMode() *schema.Resource {
	fields := map[string]*schema.Schema{
		"mode": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "default",
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"custom", "default", "disabled"}, false),
		},
		"action": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"equal", "iequal", "regex", "absent"}, false),
						ForceNew:     true,
					},
					"value": {
						Type:     schema.TypeString,
						Optional: true,
						ForceNew: true,
						Computed: true,
					},
					"point": {
						Type:     schema.TypeMap,
						Optional: true,
						ForceNew: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"header": {
									Type:     schema.TypeList,
									Optional: true,
									ForceNew: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"method": {
									Type:     schema.TypeString,
									Optional: true,
									ForceNew: true,
									ValidateFunc: validation.StringInSlice([]string{"GET", "HEAD", "POST",
										"PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}, false),
								},
								"path": {
									Type:         schema.TypeInt,
									Optional:     true,
									ForceNew:     true,
									ValidateFunc: validation.IntBetween(0, 60),
								},
								"action_name": {
									Type:     schema.TypeString,
									Optional: true,
									ForceNew: true,
								},
								"action_ext": {
									Type:     schema.TypeString,
									Optional: true,
									ForceNew: true,
								},
								"query": {
									Type:     schema.TypeString,
									Optional: true,
									ForceNew: true,
								},
								"proto": {
									Type:         schema.TypeString,
									Optional:     true,
									ForceNew:     true,
									ValidateFunc: validation.StringInSlice([]string{"1.0", "1.1", "2.0", "3.0"}, false),
								},
								"scheme": {
									Type:         schema.TypeString,
									Optional:     true,
									ForceNew:     true,
									ValidateFunc: validation.StringInSlice([]string{"http", "https"}, true),
								},
								"uri": {
									Type:     schema.TypeString,
									Optional: true,
									ForceNew: true,
								},
								"instance": {
									Type:         schema.TypeInt,
									Optional:     true,
									ForceNew:     true,
									ValidateFunc: validation.IntAtLeast(-1),
								},
							},
						},
					},
				},
			},
		},
	}
	return &schema.Resource{
		Create: resourceWallarmCredentialStuffingModeCreate,
		Read:   resourceWallarmCredentialStuffingModeRead,
		Delete: resourceWallarmCredentialStuffingModeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmCredentialStuffingModeImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmCredentialStuffingModeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	mode := d.Get("mode").(string)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	resp, err := client.HintCreate(&wallarm.ActionCreate{
		Type:                "cred_stuff_mode",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Mode:                mode,
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
	d.Set("rule_id", resp.Body.ID)

	return resourceWallarmCredentialStuffingModeRead(d, m)
}

func resourceWallarmCredentialStuffingModeRead(d *schema.ResourceData, m interface{}) error {
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

	d.Set("mode", rule.Mode)
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
	d.Set("action", &actionsSet)

	return nil
}

func resourceWallarmCredentialStuffingModeDelete(d *schema.ResourceData, m interface{}) error {
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

func resourceWallarmCredentialStuffingModeImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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

	_, err = findRule(client, clientID, ruleID)
	if err != nil {
		return nil, err
	}

	d.Set("client_id", clientID)
	d.Set("rule_id", ruleID)

	return []*schema.ResourceData{d}, nil
}
