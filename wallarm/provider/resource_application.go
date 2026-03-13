package wallarm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmApp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmAppCreate,
		ReadContext:   resourceWallarmAppRead,
		UpdateContext: resourceWallarmAppUpdate,
		DeleteContext: resourceWallarmAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmAppImport,
		},

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"app_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v != -1 && v <= 0 {
						errs = append(errs, fmt.Errorf("%q must be -1 (default application) or a positive integer, got: %d", key, v))
					}
					return
				},
			},
		},
	}
}

func resourceWallarmAppCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	appID := d.Get("app_id").(int)

	appBody := &wallarm.AppCreate{
		Name:     name,
		ID:       &appID,
		Clientid: clientID,
	}

	if err := client.AppCreate(appBody); err != nil {
		if errors.Is(err, wallarm.ErrExistingResource) {
			existingID := fmt.Sprintf("%d/%d", clientID, appID)
			return diag.FromErr(ImportAsExistsError("wallarm_application", existingID))
		}
		return diag.FromErr(err)
	}

	d.Set("app_id", appID)

	resID := fmt.Sprintf("%d/%d", clientID, appID)
	d.SetId(resID)

	return resourceWallarmAppRead(context.TODO(), d, m)
}

func resourceWallarmAppRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	appID := d.Get("app_id").(int)

	appRead := &wallarm.AppRead{
		Limit:  DefaultAPIListLimit,
		Offset: 0,
		Filter: &wallarm.AppReadFilter{
			Clientid: []int{clientID},
		},
	}
	appResp, err := client.AppRead(appRead)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, app := range appResp.Body {
		if app.ID != nil && *app.ID == appID {
			d.Set("name", app.Name)
			d.Set("app_id", app.ID)
			d.Set("client_id", clientID)
			return nil
		}
	}
	d.SetId("")
	return nil
}

func resourceWallarmAppUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	appID := d.Get("app_id").(int)
	name := d.Get("name").(string)

	if d.HasChange("name") {
		appBody := &wallarm.AppUpdate{
			Filter: &wallarm.AppUpdateFilter{
				ID:            appID,
				AppReadFilter: &wallarm.AppReadFilter{Clientid: []int{clientID}},
			},
			Fields: &wallarm.AppUpdateFields{
				Name: name,
			},
		}

		if err := client.AppUpdate(appBody); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceWallarmAppRead(context.TODO(), d, m)
}

func resourceWallarmAppDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	appID := d.Get("app_id").(int)

	// The default application (ID -1) cannot be deleted.
	if appID == -1 {
		return nil
	}

	appBody := &wallarm.AppDelete{
		Filter: &wallarm.AppFilter{
			ID:       appID,
			Clientid: clientID,
		},
	}

	if err := client.AppDelete(appBody); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWallarmAppImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 2)
	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{appID}\"", d.Id())
	}

	clientID, err := strconv.Atoi(idAttr[0])
	if err != nil {
		return nil, err
	}
	appID, err := strconv.Atoi(idAttr[1])
	if err != nil {
		return nil, err
	}

	d.Set("client_id", clientID)
	d.Set("app_id", appID)
	d.SetId(fmt.Sprintf("%d/%d", clientID, appID))

	if diags := resourceWallarmAppRead(context.TODO(), d, m); diags.HasError() {
		return nil, fmt.Errorf("%s", diags[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
