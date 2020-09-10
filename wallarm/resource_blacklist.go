package wallarm

import (
	"fmt"
	"time"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmBlacklist() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmBlacklistCreate,
		Read:   resourceWallarmBlacklistRead,
		Update: resourceWallarmBlacklistUpdate,
		Delete: resourceWallarmBlacklistDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Client ID to perform changes",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v <= 0 {
						errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
					}
					return
				},
			},

			"ip_range": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"application": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"time": {
				Type:     schema.TypeInt,
				Required: true,
				// TODO: respect Date as an input
				// ValidateFunc: validation.ValidateRFC3339TimeString,
			},
			"reason": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Terraform managed Blacklist",
			},
		},
	}
}

func resourceWallarmBlacklistCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*wallarm.API)
	clientID := retrieveClientID(d, client)
	IPRange := d.Get("ip_range").([]interface{})
	ips := make([]string, len(IPRange))
	for i := range IPRange {
		ips[i] = IPRange[i].(string)
	}
	apps := []int{}
	if v, ok := d.GetOk("application"); ok {
		applications := v.([]interface{})
		apps = make([]int, len(applications))
		for i := range applications {
			apps[i] = applications[i].(int)
		}
	} else {
		pools := &wallarm.AppRead{
			Limit:  1000,
			Offset: 0,
			Filter: &wallarm.AppReadFilter{
				Clientid: []int{clientID},
			},
		}
		appResp, err := client.AppRead(pools)
		if err != nil {
			return err
		}

		apps = make([]int, len(appResp.Body))
		for i, app := range appResp.Body {
			apps[i] = app.ID
		}

	}

	expireTime := d.Get("time").(int)
	if expireTime == 0 {
		expireTime = 60045120
	}
	currTime := time.Now()
	shiftTime := currTime.Add(time.Minute * time.Duration(expireTime))
	unixTime := int(shiftTime.Unix())

	reason := d.Get("reason").(string)
	var bulk []wallarm.Bulk
	for _, ip := range ips {
		for _, app := range apps {
			b := wallarm.Bulk{
				IP:       ip,
				Poolid:   app,
				ExpireAt: unixTime,
				Reason:   reason,
				Clientid: clientID,
			}
			bulk = append(bulk, b)
		}
	}

	blacklistBody := wallarm.BlacklistCreate{Bulks: &bulk}

	if err := client.BlacklistCreate(&blacklistBody); err != nil {
		return err
	}

	d.SetId(reason)

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmBlacklistRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*wallarm.API)
	clientID := retrieveClientID(d, client)
	if err := client.BlacklistRead(clientID); err != nil {
		return err
	}
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmBlacklistUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceWallarmBlacklistCreate(d, m)
}

func resourceWallarmBlacklistDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*wallarm.API)
	clientID := retrieveClientID(d, client)
	if err := client.BlacklistDelete(clientID); err != nil {
		return err
	}

	return nil
}
