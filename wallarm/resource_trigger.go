package wallarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmTriggerCreate,
		Read:   resourceWallarmTriggerRead,
		Update: resourceWallarmTriggerUpdate,
		Delete: resourceWallarmTriggerDelete,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"template_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{"user_created", "attacks_exceeded",
					"hits_exceeded", "incidents_exceeded", "vector_attack", "forced_browsing_started", "bruteforce_started", "bola_search_started",
					"blacklist_ip_added", "api_structure_changed", "attack_ip_grouping", "compromised_logins",
					"rogue_api_detected"}, false),
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Terraform managed trigger",
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "This trigger set by Terraform",
			},

			"filters": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 6,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_id": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{"ip_address", "pool", "attack_type",
								"domain", "target", "response_status", "url", "hint_tag", "pii", "change_type", "deviation_type",
								"api_spec_ids"}, false),
						},

						"operator": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"eq", "ne"}, false),
						},

						"value": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"actions": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"send_notification", "block_ips", "mark_as_brute", "group_attack_by_ip", "add_to_graylist"}, false),
						},

						"integration_id": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},

						"lock_time": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"lock_time_format": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Seconds",
							ValidateFunc: validation.StringInSlice([]string{"Seconds", "Minutes", "Hours", "Days", "Weeks", "Months"}, false),
						},
					},
				},
			},

			"threshold": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"period": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"operator": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"gt"}, false),
						},

						"count": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"time_format": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Seconds",
							ValidateFunc: validation.StringInSlice([]string{"Seconds", "Minutes"}, false),
						},
					},
				},
			},

			"trigger_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Second),
		},
	}
}

func resourceWallarmTriggerCreate(d *schema.ResourceData, m interface{}) error {
	var (
		err         error
		triggerResp *wallarm.TriggerCreateResp
	)

	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	name := d.Get("name").(string)
	comment := d.Get("comment").(string)
	templateID := d.Get("template_id").(string)
	enabled := d.Get("enabled").(bool)

	switch templateID {
	case "attacks_exceeded",
		"hits_exceeded", "incidents_exceeded",
		"vector_attack", "bruteforce_started":
		if _, ok := d.GetOk("threshold"); !ok {
			return fmt.Errorf(`"threshold" must be presented with the "%s" template`, templateID)
		}
	}

	filters, err := expandWallarmTriggerFilter(d.Get("filters"))
	if err != nil {
		return err
	}
	actions := expandWallarmTriggerAction(d.Get("actions"))
	if _, ok := d.GetOk("threshold"); ok {
		threshold, err := expandWallarmTriggerThreshold(d.Get("threshold"))
		if err != nil {
			return err
		}

		triggerBody := wallarm.TriggerCreate{
			Trigger: &wallarm.TriggerParam{
				Name:       name,
				Comment:    comment,
				TemplateID: templateID,
				Enabled:    enabled,
				Filters:    filters,
				Actions:    actions,
				Threshold:  threshold,
			},
		}

		triggerResp, err = client.TriggerCreate(&triggerBody, clientID)
		if err != nil {
			return err
		}

	} else {
		triggerBody := wallarm.TriggerCreate{
			Trigger: &wallarm.TriggerParam{
				Name:       name,
				Comment:    comment,
				TemplateID: templateID,
				Enabled:    enabled,
				Filters:    filters,
				Actions:    actions,
			},
		}

		triggerResp, err = client.TriggerCreate(&triggerBody, clientID)
		if err != nil {
			return err
		}
	}

	triggerID := triggerResp.ID
	d.Set("trigger_id", triggerID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, templateID, triggerID)
	d.SetId(resID)

	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		triggers, err := client.TriggerRead(clientID)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		for _, t := range triggers.Triggers {
			if t.ID == triggerID {
				if err = d.Set("trigger_id", t.ID); err != nil {
					return resource.NonRetryableError(err)
				}
				if err = d.Set("client_id", clientID); err != nil {
					return resource.NonRetryableError(err)
				}
				return resource.NonRetryableError(resourceWallarmTriggerRead(d, m))
			}
		}
		return resource.RetryableError(fmt.Errorf("can't find a trigger with ID: %d", triggerID))
	})
}

func resourceWallarmTriggerRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	triggerID := d.Get("trigger_id").(int)

	triggers, err := client.TriggerRead(clientID)
	if err != nil {
		return nil
	}

	for _, t := range triggers.Triggers {
		if t.ID == triggerID {
			d.Set("trigger_id", t.ID)
			d.Set("client_id", clientID)
			return nil
		}
	}
	d.SetId("")
	return nil
}

func resourceWallarmTriggerUpdate(d *schema.ResourceData, m interface{}) error {
	var (
		err         error
		triggerResp *wallarm.TriggerCreateResp
	)

	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	name := d.Get("name").(string)
	comment := d.Get("comment").(string)
	templateID := d.Get("template_id").(string)
	enabled := d.Get("enabled").(bool)
	triggerID := d.Get("trigger_id").(int)

	filters, err := expandWallarmTriggerFilter(d.Get("filters"))
	if err != nil {
		return err
	}

	actions := expandWallarmTriggerAction(d.Get("actions"))

	if _, ok := d.GetOk("threshold"); ok {
		threshold, err := expandWallarmTriggerThreshold(d.Get("threshold"))
		if err != nil {
			return err
		}

		triggerBody := wallarm.TriggerCreate{
			Trigger: &wallarm.TriggerParam{
				Name:       name,
				Comment:    comment,
				TemplateID: templateID,
				Enabled:    enabled,
				Filters:    filters,
				Actions:    actions,
				Threshold:  threshold,
			},
		}

		triggerResp, err = client.TriggerUpdate(&triggerBody, clientID, triggerID)
		if err != nil {
			return err
		}

	} else {
		triggerBody := wallarm.TriggerCreate{
			Trigger: &wallarm.TriggerParam{
				Name:       name,
				Comment:    comment,
				TemplateID: templateID,
				Enabled:    enabled,
				Filters:    filters,
				Actions:    actions,
			},
		}

		triggerResp, err = client.TriggerUpdate(&triggerBody, clientID, triggerID)
		if err != nil {
			return err
		}

	}

	triggerID = triggerResp.ID
	d.Set("trigger_id", triggerID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, templateID, triggerID)
	d.SetId(resID)

	return resourceWallarmTriggerRead(d, m)
}

func resourceWallarmTriggerDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	triggerID := d.Get("trigger_id").(int)

	if err := client.TriggerDelete(clientID, triggerID); err != nil {
		return err
	}

	return nil
}

func expandWallarmTriggerFilter(d interface{}) (*[]wallarm.TriggerFilters, error) {
	cfg := d.([]interface{})
	filters := []wallarm.TriggerFilters{}
	if len(cfg) == 0 || cfg[0] == nil {
		return &filters, nil
	}

	for _, conf := range cfg {

		m := conf.(map[string]interface{})
		t := wallarm.TriggerFilters{}
		filterID, ok := m["filter_id"]
		if ok {
			t.ID = filterID.(string)
		}

		operator, ok := m["operator"]
		if ok {
			t.Operator = operator.(string)
		}

		value, ok := m["value"]
		if ok {
			value := value.([]interface{})
			responseFallthrough := false

			switch filterID {
			case "pool", "api_spec_ids":
				var values []interface{}
				for _, v := range value {
					vString := v.(string)
					vInt, err := strconv.Atoi(vString)
					if err != nil {
						return nil, err
					}
					values = append(values, vInt)
				}
				t.Values = values
			case "response_status":
				var values []interface{}
				for _, v := range value {
					vString := v.(string)
					if vString[len(vString)-2:] != "xx" {
						vInt, err := strconv.Atoi(vString)
						if err != nil {
							return nil, err
						}
						values = append(values, vInt)
						responseFallthrough = true
					} else {
						values = append(values, v)
					}
				}
				t.Values = values
				fallthrough
			default:
				if !responseFallthrough {
					t.Values = value
				}
			}
		}

		filters = append(filters, t)
	}
	return &filters, nil
}

func expandWallarmTriggerAction(d interface{}) *[]wallarm.TriggerActions {
	cfg := d.([]interface{})
	actions := []wallarm.TriggerActions{}
	if len(cfg) == 0 || cfg[0] == nil {
		return &actions
	}

	for _, conf := range cfg {

		m := conf.(map[string]interface{})
		a := wallarm.TriggerActions{}
		actionID, ok := m["action_id"]
		if ok {
			a.ID = actionID.(string)
		}

		integrationID, ok := m["integration_id"]
		if ok {
			integrationID := integrationID.([]interface{})
			var integrationIDs []int
			for _, intID := range integrationID {
				integrationIDs = append(integrationIDs, intID.(int))
			}
			a.Params.IntegrationIds = integrationIDs
		}

		lockTime, ok := m["lock_time"]

		if ok && (a.ID == "block_ips" || a.ID == "add_to_graylist") {

			lockTimeInt := lockTime.(int)
			switch m["lock_time_format"] {
			case "Minutes":
				lockTimeInt *= 60
			case "Hours":
				lockTimeInt = lockTimeInt * 60 * 60
			case "Days":
				lockTimeInt = lockTimeInt * 60 * 60 * 24
			case "Weeks":
				lockTimeInt = lockTimeInt * 60 * 60 * 24 * 7
			case "Months":
				currTime := time.Now()
				shiftTime := currTime.AddDate(0, lockTimeInt, 0)
				lockTimeInt = int(shiftTime.Sub(currTime).Seconds())
			}
			if lockTimeInt == 0 {
				lockTimeInt = 315360000 // forever, our max time
			}
			a.Params.LockTime = lockTimeInt
		}

		actions = append(actions, a)

	}
	return &actions
}

func expandWallarmTriggerThreshold(cfg interface{}) (*wallarm.TriggerThreshold, error) {
	threshold := wallarm.TriggerThreshold{}
	m := cfg.(map[string]interface{})

	period, ok := m["period"]
	if ok {
		periodInt, err := strconv.Atoi(period.(string))
		if err != nil {
			return nil, err
		}
		threshold.Period = periodInt
	}

	timeFormat, ok := m["time_format"]
	if ok {
		if timeFormat == "Minutes" {
			threshold.Period = 60 * threshold.Period
		}
	}

	operator, ok := m["operator"]
	if ok {
		threshold.Operator = operator.(string)
	}

	threshold.AllowedOperators = []string{"gt"}

	count, ok := m["count"]
	if ok {
		countInt, err := strconv.Atoi(count.(string))
		if err != nil {
			return nil, err
		}
		threshold.Count = countInt
	}

	return &threshold, nil
}
