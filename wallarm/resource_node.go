package wallarm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmNode() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmNodeCreate,
		Read:   resourceWallarmNodeRead,
		Delete: resourceWallarmNodeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWallarmNodeImport,
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

			"hostname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"node_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"node_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWallarmNodeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	hostname := d.Get("hostname").(string)

	nodeBody := &wallarm.NodeCreate{
		Hostname: hostname,
		Type:     "cloud_node",
		Clientid: clientID,
	}

	d.SetId(hostname)

	nodeResp, err := client.NodeCreate(nodeBody)
	if err != nil {
		if errors.Is(err, wallarm.ErrExistingResource) {
			existingID := fmt.Sprintf("%d/%s", clientID, hostname)
			return ImportAsExistsError("wallarm_node", existingID)
		}
		return err
	}

	if err := d.Set("node_id", nodeResp.Body.ID); err != nil {
		return err
	}

	if err := d.Set("node_uuid", nodeResp.Body.UUID); err != nil {
		return err
	}

	if err := d.Set("token", nodeResp.Body.Token); err != nil {
		return err
	}

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmNodeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	hostname := d.Get("hostname").(string)

	nodes, err := client.NodeRead(clientID, "all")
	if err != nil {
		return err
	}

	for _, node := range nodes.Body {
		if node.Hostname == hostname {

			if err := d.Set("hostname", node.Hostname); err != nil {
				return err
			}

			if err := d.Set("node_id", node.ID); err != nil {
				return err
			}

			if err := d.Set("node_uuid", node.UUID); err != nil {
				return err
			}

			if err := d.Set("token", node.Token); err != nil {
				return err
			}

			if err := d.Set("client_id", node.Clientid); err != nil {
				return err
			}
		}

	}

	return nil
}

func resourceWallarmNodeDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	nodeID := d.Get("node_id").(int)
	if err := client.NodeDelete(nodeID); err != nil {
		return err
	}
	return nil
}

func resourceWallarmNodeImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(wallarm.API)
	idAttr := strings.SplitN(d.Id(), "/", 2)
	if len(idAttr) == 2 {
		clientID, err := strconv.Atoi(idAttr[0])
		if err != nil {
			return nil, err
		}
		hostname := idAttr[1]

		d.Set("hostname", hostname)
		nodes, err := client.NodeRead(clientID, "all")
		if err != nil {
			return nil, err
		}

		for _, node := range nodes.Body {
			if node.Hostname == hostname {

				if err := d.Set("hostname", node.Hostname); err != nil {
					return nil, err
				}

				if err := d.Set("node_id", node.ID); err != nil {
					return nil, err
				}

				if err := d.Set("node_uuid", node.UUID); err != nil {
					return nil, err
				}

				if err := d.Set("token", node.Token); err != nil {
					return nil, err
				}

				if err := d.Set("client_id", node.Clientid); err != nil {
					return nil, err
				}
			}

		}

		existingID := fmt.Sprintf("%d/%s", clientID, hostname)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{hostname}\"", d.Id())
	}

	resourceWallarmNodeRead(d, m)

	return []*schema.ResourceData{d}, nil
}
