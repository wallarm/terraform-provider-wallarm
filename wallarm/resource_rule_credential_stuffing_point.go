package wallarm

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmCredentialStuffingPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmCredentialStuffingPointCreate,
		Read:   resourceWallarmCredentialStuffingPointRead,
		Delete: resourceWallarmCredentialStuffingPointDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmCredentialStuffingPointImport,
		},
		Schema: map[string]*schema.Schema{
			"rule_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"action_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"rule_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "The Client ID to perform changes",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
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
		},
	}
}

func resourceWallarmCredentialStuffingPointCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)
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
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	resp, err := client.HintCreate(&wallarm.ActionCreate{
		Type:                "credentials_point",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		Comment:             comment,
		VariativityDisabled: true,
		Point:               point,
		LoginPoint:          loginPoint,
		CredStuffType:       credStuffType,
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
	clientID := retrieveClientID(d, client)
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
	actionsSet := schema.Set{F: hashResponseActionDetails}
	for _, a := range rule.Action {
		acts, err := actionDetailsToMap(a)
		if err != nil {
			return err
		}
		actionsSet.Add(acts)
	}
	if err := d.Set("action", &actionsSet); err != nil {
		return err
	}

	return nil
}

func resourceWallarmCredentialStuffingPointDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
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
