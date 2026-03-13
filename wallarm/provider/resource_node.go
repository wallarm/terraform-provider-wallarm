package wallarm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmNode() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmNodeCreate,
		ReadContext:   resourceWallarmNodeRead,
		DeleteContext: resourceWallarmNodeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmNodeImport,
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

func resourceWallarmNodeCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	hostname := d.Get("hostname").(string)
	partnerMode := d.Get("partner_mode").(bool)

	nodeBody := &wallarm.NodeCreate{
		Hostname:    hostname,
		Type:        "cloud_node",
		Clientid:    clientID,
		PartnerMode: partnerMode,
	}

	nodeResp, err := client.NodeCreate(nodeBody)
	if err != nil {
		if errors.Is(err, wallarm.ErrExistingResource) {
			return diag.FromErr(ImportAsExistsError("wallarm_node", "{client_id}/{node_id}"))
		}
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d/%d", clientID, nodeResp.Body.ID))
	d.Set("node_id", nodeResp.Body.ID)
	d.Set("node_uuid", nodeResp.Body.UUID)
	d.Set("token", nodeResp.Body.Token)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmNodeRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	nodeID := d.Get("node_id").(int)

	nodes, err := client.NodeRead(clientID, "all")
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, node := range nodes.Body {
		if node.ID == nodeID {
			found = true
			d.Set("hostname", node.Hostname)
			d.Set("node_id", node.ID)
			d.Set("node_uuid", node.UUID)
			d.Set("token", node.Token)
			d.Set("client_id", node.Clientid)
			break
		}
	}

	if !found {
		log.Printf("[WARN] Node %d not found, removing from state", nodeID)
		d.SetId("")
	}

	return nil
}

func resourceWallarmNodeDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	nodeID := d.Get("node_id").(int)
	if err := client.NodeDelete(nodeID); err != nil {
		if isNotFoundError(err) {
			log.Printf("[WARN] Node %d has already been deleted", nodeID)
		} else {
			return diag.FromErr(err)
		}
	}
	return nil
}

// resourceWallarmNodeImport handles terraform import.
// Format: {client_id}/{node_id}
// Example: terraform import wallarm_node.my_node 8649/12345
func resourceWallarmNodeImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := apiClient(m)
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{client_id}/{node_id}\"", d.Id())
	}

	clientID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid client_id %q: %w", idParts[0], err)
	}
	nodeID, err := strconv.Atoi(idParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid node_id %q: %w", idParts[1], err)
	}

	nodes, err := client.NodeRead(clientID, "all")
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Body {
		if node.ID == nodeID {
			d.SetId(fmt.Sprintf("%d/%d", clientID, nodeID))
			d.Set("client_id", clientID)
			d.Set("node_id", node.ID)
			d.Set("hostname", node.Hostname)
			d.Set("node_uuid", node.UUID)
			d.Set("token", node.Token)
			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf("node with id %d not found for client %d", nodeID, clientID)
}
