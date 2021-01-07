package wallarm

import (
	"encoding/json"
	"fmt"
	"log"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmGlobalMode() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmGlobalModeCreate,
		Read:   resourceWallarmGlobalModeRead,
		Update: resourceWallarmGlobalModeUpdate,
		Delete: resourceWallarmGlobalModeDelete,

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

			"waf_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"default", "monitoring", "block"}, false),
			},

			"scanner_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},

			"rechecker_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},
		},
	}
}

func resourceWallarmGlobalModeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)

	wafMode := d.Get("waf_mode").(string)
	if wafMode == "block" {
		wafMode = "blocking"
	}

	scannerMode := d.Get("scanner_mode").(string)
	if scannerMode == "on" {
		scannerMode = "classic"
	}

	recheckerMode := d.Get("rechecker_mode").(string)

	wafmode := &wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{
			ID: clientID,
		},
		Fields: &wallarm.ClientFields{
			Mode:                wafMode,
			ScannerMode:         scannerMode,
			AttackRecheckerMode: recheckerMode,
		},
	}
	_, err := client.ClientUpdate(wafmode)
	if err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%s/%s", clientID, wafMode, scannerMode, recheckerMode)
	d.SetId(resID)

	d.Set("client_id", clientID)

	return resourceWallarmGlobalModeRead(d, m)
}

func resourceWallarmGlobalModeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	clientInfo := &wallarm.ClientRead{
		Filter: &wallarm.ClientReadFilter{
			Enabled: true,
			ClientFilter: wallarm.ClientFilter{
				ID: clientID},
		},
		Limit:  1000,
		Offset: 0,
	}
	res, err := client.ClientRead(clientInfo)
	if err != nil {
		return err
	}

	if len(res.Body) == 0 {
		body, err := json.Marshal(res)
		if err != nil {
			return err
		}
		log.Printf("[WARN] Client hasn't been found in API. Body: %s", body)

		d.SetId("")
		return nil
	}

	wafMode := res.Body[0].Mode
	if wafMode == "blocking" {
		wafMode = "block"
	}

	if err := d.Set("waf_mode", wafMode); err != nil {
		return err
	}

	scannerMode := res.Body[0].ScannerMode
	if scannerMode == "classic" {
		scannerMode = "on"
	}

	if err := d.Set("scanner_mode", scannerMode); err != nil {
		return err
	}

	recheckerMode := res.Body[0].AttackRecheckerMode

	if err := d.Set("rechecker_mode", recheckerMode); err != nil {
		return err
	}

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmGlobalModeUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceWallarmGlobalModeCreate(d, m)
}

func resourceWallarmGlobalModeDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
