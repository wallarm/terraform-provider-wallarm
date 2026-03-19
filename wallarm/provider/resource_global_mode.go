package wallarm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmGlobalMode() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmGlobalModeCreate,
		ReadContext:   resourceWallarmGlobalModeRead,
		UpdateContext: resourceWallarmGlobalModeUpdate,
		DeleteContext: resourceWallarmGlobalModeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			// wallarm_mode settings (PUT /v2/client/{id}/rules/wallarm_mode)
			"filtration_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "default",
				ValidateFunc: validation.StringInSlice([]string{"default", "monitoring", "block", "safe_blocking", "off"}, false),
			},

			// Client-level rechecker setting (ClientUpdate API)
			"rechecker_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},

			// overlimit_res_settings (PUT /v2/client/{id}/rules/overlimit_res_settings)
			"overlimit_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Time limit for single request processing (ms). Range: 0-10000.",
				ValidateFunc: validation.IntBetween(0, 10000),
			},
			"overlimit_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Action when overlimit_time is exceeded: blocking or monitoring.",
				ValidateFunc: validation.StringInSlice([]string{"blocking", "monitoring"}, false),
			},
		},
	}
}

func resourceWallarmGlobalModeCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update wallarm_mode
	filtrationMode := d.Get("filtration_mode").(string)
	if _, err := client.WallarmModeUpdate(&wallarm.WallarmModeParams{Mode: filtrationMode}, clientID); err != nil {
		return diag.FromErr(err)
	}

	// Update rechecker_mode
	recheckerMode := d.Get("rechecker_mode").(string)
	mode := &wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{
			ID: clientID,
		},
		Fields: &wallarm.ClientFields{
			AttackRecheckerMode: recheckerMode,
		},
	}
	if _, err := client.ClientUpdate(mode); err != nil {
		return diag.FromErr(err)
	}

	// Update overlimit_res_settings
	if err := updateOverlimitResSettings(d, client, clientID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d/global_mode", clientID))
	d.Set("client_id", clientID)

	return resourceWallarmGlobalModeRead(context.TODO(), d, m)
}

func resourceWallarmGlobalModeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)

	// Parse client_id from composite ID on import.
	// ID format: "{clientID}/global_mode"
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id := d.Id(); strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) == 2 {
			var parsed int
			if _, err := fmt.Sscanf(parts[0], "%d", &parsed); err == nil {
				clientID = parsed
			}
		}
	}

	// Read wallarm_mode
	wallarmModeResp, err := client.WallarmModeRead(clientID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("filtration_mode", wallarmModeResp.Body.Mode)

	// Read rechecker_mode from client info
	clientInfo := &wallarm.ClientRead{
		Filter: &wallarm.ClientReadFilter{
			Enabled: true,
			ClientFilter: wallarm.ClientFilter{
				ID: clientID,
			},
		},
		Limit:  APIListLimit,
		Offset: 0,
	}
	otherModesResp, err := client.ClientRead(clientInfo)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(otherModesResp.Body) == 0 {
		log.Printf("[WARN] Client %d not found in API, removing from state", clientID)
		d.SetId("")
		return nil
	}
	d.Set("rechecker_mode", otherModesResp.Body[0].AttackRecheckerMode)

	// Read overlimit_res_settings
	overlimitResp, err := client.OverlimitResSettingsRead(clientID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("overlimit_time", overlimitResp.Body.OverlimitTime)
	d.Set("overlimit_mode", overlimitResp.Body.Mode)

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmGlobalModeUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("filtration_mode") {
		filtrationMode := d.Get("filtration_mode").(string)
		if _, err := client.WallarmModeUpdate(&wallarm.WallarmModeParams{Mode: filtrationMode}, clientID); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("rechecker_mode") {
		recheckerMode := d.Get("rechecker_mode").(string)
		mode := &wallarm.ClientUpdate{
			Filter: &wallarm.ClientFilter{
				ID: clientID,
			},
			Fields: &wallarm.ClientFields{
				AttackRecheckerMode: recheckerMode,
			},
		}
		if _, err := client.ClientUpdate(mode); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("overlimit_time") || d.HasChange("overlimit_mode") {
		if err := updateOverlimitResSettings(d, client, clientID); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceWallarmGlobalModeRead(context.TODO(), d, m)
}

func resourceWallarmGlobalModeDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Global settings are a singleton — cannot be deleted, only modified.
	return nil
}

// updateOverlimitResSettings sends the current overlimit_time and overlimit_mode
// to the API. Both fields are always sent together since the API expects the
// full overlimit_res_settings object.
func updateOverlimitResSettings(d *schema.ResourceData, client wallarm.API, clientID int) error {
	overlimitTime := d.Get("overlimit_time").(int)
	overlimitMode := d.Get("overlimit_mode").(string)

	// Skip if neither field is configured (both at zero values).
	if overlimitTime == 0 && overlimitMode == "" {
		return nil
	}

	_, err := client.OverlimitResSettingsUpdate(&wallarm.OverlimitResSettingsParams{
		OverlimitTime: overlimitTime,
		Mode:          overlimitMode,
	}, clientID)
	return err
}
