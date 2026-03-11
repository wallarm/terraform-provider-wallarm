package wallarm

import (
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWallarmApplications() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWallarmApplicationsRead,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"applications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWallarmApplicationsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

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

	apps := make([]interface{}, 0)
	for _, app := range appResp.Body {
		if app.Deleted || app.ID == nil {
			continue
		}
		apps = append(apps, map[string]interface{}{
			"app_id":    *app.ID,
			"name":      app.Name,
			"client_id": app.Clientid,
		})
	}

	d.SetId(fmt.Sprintf("applications_%d", clientID))
	if err := d.Set("applications", apps); err != nil {
		return fmt.Errorf("error setting applications: %w", err)
	}

	return nil
}
