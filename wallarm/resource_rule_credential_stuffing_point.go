package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmCredentialStuffingPoint() *schema.Resource {
	fields := map[string]*schema.Schema{
		"point": {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
		},
		"login_point": {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
		},
		"cred_stuff_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "default",
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"custom", "default"}, false),
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
	d.Set("point", rule.Point)
	d.Set("login_point", rule.LoginPoint)
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
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{VariativityDisabled: lo.ToPtr(true)})
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
		d.Set("action", &actionsSet)
	}

	pointInterface := (*actionHints.Body)[0].Point
	point := wrapPointElements(pointInterface)
	d.Set("point", point)
	pointInterface = (*actionHints.Body)[0].LoginPoint
	point = wrapPointElements(pointInterface)
	d.Set("login_point", point)

	existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
	d.SetId(existingID)

	return []*schema.ResourceData{d}, nil
}
