package wallarm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmAppCreate,
		Read:   resourceWallarmAppRead,
		Update: resourceWallarmAppUpdate,
		Delete: resourceWallarmAppDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmAppImport,
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

func resourceWallarmAppCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
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
			return ImportAsExistsError("wallarm_application", existingID)
		}
		return err
	}

	d.Set("app_id", appID)

	resID := fmt.Sprintf("%d/%d", clientID, appID)
	d.SetId(resID)

	return resourceWallarmAppRead(d, m)
}

func resourceWallarmAppRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	appID := d.Get("app_id").(int)

	appRead := &wallarm.AppRead{
		Limit:  1000,
		Offset: 0,
		Filter: &wallarm.AppReadFilter{
			Clientid: []int{clientID},
		},
	}
	appResp, err := client.AppRead(appRead)
	if err != nil {
		return err
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

func resourceWallarmAppUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
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
			return err
		}
	}

	return resourceWallarmAppRead(d, m)
}

func resourceWallarmAppDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	appID := d.Get("app_id").(int)

	appBody := &wallarm.AppDelete{
		Filter: &wallarm.AppFilter{
			ID:       appID,
			Clientid: clientID,
		},
	}

	if err := client.AppDelete(appBody); err != nil {
		return err
	}

	return nil
}

func resourceWallarmAppImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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

	if err := resourceWallarmAppRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
