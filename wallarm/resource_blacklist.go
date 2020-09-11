package wallarm

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmBlacklist() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmBlacklistCreate,
		Read:   resourceWallarmBlacklistRead,
		Update: resourceWallarmBlacklistUpdate,
		Delete: resourceWallarmBlacklistDelete,

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

			"ip_range": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"application": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"time": {
				Type:     schema.TypeInt,
				Required: true,
				// TODO: respect Date as an input
				// ValidateFunc: validation.ValidateRFC3339TimeString,
			},
			"reason": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Terraform managed Blacklist",
			},
			"address_id": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_addr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"app": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ip_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceWallarmBlacklistCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*wallarm.API)
	clientID := retrieveClientID(d, client)

	var ips []string
	if v, ok := d.GetOk("ip_range"); ok {
		IPRange := v.([]interface{})
		for _, v := range IPRange {
			ips = append(ips, v.(string))
		}
	} else {
		return errors.New(`"ip_range" must be specified, got an empty atribute`)
	}

	apps := []int{}
	if v, ok := d.GetOk("application"); ok {
		applications := v.([]interface{})
		apps = make([]int, len(applications))
		for i := range applications {
			apps[i] = applications[i].(int)
		}
	} else {
		pools := &wallarm.AppRead{
			Limit:  1000,
			Offset: 0,
			Filter: &wallarm.AppReadFilter{
				Clientid: []int{clientID},
			},
		}
		appResp, err := client.AppRead(pools)
		if err != nil {
			return err
		}
		apps = make([]int, len(appResp.Body))
		for i, app := range appResp.Body {
			apps[i] = app.ID
		}
	}

	expireTime := d.Get("time").(int)
	if expireTime == 0 {
		expireTime = 60045120
	}
	currTime := time.Now()
	shiftTime := currTime.Add(time.Minute * time.Duration(expireTime))
	unixTime := int(shiftTime.Unix())

	reason := d.Get("reason").(string)
	var bulk []wallarm.Bulk
	for _, ip := range ips {
		for _, app := range apps {
			b := wallarm.Bulk{
				IP:       ip,
				Poolid:   app,
				ExpireAt: unixTime,
				Reason:   reason,
				Clientid: clientID,
			}
			bulk = append(bulk, b)
		}
	}

	blacklistBody := wallarm.BlacklistCreate{Bulks: &bulk}

	if err := client.BlacklistCreate(&blacklistBody); err != nil {
		return err
	}

	d.SetId(reason)

	return resourceWallarmBlacklistRead(d, m)
}

func resourceWallarmBlacklistRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*wallarm.API)
	clientID := retrieveClientID(d, client)
	IPRange := d.Get("ip_range").([]interface{})
	ips := make([]string, len(IPRange))
	for i := range IPRange {
		ips[i] = IPRange[i].(string)
	}

	apps := []int{}
	if v, ok := d.GetOk("application"); ok {
		applications := v.([]interface{})
		apps = make([]int, len(applications))
		for i := range applications {
			apps[i] = applications[i].(int)
		}
	} else {
		pools := &wallarm.AppRead{
			Limit:  1000,
			Offset: 0,
			Filter: &wallarm.AppReadFilter{
				Clientid: []int{clientID},
			},
		}
		appResp, err := client.AppRead(pools)
		if err != nil {
			return err
		}
		apps = make([]int, len(appResp.Body))
		for i, app := range appResp.Body {
			apps[i] = app.ID
		}
	}

	var blacklistFromTerraform []struct {
		IP          string
		Application int
	}

	for _, ip := range ips {
		for _, app := range apps {
			if strings.Contains(ip, "/") {
				subnet, err := hosts(ip)
				if err != nil {
					return err
				}
				for _, subnetIP := range subnet {
					blacklistFromTerraform = append(blacklistFromTerraform, struct {
						IP          string
						Application int
					}{subnetIP, app})
				}
			} else {
				blacklistFromTerraform = append(blacklistFromTerraform, struct {
					IP          string
					Application int
				}{ip, app})
			}
		}
	}

	derivedIPaddr := make([]string, len(blacklistFromTerraform))
	for k, b := range blacklistFromTerraform {
		derivedIPaddr[k] = b.IP
	}

	if err := d.Set("ip_range", IPRange); err != nil {
		return fmt.Errorf("cannot set content for ip_range: %v", err)
	}

	blacklistFromAPI, err := client.BlacklistRead(clientID)
	if err != nil {
		return err
	}

	addrAppIDs := make([]interface{}, 0)
	for _, maxEntry := range blacklistFromAPI {
		for _, IPcontainer := range maxEntry.Body.Objects {
			if wallarm.Contains(derivedIPaddr, IPcontainer.IP) {
				addrAppIDs = append(addrAppIDs, map[string]interface{}{
					"ip_addr": IPcontainer.IP,
					"app":     IPcontainer.Poolid,
					"ip_id":   IPcontainer.ID,
				})
			}
		}
	}

	if err := d.Set("address_id", addrAppIDs); err != nil {
		return fmt.Errorf("cannot set content for ip_range: %v", err)
	}

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmBlacklistUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceWallarmBlacklistCreate(d, m)
}

func resourceWallarmBlacklistDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*wallarm.API)
	clientID := retrieveClientID(d, client)
	addrIDInterface := d.Get("address_id").([]interface{})
	addrIDs := make([]map[string]interface{}, len(addrIDInterface))
	for i := range addrIDInterface {
		addrIDs[i] = addrIDInterface[i].(map[string]interface{})
	}

	var derivedIDs []int
	for _, id := range addrIDs {
		derivedIDs = append(derivedIDs, id["ip_id"].(int))
	}

	if err := client.BlacklistDelete(clientID, derivedIDs); err != nil {
		return err
	}

	return nil
}

// Pull out the raw IP addresses from the Subnet.
func hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	return ips, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
