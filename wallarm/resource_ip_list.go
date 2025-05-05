package wallarm

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmIPList(listType wallarm.IPListType) *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmIPListCreate(listType),
		Read:   resourceWallarmIPListRead(listType),
		Update: resourceWallarmIPListUpdate(listType),
		Delete: resourceWallarmIPListDelete(listType),
		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
			"ip_range": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"application": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"time_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Minutes", "RFC3339", "Hours", "Days", "Weeks", "Months"}, false),
			},
			"time": {
				Type:     schema.TypeString,
				Required: true,
			},
			"reason": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Terraform managed Denylist",
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

func resourceWallarmIPListCreate(listType wallarm.IPListType) schema.CreateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		client := m.(wallarm.API)
		clientID := retrieveClientID(d)

		var ips []string
		v := d.Get("ip_range")
		IPRange := v.([]interface{})
		for _, v := range IPRange {
			if strings.Contains(v.(string), "/") {
				subNetwork, err := strconv.Atoi(strings.Split(v.(string), "/")[1])
				if err != nil {
					return fmt.Errorf("cannot parse subnet to integer. must be the number, got %v", err)
				}
				if subNetwork < 20 {
					return fmt.Errorf("subnet must be >= /20, got %v", subNetwork)
				}
			}
			ips = append(ips, v.(string))
		}

		var apps []int
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

		var unixTime int
		switch d.Get("time_format") {
		case Minutes:
			expireTime, err := strconv.Atoi(d.Get("time").(string))
			if err != nil {
				return fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Minute`, got %v", err)
			}
			if expireTime == 0 {
				expireTime = 60045120
			}
			currTime := time.Now()
			shiftTime := currTime.Add(time.Minute * time.Duration(expireTime))
			unixTime = int(shiftTime.Unix())
		case "Hours":
			expireTime, err := strconv.Atoi(d.Get("time").(string))
			if err != nil {
				return fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Hours`, got %v", err)
			}
			if expireTime == 0 {
				expireTime = 60045120
			}
			currTime := time.Now()
			shiftTime := currTime.Add(time.Hour * time.Duration(expireTime))
			unixTime = int(shiftTime.Unix())
		case "Days":
			expireTime, err := strconv.Atoi(d.Get("time").(string))
			if err != nil {
				return fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Days`, got %v", err)
			}
			if expireTime == 0 {
				expireTime = 60045120
			}
			currTime := time.Now()
			shiftTime := currTime.Add(24 * time.Hour * time.Duration(expireTime))
			unixTime = int(shiftTime.Unix())
		case "Weeks":
			expireTime, err := strconv.Atoi(d.Get("time").(string))
			if err != nil {
				return fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Weeks`, got %v", err)
			}
			if expireTime == 0 {
				expireTime = 60045120
			}
			currTime := time.Now()
			shiftTime := currTime.Add(7 * 24 * time.Hour * time.Duration(expireTime))
			unixTime = int(shiftTime.Unix())
		case "Months":
			expireTime, err := strconv.Atoi(d.Get("time").(string))
			if err != nil {
				return fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Months`, got %v", err)
			}
			if expireTime == 0 {
				expireTime = 60045120
			}
			currTime := time.Now()
			shiftTime := currTime.AddDate(0, expireTime, 0)
			unixTime = int(shiftTime.Unix())
		case "RFC3339":
			expireTime, err := time.Parse(time.RFC3339, d.Get("time").(string))
			if err != nil {
				return fmt.Errorf("cannot parse time to integer. must be the valid RFC3339 time when `time_format` equals `RFC3339`, got %v.\nExample: 2006-01-02T15:04:05+07:00", err)
			}
			unixTime = int(expireTime.Unix())
		}

		reason := d.Get("reason").(string)

		for _, ip := range ips {
			params := wallarm.IPRuleCreationParams{
				ExpiredAt: unixTime,
				List:      listType,
				Pools:     apps,
				Reason:    reason,
				RuleType:  "ip_range",
				Subnet:    ip,
			}

			if err := client.IPListCreate(clientID, params); err != nil {
				return err
			}
		}

		d.SetId(reason)

		return resourceWallarmIPListRead(listType)(d, m)
	}
}

func resourceWallarmIPListRead(listType wallarm.IPListType) schema.ReadFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		client := m.(wallarm.API)
		clientID := retrieveClientID(d)
		IPRange := d.Get("ip_range").([]interface{})
		ips := make([]string, len(IPRange))
		for i := range IPRange {
			ips[i] = IPRange[i].(string)
		}

		var apps []int
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

		var ipListFromTerraform []struct {
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
						ipListFromTerraform = append(ipListFromTerraform, struct {
							IP          string
							Application int
						}{subnetIP, app})
					}
				} else {
					ipListFromTerraform = append(ipListFromTerraform, struct {
						IP          string
						Application int
					}{ip, app})
				}
			}
		}

		derivedIPaddr := make([]string, len(ipListFromTerraform))
		for k, b := range ipListFromTerraform {
			derivedIPaddr[k] = b.IP
		}

		ipListsFromAPI, err := client.IPListRead(listType, clientID)
		if err != nil {
			return err
		}

		addrIDs := make([]interface{}, 0)
		found := false
		for _, ipList := range ipListsFromAPI {
			if wallarm.Contains(derivedIPaddr, strings.Split(ipList.Subnet, "/")[0]) {
				found = true
				addrIDs = append(addrIDs, map[string]interface{}{
					"ip_addr": ipList.Subnet,
					"ip_id":   ipList.ID,
				})
			}
		}
		if !found {
			d.SetId("")
			return nil
		}

		if err = d.Set("address_id", addrIDs); err != nil {
			return fmt.Errorf("cannot set content for ip_range: %v", err)
		}

		d.Set("client_id", clientID)

		return nil
	}
}

func resourceWallarmIPListUpdate(listType wallarm.IPListType) schema.UpdateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		if err := resourceWallarmIPListDelete(listType)(d, m); err != nil {
			return err
		}
		return resourceWallarmIPListCreate(listType)(d, m)
	}
}

func resourceWallarmIPListDelete(listType wallarm.IPListType) schema.DeleteFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		client := m.(wallarm.API)
		clientID := retrieveClientID(d)
		addrIDInterface := d.Get("address_id").([]interface{})
		addrIDs := make([]map[string]interface{}, len(addrIDInterface))
		for i := range addrIDInterface {
			addrIDs[i] = addrIDInterface[i].(map[string]interface{})
		}

		var derivedIDs []int
		for _, id := range addrIDs {
			derivedIDs = append(derivedIDs, id["ip_id"].(int))
		}

		if len(derivedIDs) == 0 {
			derivedIDs = append(derivedIDs, 0)
		}

		if err := client.IPListDelete(listType, clientID, derivedIDs); err != nil {
			return err
		}

		return nil
	}
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
