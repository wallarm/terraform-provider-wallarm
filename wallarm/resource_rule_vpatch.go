package wallarm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmVpatch() *schema.Resource {
	fields := map[string]*schema.Schema{
		"attack_type": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			Description: `Possible values: "any", "sqli", "rce", "crlf", "nosqli", "ptrav",
				"xxe", "ptrav", "xss", "scanner", "redir", "ldapi"`,
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
									Type:     schema.TypeInt,
									Optional: true,
									ForceNew: true,
									ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
										v := val.(int)
										if v < 0 || v > 60 {
											errs = append(errs, fmt.Errorf("%q must be between 0 and 60 inclusive, got: %d", key, v))
										}
										return
									},
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
									Type:     schema.TypeInt,
									Optional: true,
									ForceNew: true,
									ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
										v := val.(int)
										if v < -1 {
											errs = append(errs, fmt.Errorf("%q must be be greater than -1 inclusive, got: %d", key, v))
										}
										return
									},
								},
							},
						},
					},
				},
			},
		},
		"point": {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
	return &schema.Resource{
		Create: resourceWallarmVpatchCreate,
		Read:   resourceWallarmVpatchRead,
		Update: resourceWallarmVpatchUpdate,
		Delete: resourceWallarmVpatchDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmVpatchImport,
		},
		Schema: lo.Assign(fields, commonResourceRuleFields),
	}
}

func resourceWallarmVpatchCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	fields := getCommonResourceRuleFieldsDTOFromResourceData(d)
	attackType := d.Get("attack_type").(string)

	ps := d.Get("point").([]interface{})
	d.Set("point", ps)

	points, err := expandPointsToTwoDimensionalArray(ps)
	if err != nil {
		return err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	vp := &wallarm.ActionCreate{
		Type:                "vpatch",
		AttackType:          attackType,
		Clientid:            clientID,
		Action:              &action,
		Point:               points,
		Validated:           false,
		Comment:             fields.Comment,
		VariativityDisabled: true,
		Set:                 fields.Set,
		Active:              fields.Active,
		Title:               fields.Title,
		Mitigation:          fields.Mitigation,
	}

	actionResp, err := client.HintCreate(vp)
	if err != nil {
		return err
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)
	d.Set("attack_type", actionResp.Body.AttackType)
	d.Set("client_id", clientID)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return resourceWallarmVpatchRead(d, m)
}

func resourceWallarmVpatchRead(d *schema.ResourceData, m interface{}) error {
	return resourcerule.ResourceRuleWallarmRead(d, retrieveClientID(d), m.(wallarm.API),
		common.ReadOptionWithPoint)
}

func resourceWallarmVpatchDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	actionID := d.Get("action_id").(int)

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{"vpatch"},
			Clientid: []int{clientID},
			ID:       []int{actionID},
		},
		Limit:  1000,
		Offset: 0,
	}
	respRules, err := client.RuleRead(rule)
	if err != nil {
		return err
	}

	if len(respRules.Body) == 1 && respRules.Body[0].Hints == 1 && respRules.Body[0].GroupedHintsCount == 1 {
		if err := client.RuleDelete(actionID); err != nil {
			return err
		}
	} else {
		ruleID := d.Get("rule_id").(int)
		h := &wallarm.HintDelete{
			Filter: &wallarm.HintDeleteFilter{
				Clientid: []int{clientID},
				ID:       ruleID,
			},
		}

		if err := client.HintDelete(h); err != nil {
			return err
		}
	}
	return nil
}

func resourceWallarmVpatchUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	variativityDisabled, _ := d.Get("variativity_disabled").(bool)
	comment, _ := d.Get("comment").(string)
	_, err := client.HintUpdateV3(d.Get("rule_id").(int), &wallarm.HintUpdateV3Params{
		VariativityDisabled: lo.ToPtr(variativityDisabled),
		Comment:             lo.ToPtr(comment),
	})
	return err
}

// nolint:dupl
func resourceWallarmVpatchImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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
		d.Set("rule_type", "vpatch")

		existingID := fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
