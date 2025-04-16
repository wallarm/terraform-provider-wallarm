package wallarm

import (
	"fmt"
	"log"
	"time"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceWallarmVuln() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWallarmVulnRead,

		Schema: map[string]*schema.Schema{

			"client_id": defaultClientIDWithValidationSchema,

			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "open",
							ValidateFunc: validation.StringInSlice([]string{"open", "closed", "falsepositive"}, false),
						},

						"limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1000,
							ValidateFunc: validation.IntBetween(0, 1000),
						},

						"offset": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
					},
				},
			},

			"vuln": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vuln_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"wid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parameter": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"additional": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"exploit_example": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"detection_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWallarmVulnRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)

	// Prepare the filters to be applied to the search
	filter, err := expandWallarmVuln(d.Get("filter"))
	if err != nil {
		return err
	}

	vulns := make([]interface{}, 0)

	vuln := &wallarm.GetVulnRead{
		Limit:     filter.Limit,
		Offset:    filter.Offset,
		OrderBy:   "threat",
		OrderDesc: true,
		Filter: &wallarm.GetVulnFilter{
			Status: filter.Status,
		},
	}

	vulnPOST, err := client.GetVulnRead(vuln)
	if err != nil {
		return err
	}

	for _, v := range vulnPOST.Body {
		vulns = append(vulns, map[string]interface{}{
			"vuln_id":          v.ID,
			"wid":              v.Wid,
			"status":           v.Status,
			"type":             v.Type,
			"client_id":        v.Clientid,
			"method":           v.Method,
			"domain":           v.Domain,
			"path":             v.Path,
			"parameter":        v.Parameter,
			"title":            v.Type,
			"description":      v.Description,
			"additional":       v.Additional,
			"exploit_example":  v.ExploitExample,
			"detection_method": v.DetectionMethod,
		})
	}

	d.SetId(fmt.Sprintf("Vulns_%s", time.Now().UTC().String()))

	if err = d.Set("vuln", vulns); err != nil {
		return fmt.Errorf("Error setting Vulns: %s", err)
	}
	return nil
}

func expandWallarmVuln(d interface{}) (*searchFilterWallarmVuln, error) {
	cfg := d.([]interface{})
	log.Println("CFG", cfg)
	filter := &searchFilterWallarmVuln{
		Status: "open",
		Limit:  1000,
		Offset: 0,
	}
	if len(cfg) == 0 || cfg[0] == nil {
		return filter, nil
	}

	m := cfg[0].(map[string]interface{})

	status, ok := m["status"]
	if ok {
		filter.Status = status.(string)
	}

	limit, ok := m["limit"]
	if ok {
		filter.Limit = limit.(int)
	}

	offset, ok := m["offset"]
	if ok {
		filter.Offset = offset.(int)
	}

	return filter, nil
}

type searchFilterWallarmVuln struct {
	Status string
	Limit  int
	Offset int
}
