package wallarm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
					if v < 0 {
						errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
					}
					return
				},
			},
		},
	}
}

func resourceWallarmAppCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	appID := d.Get("app_id").(int)

	appBody := &wallarm.AppCreate{
		Name: name,
		AppFilter: &wallarm.AppFilter{
			ID:       appID,
			Clientid: clientID},
	}

	if err := client.AppCreate(appBody); err != nil {
		if errors.Is(err, wallarm.ErrExistingResource) {
			existingID := fmt.Sprintf("%d/%s/%d", clientID, name, appID)
			return ImportAsExistsError("wallarm_application", existingID)
		}
		return err
	}

	if err := d.Set("app_id", appID); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%d", clientID, name, appID)
	d.SetId(resID)

	return resourceWallarmAppRead(d, m)
}

func resourceWallarmAppRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
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
	found := false
	for _, app := range appResp.Body {
		if app.ID == appID {
			found = true
			if err = d.Set("name", name); err != nil {
				return err
			}
			if err = d.Set("app_id", app.ID); err != nil {
				return err
			}
			if err = d.Set("client_id", clientID); err != nil {
				return err
			}
		}
	}
	if !found {
		d.SetId("")
	}
	return nil
}

func resourceWallarmAppUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
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

		resID := fmt.Sprintf("%d/%s/%d", clientID, name, appID)
		d.SetId(resID)

		return resourceWallarmAppRead(d, m)
	}
	return resourceWallarmAppCreate(d, m)

}

func resourceWallarmAppDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
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
	client := m.(wallarm.API)
	idAttr := strings.SplitN(d.Id(), "/", 3)
	if len(idAttr) == 3 {
		clientID, err := strconv.Atoi(idAttr[0])
		if err != nil {
			return nil, err
		}
		name := idAttr[1]
		appID, err := strconv.Atoi(idAttr[2])
		if err != nil {
			return nil, err
		}

		appRead := &wallarm.AppRead{
			Limit:  1000,
			Offset: 0,
			Filter: &wallarm.AppReadFilter{
				Clientid: []int{clientID},
			},
		}
		appResp, err := client.AppRead(appRead)
		if err != nil {
			return nil, err
		}

		for _, app := range appResp.Body {
			if app.ID == appID {
				if err = d.Set("name", name); err != nil {
					return nil, err
				}
				if err = d.Set("app_id", app.ID); err != nil {
					return nil, err
				}
				if err = d.Set("client_id", clientID); err != nil {
					return nil, err
				}
			}
		}

		existingID := fmt.Sprintf("%d/%s/%d", clientID, name, appID)
		d.SetId(existingID)
	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{name}/{id}\"", d.Id())
	}

	if err := resourceWallarmAppRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
