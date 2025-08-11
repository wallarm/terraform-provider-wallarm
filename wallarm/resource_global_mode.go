package wallarm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/wallarm/wallarm-go"

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
			"client_id": defaultClientIDWithValidationSchema,

			"filtration_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "default",
				ValidateFunc: validation.StringInSlice([]string{"default", "monitoring", "block", "safe_blocking", "off"}, false),
			},

			"rechecker_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},
		},
	}
}

func resourceWallarmGlobalModeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	filtrationMode := d.Get("filtration_mode").(string)

	_, err := client.WallarmModeUpdate(&wallarm.WallarmModeParams{Mode: filtrationMode}, clientID)
	if err != nil {
		return err
	}

	recheckerMode := d.Get("rechecker_mode").(string)

	mode := &wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{
			ID: clientID,
		},
		Fields: &wallarm.ClientFields{
			AttackRecheckerMode: recheckerMode,
		},
	}
	_, err = client.ClientUpdate(mode)
	if err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%s/%s", clientID, filtrationMode, "", recheckerMode)
	d.SetId(resID)

	d.Set("client_id", clientID)

	return resourceWallarmGlobalModeRead(d, m)
}

func resourceWallarmGlobalModeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	wallarmModeResp, err := client.WallarmModeRead(clientID)
	if err != nil {
		return err
	}
	if wallarmModeResp.Status != http.StatusOK {
		body, err := json.Marshal(wallarmModeResp)
		if err != nil {
			return err
		}
		log.Printf("[WARN] Couldn't fetch wallarm_mode. Body: %s", body)

		d.SetId("")
		return nil
	}

	filtrationMode := wallarmModeResp.Body.Mode
	d.Set("filtration_mode", filtrationMode)

	clientInfo := &wallarm.ClientRead{
		Filter: &wallarm.ClientReadFilter{
			Enabled: true,
			ClientFilter: wallarm.ClientFilter{
				ID: clientID},
		},
		Limit:  1000,
		Offset: 0,
	}

	otherModesResp, err := client.ClientRead(clientInfo)
	if err != nil {
		return err
	}
	if len(otherModesResp.Body) == 0 {
		body, err := json.Marshal(otherModesResp)
		if err != nil {
			return err
		}
		log.Printf("[WARN] Client hasn't been found in API. Body: %s", body)

		d.SetId("")
		return nil
	}

	recheckerMode := otherModesResp.Body[0].AttackRecheckerMode

	d.Set("rechecker_mode", recheckerMode)

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmGlobalModeUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceWallarmGlobalModeCreate(d, m)
}

func resourceWallarmGlobalModeDelete(_ *schema.ResourceData, _ interface{}) error {
	return nil
}
