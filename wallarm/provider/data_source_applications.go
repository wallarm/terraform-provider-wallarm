package wallarm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"
)

func dataSourceWallarmApplications() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWallarmApplicationsRead,

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

func dataSourceWallarmApplicationsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	appRead := &wallarm.AppRead{
		Limit:  APIListLimit,
		Offset: 0,
		Filter: &wallarm.AppReadFilter{
			Clientid: []int{clientID},
		},
	}

	appResp, err := client.AppRead(appRead)
	if err != nil {
		return diag.FromErr(err)
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
		return diag.FromErr(fmt.Errorf("error setting applications: %w", err))
	}

	return nil
}
