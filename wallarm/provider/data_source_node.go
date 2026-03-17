package wallarm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/wallarm/wallarm-go"
)

func dataSourceWallarmNode() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWallarmNodeRead,

		Schema: map[string]*schema.Schema{

			"client_id": defaultClientIDWithValidationSchema,

			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"cloud_node", "node", "fast_node"}, false),
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"client_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"instance_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"active_instance_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"token": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"requests_amount": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"proton": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"lom": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWallarmNodeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	// Prepare the filters to be applied to the search
	filter := expandWallarmNode(d.Get("filter"))
	if filter.Type == "" {
		filter.Type = "all"
	}

	nodes := make([]interface{}, 0)
	var node *wallarm.NodeRead
	var nodePOST *wallarm.NodeReadPOST
	nodeReadBody := wallarm.NodeReadByFilter{
		Filter:    &wallarm.NodeFilter{},
		Limit:     DefaultAPIListLimit,
		Offset:    0,
		OrderBy:   "id",
		OrderDesc: false,
	}
	var POST bool
	switch {
	case filter.UUID != "":
		nodeReadBody.Filter.UUID = filter.UUID
		nodePOST, err = client.NodeReadByFilter(&nodeReadBody)
		if err != nil {
			return diag.FromErr(err)
		}
		POST = true
	case filter.Hostname != "":
		nodeReadBody.Filter.Hostname = filter.Hostname
		nodePOST, err = client.NodeReadByFilter(&nodeReadBody)
		if err != nil {
			return diag.FromErr(err)
		}
		POST = true
	default:
		node, err = client.NodeRead(clientID, filter.Type)
		if err != nil {
			return diag.FromErr(err)
		}
		POST = false
	}
	if POST {
		for _, b := range nodePOST.Body {
			nodes = append(nodes, map[string]interface{}{
				"type":      b.Type,
				"uuid":      b.ID,
				"hostname":  b.Hostname,
				"enabled":   b.Enabled,
				"client_id": b.Clientid,
				"active":    b.Active,
				"ip":        b.IP,
				"proton":    b.ProtondbVersion,
				"lom":       b.LomVersion,
			})
		}
	} else {
		for _, b := range node.Body {
			nodes = append(nodes, map[string]interface{}{
				"type":                  b.Type,
				"id":                    b.ID,
				"uuid":                  b.UUID,
				"hostname":              b.Hostname,
				"enabled":               b.Enabled,
				"client_id":             b.Clientid,
				"active":                b.Active,
				"ip":                    interfaceToString(b.IP),
				"proton":                interfaceToInt(b.ProtondbVersion),
				"lom":                   interfaceToInt(b.LomVersion),
				"instance_count":        b.InstanceCount,
				"active_instance_count": b.ActiveInstanceCount,
				"token":                 b.Token,
				"requests_amount":       b.RequestsAmount,
			})
		}
	}

	if err := d.Set("nodes", nodes); err != nil {
		return diag.FromErr(fmt.Errorf("error setting nodes: %w", err))
	}

	d.SetId(fmt.Sprintf("nodes_%d_%s", clientID, filter.Type))
	return nil
}

func expandWallarmNode(d interface{}) *searchFilterWallarmNode {
	cfg := d.([]interface{})
	filter := &searchFilterWallarmNode{}
	if len(cfg) == 0 || cfg[0] == nil {
		return filter
	}

	m := cfg[0].(map[string]interface{})

	typeNode, ok := m["type"]
	if ok {
		filter.Type = typeNode.(string)
	}

	enabled, ok := m["enabled"]
	if ok {
		filter.Enabled = enabled.(bool)
	}

	hostname, ok := m["hostname"]
	if ok {
		filter.Hostname = hostname.(string)
	}

	uuid, ok := m["uuid"]
	if ok {
		filter.UUID = uuid.(string)
	}

	return filter
}

type searchFilterWallarmNode struct {
	Type     string
	Enabled  bool
	Hostname string
	UUID     string
}
