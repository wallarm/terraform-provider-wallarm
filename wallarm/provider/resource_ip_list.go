package wallarm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmIPList(listType wallarm.IPListType) *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmIPListCreate(listType),
		ReadContext:   resourceWallarmIPListRead(listType),
		UpdateContext: resourceWallarmIPListUpdate(listType),
		DeleteContext: resourceWallarmIPListDelete(listType),
		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
			"ip_range": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"country", "datacenter", "proxy_type"},
			},
			"country": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"ip_range", "datacenter", "proxy_type"},
			},
			"datacenter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"alibaba", "aws", "azure", "docean", "gce", "hetzner", "huawei", "ibm", "linode", "oracle", "ovh", "plusserver", "rackspace", "tencent"}, false),
				},
				ConflictsWith: []string{"ip_range", "country", "proxy_type"},
			},
			"proxy_type": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"ip_range", "country", "datacenter"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"MIP", "PUB", "WEB", "SES", "TOR", "VPN"}, false),
				},
			},
			"application": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"time_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Minutes", "RFC3339", "Hours", "Days", "Weeks", "Months", "Forever"}, true),
			},
			"time": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"reason": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Terraform managed IP list",
			},
			"address_id": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
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

func resourceWallarmIPListCreate(listType wallarm.IPListType) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(wallarm.API)
		clientID := retrieveClientID(d)

		rules, diags := buildRulesFromSchema(d)
		if diags != nil {
			return diags
		}
		if len(rules) == 0 {
			return diag.FromErr(fmt.Errorf("at least one of ip_range, country, datacenter, or proxy_type must be specified"))
		}

		var apps []int
		if v, ok := d.GetOk("application"); ok {
			applications := v.([]interface{})
			apps = make([]int, len(applications))
			for i := range applications {
				apps[i] = applications[i].(int)
			}
		} else {
			apps = []int{0} // 0 means all applications
		}

		unixTime, diags := parseExpireTime(d)
		if diags != nil {
			return diags
		}

		reason := d.Get("reason").(string)

		params := wallarm.AccessRuleCreateRequest{
			List:           listType,
			Force:          false,
			Reason:         reason,
			ApplicationIDs: apps,
			ExpiredAt:      unixTime,
			Rules:          rules,
		}

		if err := client.IPListCreate(clientID, params); err != nil {
			return diag.FromErr(err)
		}

		ruleType := ipListRuleTypes(rules)
		valuesHash := ipListValuesHash(rules)
		d.SetId(fmt.Sprintf("%d/%s/%s/%s", clientID, ipListFriendlyType(listType), ruleType, valuesHash))

		return resourceWallarmIPListRead(listType)(ctx, d, m)
	}
}

func resourceWallarmIPListRead(listType wallarm.IPListType) schema.ReadContextFunc {
	return func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(wallarm.API)
		clientID := retrieveClientID(d)

		// Build lookup sets for each rule type from the schema.
		subnetNets, subnetIPs := buildSubnetMatchers(d)
		countryLookup := buildStringLookup(d, "country")
		sourceLookup := buildStringLookup(d, "datacenter")
		proxyLookup := buildStringLookup(d, "proxy_type")

		ipListsFromAPI, err := client.IPListRead(listType, clientID)
		if err != nil {
			return diag.FromErr(err)
		}

		addrIDs := make([]interface{}, 0)
		found := false
		now := int(time.Now().Unix())
		for _, ipRule := range ipListsFromAPI {
			// Skip expired entries — the API may still return them but
			// they are effectively gone and should be removed from state.
			if ipRule.ExpiredAt > 0 && ipRule.ExpiredAt < now {
				continue
			}
			switch ipRule.RuleType {
			case "subnet":
				for _, val := range ipRule.Values {
					ipAddr := strings.Split(val, "/")[0]
					if subnetMatch(ipAddr, subnetNets, subnetIPs) {
						found = true
						addrIDs = append(addrIDs, map[string]interface{}{
							"rule_type": "subnet",
							"value":     val,
							"ip_id":     ipRule.ID,
						})
					}
				}
			case "location":
				for _, val := range ipRule.Values {
					if countryLookup[val] {
						found = true
						addrIDs = append(addrIDs, map[string]interface{}{
							"rule_type": "location",
							"value":     val,
							"ip_id":     ipRule.ID,
						})
					}
				}
			case "datacenter":
				for _, val := range ipRule.Values {
					if sourceLookup[val] {
						found = true
						addrIDs = append(addrIDs, map[string]interface{}{
							"rule_type": "datacenter",
							"value":     val,
							"ip_id":     ipRule.ID,
						})
					}
				}
			case "proxy_type":
				for _, val := range ipRule.Values {
					if proxyLookup[val] {
						found = true
						addrIDs = append(addrIDs, map[string]interface{}{
							"rule_type": "proxy_type",
							"value":     val,
							"ip_id":     ipRule.ID,
						})
					}
				}
			}
		}
		if !found {
			d.SetId("")
			return nil
		}

		if err = d.Set("address_id", addrIDs); err != nil {
			return diag.FromErr(fmt.Errorf("cannot set content for address_id: %v", err))
		}

		d.Set("client_id", clientID)

		return nil
	}
}

func resourceWallarmIPListUpdate(listType wallarm.IPListType) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		if err := resourceWallarmIPListDelete(listType)(ctx, d, m); err != nil {
			return err
		}
		if createErr := resourceWallarmIPListCreate(listType)(ctx, d, m); createErr != nil {
			return append(createErr, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "IP list entries were deleted but re-creation failed — manual intervention may be required",
			})
		}
		return nil
	}
}

func resourceWallarmIPListDelete(_ wallarm.IPListType) schema.DeleteContextFunc {
	return func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(wallarm.API)
		clientID := retrieveClientID(d)
		addrIDInterface := d.Get("address_id").([]interface{})
		addrIDs := make([]map[string]interface{}, len(addrIDInterface))
		for i := range addrIDInterface {
			addrIDs[i] = addrIDInterface[i].(map[string]interface{})
		}

		// Group IDs by rule_type for the delete request.
		ruleTypeIDs := make(map[string][]int)
		for _, entry := range addrIDs {
			ruleType := entry["rule_type"].(string)
			id := entry["ip_id"].(int)
			ruleTypeIDs[ruleType] = append(ruleTypeIDs[ruleType], id)
		}

		if len(ruleTypeIDs) == 0 {
			return nil
		}

		var deleteRules []wallarm.AccessRuleDeleteEntry
		for ruleType, ids := range ruleTypeIDs {
			deleteRules = append(deleteRules, wallarm.AccessRuleDeleteEntry{
				RuleType: ruleType,
				IDs:      ids,
			})
		}

		if err := client.IPListDelete(clientID, deleteRules); err != nil {
			return diag.FromErr(err)
		}

		return nil
	}
}

// buildRulesFromSchema constructs AccessRuleEntry slice from all configured rule type fields.
func buildRulesFromSchema(d *schema.ResourceData) ([]wallarm.AccessRuleEntry, diag.Diagnostics) {
	var rules []wallarm.AccessRuleEntry

	// subnet rules from ip_range
	if v, ok := d.GetOk("ip_range"); ok {
		ipRange := v.([]interface{})
		var ips []string
		for _, v := range ipRange {
			ip := v.(string)
			if strings.Contains(ip, "/") {
				subNetwork, err := strconv.Atoi(strings.Split(ip, "/")[1])
				if err != nil {
					return nil, diag.FromErr(fmt.Errorf("cannot parse subnet to integer. must be the number, got %v", err))
				}
				if subNetwork < 8 {
					return nil, diag.FromErr(fmt.Errorf("subnet must be >= /8, got %v", subNetwork))
				}
			}
			ips = append(ips, ip)
		}
		if len(ips) > 0 {
			rules = append(rules, wallarm.AccessRuleEntry{
				RulesType: "subnet",
				Values:    ips,
			})
		}
	}

	// location rules from country
	if v, ok := d.GetOk("country"); ok {
		countries := v.([]interface{})
		var vals []string
		for _, c := range countries {
			vals = append(vals, c.(string))
		}
		if len(vals) > 0 {
			rules = append(rules, wallarm.AccessRuleEntry{
				RulesType: "location",
				Values:    vals,
			})
		}
	}

	// datacenter rules from datacenter
	if v, ok := d.GetOk("datacenter"); ok {
		sources := v.([]interface{})
		var vals []string
		for _, s := range sources {
			vals = append(vals, s.(string))
		}
		if len(vals) > 0 {
			rules = append(rules, wallarm.AccessRuleEntry{
				RulesType: "datacenter",
				Values:    vals,
			})
		}
	}

	// proxy_type rules from proxy_type
	if v, ok := d.GetOk("proxy_type"); ok {
		proxies := v.([]interface{})
		var vals []string
		for _, p := range proxies {
			vals = append(vals, p.(string))
		}
		if len(vals) > 0 {
			rules = append(rules, wallarm.AccessRuleEntry{
				RulesType: "proxy_type",
				Values:    vals,
			})
		}
	}

	return rules, nil
}

// buildSubnetMatchers parses configured ip_range values into net.IPNet (for CIDRs)
// and a plain IP set (for single addresses), avoiding memory-expensive CIDR expansion.
func buildSubnetMatchers(d *schema.ResourceData) ([]*net.IPNet, map[string]bool) {
	var nets []*net.IPNet
	ips := make(map[string]bool)
	ipRange := d.Get("ip_range").([]interface{})
	for _, v := range ipRange {
		s := v.(string)
		if strings.Contains(s, "/") {
			_, ipNet, err := net.ParseCIDR(s)
			if err == nil {
				nets = append(nets, ipNet)
			}
		} else {
			ips[s] = true
		}
	}
	return nets, ips
}

// subnetMatch checks whether ipAddr is contained in any of the configured CIDRs or exact IPs.
func subnetMatch(ipAddr string, nets []*net.IPNet, ips map[string]bool) bool {
	if ips[ipAddr] {
		return true
	}
	parsed := net.ParseIP(ipAddr)
	if parsed == nil {
		return false
	}
	for _, n := range nets {
		if n.Contains(parsed) {
			return true
		}
	}
	return false
}

// buildStringLookup builds a set from a TypeList string field.
func buildStringLookup(d *schema.ResourceData, field string) map[string]bool {
	lookup := make(map[string]bool)
	if v, ok := d.GetOk(field); ok {
		items := v.([]interface{})
		for _, item := range items {
			lookup[item.(string)] = true
		}
	}
	return lookup
}

func parseExpireTime(d *schema.ResourceData) (int, diag.Diagnostics) {
	timeFormat := strings.ToLower(d.Get("time_format").(string))

	switch timeFormat {
	case "forever":
		return int(time.Now().AddDate(100, 0, 0).Unix()), nil
	case "minutes":
		expireTime, err := strconv.Atoi(d.Get("time").(string))
		if err != nil {
			return 0, diag.FromErr(fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Minutes`, got %v", err))
		}
		if expireTime == 0 {
			expireTime = 60045120
		}
		return int(time.Now().Add(time.Minute * time.Duration(expireTime)).Unix()), nil
	case "hours":
		expireTime, err := strconv.Atoi(d.Get("time").(string))
		if err != nil {
			return 0, diag.FromErr(fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Hours`, got %v", err))
		}
		if expireTime == 0 {
			expireTime = 60045120
		}
		return int(time.Now().Add(time.Hour * time.Duration(expireTime)).Unix()), nil
	case "days":
		expireTime, err := strconv.Atoi(d.Get("time").(string))
		if err != nil {
			return 0, diag.FromErr(fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Days`, got %v", err))
		}
		if expireTime == 0 {
			expireTime = 60045120
		}
		return int(time.Now().Add(24 * time.Hour * time.Duration(expireTime)).Unix()), nil
	case "weeks":
		expireTime, err := strconv.Atoi(d.Get("time").(string))
		if err != nil {
			return 0, diag.FromErr(fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Weeks`, got %v", err))
		}
		if expireTime == 0 {
			expireTime = 60045120
		}
		return int(time.Now().Add(7 * 24 * time.Hour * time.Duration(expireTime)).Unix()), nil
	case "months":
		expireTime, err := strconv.Atoi(d.Get("time").(string))
		if err != nil {
			return 0, diag.FromErr(fmt.Errorf("cannot parse time to integer. must be the number when `time_format` equals `Months`, got %v", err))
		}
		if expireTime == 0 {
			expireTime = 60045120
		}
		return int(time.Now().AddDate(0, expireTime, 0).Unix()), nil
	case "rfc3339":
		expireTime, err := time.Parse(time.RFC3339, d.Get("time").(string))
		if err != nil {
			return 0, diag.FromErr(fmt.Errorf("cannot parse time to integer. must be the valid RFC3339 time when `time_format` equals `RFC3339`, got %v.\nExample: 2006-01-02T15:04:05+07:00", err))
		}
		return int(expireTime.Unix()), nil
	}
	return 0, diag.FromErr(fmt.Errorf("unsupported time_format"))
}

// ipListFriendlyType maps API list type values to user-facing names for resource IDs.
func ipListFriendlyType(listType wallarm.IPListType) string {
	switch listType {
	case wallarm.DenylistType:
		return "deny"
	case wallarm.AllowlistType:
		return "allow"
	case wallarm.GraylistType:
		return "gray"
	default:
		return string(listType)
	}
}

// ipListRuleTypes returns a comma-separated string of rule types present in the rules slice.
func ipListRuleTypes(rules []wallarm.AccessRuleEntry) string {
	// Map API rule type names to user-facing names.
	friendly := map[string]string{
		"subnet":     "subnet",
		"location":   "country",
		"datacenter": "datacenter",
		"proxy_type": "proxy",
	}
	var types []string
	for _, r := range rules {
		if name, ok := friendly[r.RulesType]; ok {
			types = append(types, name)
		} else {
			types = append(types, r.RulesType)
		}
	}
	return strings.Join(types, ",")
}

// ipListValuesHash returns a short deterministic hash of the rule values for use in resource IDs.
func ipListValuesHash(rules []wallarm.AccessRuleEntry) string {
	var all []string
	for _, r := range rules {
		all = append(all, r.Values...)
	}
	sort.Strings(all)
	h := sha256.Sum256([]byte(strings.Join(all, ",")))
	return hex.EncodeToString(h[:4])
}
