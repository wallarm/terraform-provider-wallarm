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

func resourceWallarmCredentialStuffingRegex() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmCredentialStuffingRegexCreate,
		Read:   resourceWallarmCredentialStuffingRegexRead,
		Delete: resourceWallarmCredentialStuffingRegexDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmCredentialStuffingRegexImport,
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

func resourceWallarmCredentialStuffingRegexCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	comment := d.Get("comment").(string)
	regex := d.Get("regex").(string)
	credStuffType := d.Get("cred_stuff_type").(string)
	caseSensitive := d.Get("case_sensitive").(bool)
	loginRegex := d.Get("login_regex").(string)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return err
	}

	resp, err := client.HintCreate(&wallarm.ActionCreate{
		Type:                "credentials_regex",
		Clientid:            clientID,
		Action:              &action,
		Validated:           false,
		Comment:             comment,
		VariativityDisabled: true,
		Regex:               regex,
		LoginRegex:          loginRegex,
		CredStuffType:       credStuffType,
		CaseSensitive:       &caseSensitive,
	})
	if err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%d/%d", resp.Body.Clientid, resp.Body.ActionID, resp.Body.ID)
	d.SetId(resID)
	if err = d.Set("client_id", resp.Body.Clientid); err != nil {
		return err
	}
	if err = d.Set("rule_id", resp.Body.ID); err != nil {
		return err
	}

	return resourceWallarmCredentialStuffingRegexRead(d, m)
}

func resourceWallarmCredentialStuffingRegexRead(d *schema.ResourceData, m interface{}) error {
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

	if err = d.Set("regex", rule.Regex); err != nil {
		return err
	}
	if err = d.Set("cred_stuff_type", rule.CredStuffType); err != nil {
		return err
	}
	if err = d.Set("case_sensitive", rule.CaseSensitive); err != nil {
		return err
	}
	if err = d.Set("login_regex", rule.LoginRegex); err != nil {
		return err
	}
	if err = d.Set("rule_type", rule.Type); err != nil {
		return err
	}
	if err = d.Set("action_id", rule.ActionID); err != nil {
		return err
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
		return err
	}

	return nil
}

func resourceWallarmCredentialStuffingRegexDelete(d *schema.ResourceData, m interface{}) error {
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

	_, err = findRule(client, clientID, ruleID)
	if err != nil {
		return nil, err
	}

	if err = d.Set("client_id", clientID); err != nil {
		return nil, err
	}
	if err = d.Set("rule_id", ruleID); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
