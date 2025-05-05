package wallarm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/wallarm/wallarm-go"

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
			"client_id": defaultClientIDWithValidationSchema,

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

			"partner_mode": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Partner mode",
			},
		},
	}
}

func resourceWallarmNodeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	hostname := d.Get("hostname").(string)
	partnerMode := d.Get("partner_mode").(bool)

	nodeBody := &wallarm.NodeCreate{
		Hostname:    hostname,
		Type:        "cloud_node",
		Clientid:    clientID,
		PartnerMode: partnerMode,
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

	d.Set("node_id", nodeResp.Body.ID)
	d.Set("node_uuid", nodeResp.Body.UUID)
	d.Set("token", nodeResp.Body.Token)

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmNodeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	hostname := d.Get("hostname").(string)

	nodes, err := client.NodeRead(clientID, "all")
	if err != nil {
		return err
	}

	found := false
	for _, node := range nodes.Body {
		if node.Hostname == hostname {
			found = true
			d.Set("hostname", node.Hostname)

			d.Set("node_id", node.ID)

			d.Set("node_uuid", node.UUID)

			d.Set("token", node.Token)

			d.Set("client_id", node.Clientid)
		}

	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceWallarmNodeDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	nodeID := d.Get("node_id").(int)
	if err := client.NodeDelete(nodeID); err != nil {
		isNotFoundErr, err2 := isNotFoundError(err)
		if err2 != nil {
			return err2
		}

		if isNotFoundErr {
			fmt.Print("Resource has already been deleted")
		} else {
			return err
		}
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

				d.Set("hostname", node.Hostname)

				d.Set("node_id", node.ID)

				d.Set("node_uuid", node.UUID)

				d.Set("token", node.Token)

				d.Set("client_id", node.Clientid)
			}

		}

		existingID := fmt.Sprintf("%d/%s", clientID, hostname)
		d.SetId(existingID)

	} else {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{hostname}\"", d.Id())
	}

	if err := resourceWallarmNodeRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
