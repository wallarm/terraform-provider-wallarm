package wallarm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ruleTypeSubnet = "subnet"

func resourceWallarmIPList(listType wallarm.IPListType) *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmIPListCreate(listType),
		ReadContext:   resourceWallarmIPListRead(listType),
		UpdateContext: resourceWallarmIPListUpdate(listType),
		DeleteContext: resourceWallarmIPListDelete(listType),
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmIPListImport(listType),
		},
		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
			"ip_range": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      IPListMaxSubnets,
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
				Computed: true,
			},
			"reason": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Terraform managed IP list",
			},
			"entry_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of entries tracked in address_id.",
			},
			"untracked_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of config values not found in the API.",
			},
			"untracked_ips": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Config values not found in the API — can be removed from config if the API rejected them.",
				Elem:        &schema.Schema{Type: schema.TypeString},
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
	return func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := apiClient(m)
		clientID, err := retrieveClientID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		// Serialize Creates for the same list type to prevent concurrent
		// cache refresh races between resources sharing the same denylist/allowlist/graylist.
		cache := m.(*ProviderMeta).IPListCache
		cache.LockCreate(listType)
		defer cache.UnlockCreate(listType)

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
			var apiErr *wallarm.APIError
			if stderrors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict {
				return diag.FromErr(fmt.Errorf("IP list rule conflicts with existing entries. "+
					"Resolve the conflicts and retry.\nAPI response: %s", apiErr.Body))
			}
			return diag.FromErr(err)
		}

		ruleType := ipListRuleTypes(rules)
		valuesHash := ipListValuesHash(rules)
		d.SetId(fmt.Sprintf("%d/%s/%s/%s", clientID, ipListFriendlyType(listType), ruleType, valuesHash))

		// Collect config values and rule types for cache lookup.
		var configValues []string
		var ruleTypes []string
		for _, r := range rules {
			configValues = append(configValues, r.Values...)
			ruleTypes = append(ruleTypes, r.RulesType)
		}

		// Refresh shared cache and resolve config values to group IDs.
		found, missing := cache.RefreshUntilFound(
			client, listType, clientID, configValues, ruleTypes,
			IPListCacheMaxRetries, IPListCacheRetryDelay*time.Second,
		)

		addrIDs := cacheEntriesToAddrIDs(found)
		if err := d.Set("address_id", addrIDs); err != nil {
			return diag.FromErr(err)
		}
		d.Set("entry_count", len(configValues)-len(missing))
		d.Set("untracked_count", len(missing))
		d.Set("untracked_ips", missing)
		d.Set("client_id", clientID)

		if len(missing) > 0 {
			log.Printf("[WARN] IP list Create: %d values not found in API after retries", len(missing))
		}

		return nil
	}
}

func resourceWallarmIPListRead(listType wallarm.IPListType) schema.ReadContextFunc {
	return func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := apiClient(m)
		clientID, err := retrieveClientID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		// Collect config values for cache lookup.
		configValues := ipListConfigValues(d)
		if len(configValues) == 0 {
			d.SetId("")
			return nil
		}

		// Ensure shared cache is loaded, then look up this resource's values.
		cache := m.(*ProviderMeta).IPListCache
		if err := cache.EnsureLoaded(client, listType, clientID); err != nil {
			return diag.FromErr(err)
		}

		found, missing := cache.LookupMany(listType, configValues)

		if len(found) == 0 {
			if !d.IsNewResource() {
				oldAddrs := d.Get("address_id").([]interface{})
				if len(oldAddrs) > 0 {
					log.Printf("[WARN] IP list %s was previously tracked but no longer found — removing from state", d.Id())
					d.SetId("")
				} else {
					log.Printf("[WARN] IP list %s not yet visible in API (address_id empty), keeping in state", d.Id())
				}
			}
			return nil
		}

		addrIDs := cacheEntriesToAddrIDs(found)
		if err := d.Set("address_id", addrIDs); err != nil {
			return diag.FromErr(fmt.Errorf("cannot set address_id: %v", err))
		}
		d.Set("entry_count", len(configValues)-len(missing))
		d.Set("untracked_count", len(missing))
		d.Set("untracked_ips", missing)
		d.Set("client_id", clientID)

		return nil
	}
}

func resourceWallarmIPListUpdate(listType wallarm.IPListType) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := apiClient(m)
		clientID, err := retrieveClientID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		cache := m.(*ProviderMeta).IPListCache

		// If only ip_range changed (subnet type), do a targeted diff update.
		if d.HasChange("ip_range") && !d.HasChanges("time_format", "time", "reason", "application") {
			return ipListSubnetDiffUpdate(ctx, d, m, client, clientID, listType, cache)
		}

		// For grouped types or when metadata changed, do full delete+create.
		// Use address_id from state to delete old entries.
		addrIDs := d.Get("address_id").([]interface{})
		if len(addrIDs) > 0 {
			if diags := deleteByAddrIDs(client, clientID, addrIDs); diags != nil {
				return diags
			}
		}
		cache.Invalidate(listType)
		return resourceWallarmIPListCreate(listType)(ctx, d, m)
	}
}

// ipListSubnetDiffUpdate deletes only removed IPs and creates only added IPs.
func ipListSubnetDiffUpdate(
	_ context.Context,
	d *schema.ResourceData,
	_ interface{},
	client wallarm.API,
	clientID int,
	listType wallarm.IPListType,
	cache *IPListCache,
) diag.Diagnostics {
	oldRaw, newRaw := d.GetChange("ip_range")
	oldIPs := toStringSet(oldRaw.([]interface{}))
	newIPs := toStringSet(newRaw.([]interface{}))

	var removed []string
	for ip := range oldIPs {
		if !newIPs[ip] {
			removed = append(removed, ip)
		}
	}
	var added []string
	for ip := range newIPs {
		if !oldIPs[ip] {
			added = append(added, ip)
		}
	}

	log.Printf("[DEBUG] IP list diff: %d removed, %d added, %d unchanged",
		len(removed), len(added), len(newIPs)-len(added))

	// Delete removed IPs using cache to resolve group IDs.
	if len(removed) > 0 {
		found, _ := cache.LookupMany(listType, removed)
		var deleteIDs []int
		for _, entry := range found {
			deleteIDs = append(deleteIDs, entry.GroupID)
		}

		// Fallback: search individually for any not in cache.
		if len(deleteIDs) < len(removed) {
			for _, ip := range removed {
				if _, ok := cache.Lookup(listType, ip); !ok {
					groups, err := client.IPListSearch(listType, clientID, ruleTypeSubnet, ip)
					if err != nil {
						return diag.FromErr(err)
					}
					for _, group := range groups {
						deleteIDs = append(deleteIDs, group.ID)
					}
				}
			}
		}

		log.Printf("[DEBUG] IP list diff: deleting %d group IDs for %d removed IPs", len(deleteIDs), len(removed))

		if len(deleteIDs) > 0 {
			if err := client.IPListDelete(clientID, []wallarm.AccessRuleDeleteEntry{
				{RuleType: ruleTypeSubnet, IDs: deleteIDs},
			}); err != nil {
				return diag.FromErr(fmt.Errorf("failed to delete removed IPs: %w", err))
			}
		}
	}

	// Create added IPs.
	if len(added) > 0 {
		unixTime, diags := parseExpireTime(d)
		if diags != nil {
			return diags
		}

		var apps []int
		if v, ok := d.GetOk("application"); ok {
			for _, a := range v.([]interface{}) {
				apps = append(apps, a.(int))
			}
		} else {
			apps = []int{0}
		}

		params := wallarm.AccessRuleCreateRequest{
			List:           listType,
			Force:          false,
			Reason:         d.Get("reason").(string),
			ApplicationIDs: apps,
			ExpiredAt:      unixTime,
			Rules: []wallarm.AccessRuleEntry{
				{RulesType: ruleTypeSubnet, Values: added},
			},
		}

		if err := client.IPListCreate(clientID, params); err != nil {
			return diag.FromErr(fmt.Errorf("failed to create added IPs: %w", err))
		}
	}

	// Refresh cache after modifications.
	if len(removed) > 0 || len(added) > 0 {
		cache.Invalidate(listType)
	}

	// Update resource ID hash since values changed.
	rules, _ := buildRulesFromSchema(d)
	valuesHash := ipListValuesHash(rules)
	ruleType := ipListRuleTypes(rules)
	d.SetId(fmt.Sprintf("%d/%s/%s/%s", clientID, ipListFriendlyType(listType), ruleType, valuesHash))

	// Don't update address_id — next terraform plan refresh handles it via Read + cache.
	return nil
}

// deleteByAddrIDs deletes IP list entries using group IDs from the address_id state.
func deleteByAddrIDs(client wallarm.API, clientID int, addrIDs []interface{}) diag.Diagnostics {
	ruleTypeIDs := make(map[string][]int)
	for _, entry := range addrIDs {
		e := entry.(map[string]interface{})
		ruleType := e["rule_type"].(string)
		id := e["ip_id"].(int)
		ruleTypeIDs[ruleType] = append(ruleTypeIDs[ruleType], id)
	}

	if len(ruleTypeIDs) == 0 {
		return nil
	}

	deleteRules := make([]wallarm.AccessRuleDeleteEntry, 0, len(ruleTypeIDs))
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

// ipListConfigValues extracts all config values (ip_range, country, datacenter, proxy_type) from schema.
func ipListConfigValues(d *schema.ResourceData) []string {
	var values []string
	for _, field := range []string{"ip_range", "country", "datacenter", "proxy_type"} {
		if v, ok := d.GetOk(field); ok {
			for _, item := range v.([]interface{}) {
				values = append(values, item.(string))
			}
		}
	}
	return values
}

// cacheEntriesToAddrIDs converts cache entries to the address_id schema format, sorted by group ID.
func cacheEntriesToAddrIDs(entries []IPCacheEntry) []interface{} {
	// Sort by GroupID for deterministic ordering.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GroupID < entries[j].GroupID
	})

	addrIDs := make([]interface{}, 0, len(entries))
	for _, entry := range entries {
		addrIDs = append(addrIDs, map[string]interface{}{
			"rule_type": entry.RuleType,
			"value":     entry.RawValue,
			"ip_id":     entry.GroupID,
		})
	}
	return addrIDs
}

func toStringSet(items []interface{}) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item.(string)] = true
	}
	return set
}

func resourceWallarmIPListDelete(listType wallarm.IPListType) schema.DeleteContextFunc {
	return func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := apiClient(m)
		clientID, err := retrieveClientID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}
		cache := m.(*ProviderMeta).IPListCache

		// Primary: use address_id from state (populated by Read during refresh).
		addrIDs := d.Get("address_id").([]interface{})
		if len(addrIDs) > 0 {
			log.Printf("[DEBUG] IPListDelete: using %d address_id entries from state", len(addrIDs))
			if diags := deleteByAddrIDs(client, clientID, addrIDs); diags != nil {
				return diags
			}
		}

		// Cleanup: refresh cache and look up any config values not in address_id.
		configValues := ipListConfigValues(d)
		if len(configValues) > 0 {
			if err := cache.Refresh(client, listType, clientID); err != nil {
				log.Printf("[WARN] IPListDelete: cache refresh for cleanup failed: %v", err)
			} else {
				found, _ := cache.LookupMany(listType, configValues)
				if len(found) > 0 {
					var cleanupIDs []int
					for _, entry := range found {
						cleanupIDs = append(cleanupIDs, entry.GroupID)
					}
					log.Printf("[DEBUG] IPListDelete: cleanup sweep found %d remaining entries", len(cleanupIDs))
					ruleTypeIDs := make(map[string][]int)
					for _, entry := range found {
						ruleTypeIDs[entry.RuleType] = append(ruleTypeIDs[entry.RuleType], entry.GroupID)
					}
					cleanupRules := make([]wallarm.AccessRuleDeleteEntry, 0, len(ruleTypeIDs))
					for ruleType, ids := range ruleTypeIDs {
						cleanupRules = append(cleanupRules, wallarm.AccessRuleDeleteEntry{
							RuleType: ruleType,
							IDs:      ids,
						})
					}
					if err := client.IPListDelete(clientID, cleanupRules); err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}

		cache.Invalidate(listType)
		return nil
	}
}

// resourceWallarmIPListImport handles terraform import for IP list resources.
//
// Two import ID formats:
//
//	{clientID}/{groupID}           — import a single grouped entry (country/datacenter/proxy/single subnet)
//	{clientID}/subnet/{expiredAt}  — import all subnet entries with this expiration as one resource
//
// Examples:
//
//	terraform import wallarm_denylist.countries 8649/52000393
//	terraform import wallarm_denylist.ips 8649/subnet/1804809600
func resourceWallarmIPListImport(listType wallarm.IPListType) schema.StateContextFunc {
	return func(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
		client := apiClient(m)

		parts := strings.SplitN(d.Id(), "/", 3)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid import ID %q, expected {clientID}/{groupID} or {clientID}/subnet/{expiredAt}", d.Id())
		}

		clientID, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid client_id %q: %w", parts[0], err)
		}
		d.Set("client_id", clientID)

		// Fetch all groups for this list type.
		allGroups, err := client.IPListRead(listType, clientID, IPListPageSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read IP lists: %w", err)
		}

		// Mode 1: {clientID}/subnet/{expiredAt} — merge all subnets by expiry.
		if len(parts) == 3 && parts[1] == ruleTypeSubnet {
			expiredAt, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, fmt.Errorf("invalid expired_at %q: %w", parts[2], err)
			}

			var ips []string
			var addrIDs []interface{}
			var reason string
			var apps []int
			for _, g := range allGroups {
				if g.RuleType != ruleTypeSubnet || g.ExpiredAt != expiredAt {
					continue
				}
				for _, v := range g.Values {
					// Strip /32 from bare IPs for config compatibility.
					ips = append(ips, strings.TrimSuffix(v, "/32"))
				}
				addrIDs = append(addrIDs, map[string]interface{}{
					"rule_type": g.RuleType,
					"value":     strings.Join(g.Values, ","),
					"ip_id":     g.ID,
				})
				if reason == "" {
					reason = g.Reason
				}
				if len(apps) == 0 {
					apps = g.ApplicationIDs
				}
			}

			if len(ips) == 0 {
				return nil, fmt.Errorf("no subnet entries found with expired_at=%d", expiredAt)
			}

			d.Set("ip_range", ips)
			d.Set("reason", reason)
			d.Set("time_format", "RFC3339")
			d.Set("time", time.Unix(int64(expiredAt), 0).UTC().Format(time.RFC3339))
			if len(apps) > 0 {
				d.Set("application", apps)
			}
			d.Set("address_id", addrIDs)
			d.Set("entry_count", len(addrIDs))

			rules := []wallarm.AccessRuleEntry{{RulesType: ruleTypeSubnet, Values: ips}}
			valuesHash := ipListValuesHash(rules)
			d.SetId(fmt.Sprintf("%d/%s/subnet/%s", clientID, ipListFriendlyType(listType), valuesHash))

			return []*schema.ResourceData{d}, nil
		}

		// Mode 2: {clientID}/{groupID} — import a single group.
		groupID, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid group_id %q: %w", parts[1], err)
		}

		var found *wallarm.IPRule
		for i := range allGroups {
			if allGroups[i].ID == groupID {
				found = &allGroups[i]
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("IP list group %d not found", groupID)
		}

		addrIDs := []interface{}{
			map[string]interface{}{
				"rule_type": found.RuleType,
				"value":     strings.Join(found.Values, ","),
				"ip_id":     found.ID,
			},
		}
		d.Set("address_id", addrIDs)
		d.Set("entry_count", len(addrIDs))
		d.Set("reason", found.Reason)
		d.Set("time_format", "RFC3339")
		d.Set("time", time.Unix(int64(found.ExpiredAt), 0).UTC().Format(time.RFC3339))
		if len(found.ApplicationIDs) > 0 {
			d.Set("application", found.ApplicationIDs)
		}

		switch found.RuleType {
		case ruleTypeSubnet:
			ips := make([]string, len(found.Values))
			for i, v := range found.Values {
				ips[i] = strings.TrimSuffix(v, "/32")
			}
			d.Set("ip_range", ips)
		case "location":
			d.Set("country", found.Values)
		case "datacenter":
			d.Set("datacenter", found.Values)
		case "proxy_type":
			d.Set("proxy_type", found.Values)
		}

		rules, _ := buildRulesFromSchema(d)
		valuesHash := ipListValuesHash(rules)
		ruleType := ipListRuleTypes(rules)
		d.SetId(fmt.Sprintf("%d/%s/%s/%s", clientID, ipListFriendlyType(listType), ruleType, valuesHash))

		return []*schema.ResourceData{d}, nil
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
				RulesType: ruleTypeSubnet,
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
		ruleTypeSubnet: ruleTypeSubnet,
		"location":     "country",
		"datacenter":   "datacenter",
		"proxy_type":   "proxy",
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
