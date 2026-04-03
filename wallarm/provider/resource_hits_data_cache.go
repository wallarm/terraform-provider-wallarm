package wallarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmHitsDataCache() *schema.Resource {
	return &schema.Resource{
		Description: "Deduplicated cache for hits-derived rule data. " +
			"Stores aggregated hit data keyed by action_hash. " +
			"Multiple request_ids sharing the same action produce one cache entry.",

		CreateContext: resourceHitsDataCacheCreate,
		ReadContext:   resourceHitsDataCacheRead,
		UpdateContext: resourceHitsDataCacheUpdate,
		DeleteContext: resourceHitsDataCacheDelete,

		CustomizeDiff: customdiff.All(hitsDataCacheCustomizeDiff),

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"new_entries": {
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map of request_id -> aggregated JSON for newly fetched hits. Cleared after processing.",
			},

			"request_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All active request_ids. Used for cleanup — cache entries with no references are removed.",
			},

			// Computed outputs.
			"cache": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON map of action_hash -> aggregated JSON. Deduplicated — one entry per unique action scope.",
			},
			"request_to_action": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON map of request_id -> action_hash. Cross-reference for traceability and cleanup.",
			},
		},
	}
}

// ─── CustomizeDiff ──────────────────────────────────────────────────────────

// hitsDataCacheCustomizeDiff computes cache and request_to_action at plan time
// so that downstream for_each keys are known before apply.
func hitsDataCacheCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	// Read existing state.
	cache := readJSONMapFromDiff(d, "cache")
	reqToAction := readJSONMapFromDiff(d, "request_to_action")

	// Process new entries from config.
	if v, ok := d.GetOk("new_entries"); ok {
		entries := v.(map[string]interface{})
		for reqID, val := range entries {
			aggJSON, ok := val.(string)
			if !ok || aggJSON == "" {
				continue
			}
			if err := mergeNewEntry(cache, reqToAction, reqID, aggJSON); err != nil {
				return fmt.Errorf("failed to process entry for %s: %w", reqID, err)
			}
		}
	}

	// Build active request_ids set.
	requestIDsSet := d.Get("request_ids").(*schema.Set)
	activeIDs := make(map[string]bool, requestIDsSet.Len())
	for _, v := range requestIDsSet.List() {
		activeIDs[v.(string)] = true
	}

	// Cleanup removed request_ids and orphaned cache entries.
	cleanupCache(cache, reqToAction, activeIDs)

	// Serialize and set as known values for the plan.
	cacheJSON, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}
	reqToActionJSON, err := json.Marshal(reqToAction)
	if err != nil {
		return fmt.Errorf("failed to marshal request_to_action: %w", err)
	}

	if err := d.SetNew("cache", string(cacheJSON)); err != nil {
		return err
	}
	if err := d.SetNew("request_to_action", string(reqToActionJSON)); err != nil {
		return err
	}

	return nil
}

// ─── CRUD ───────────────────────────────────────────────────────────────────

func resourceHitsDataCacheCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("hits_data_cache_%d", clientID))
	if err := d.Set("client_id", clientID); err != nil {
		return diag.FromErr(err)
	}

	// CustomizeDiff already computed cache and request_to_action via SetNew.
	// The SDK applies those planned values to state automatically.
	// No reprocessing needed — just log.
	log.Printf("[INFO] wallarm_hits_data_cache: created")
	return nil
}

func resourceHitsDataCacheRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Clear new_entries so state has empty map. On next plan:
	// empty config matches empty state → no diff. New data triggers update.
	d.Set("new_entries", map[string]interface{}{})
	return nil
}

func resourceHitsDataCacheUpdate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// CustomizeDiff already computed cache and request_to_action via SetNew.
	// The SDK applies those planned values to state automatically.
	log.Printf("[INFO] wallarm_hits_data_cache: updated")
	return nil
}

func resourceHitsDataCacheDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

// ─── Core logic ─────────────────────────────────────────────────────────────

// mergeNewEntry parses aggregated JSON, extracts action_hash, and merges groups into the cache.
func mergeNewEntry(cache map[string]string, reqToAction map[string]string, reqID, aggJSON string) error {
	var parsed aggregatedOutput
	if err := json.Unmarshal([]byte(aggJSON), &parsed); err != nil {
		return fmt.Errorf("failed to parse aggregated JSON: %w", err)
	}

	actionHash := parsed.ActionHash
	if actionHash == "" {
		return fmt.Errorf("empty action_hash in aggregated JSON")
	}

	reqToAction[reqID] = actionHash

	existing, found := cache[actionHash]
	if !found {
		cache[actionHash] = aggJSON
		return nil
	}

	var existingParsed aggregatedOutput
	if err := json.Unmarshal([]byte(existing), &existingParsed); err != nil {
		cache[actionHash] = aggJSON
		return nil
	}

	merged := mergeGroups(existingParsed.Groups, parsed.Groups)
	existingParsed.Groups = merged

	mergedJSON, err := json.Marshal(existingParsed)
	if err != nil {
		return fmt.Errorf("failed to marshal merged cache: %w", err)
	}
	cache[actionHash] = string(mergedJSON)
	return nil
}

// mergeGroups merges two group slices by key, unioning stamps within matching groups.
func mergeGroups(existing, incoming []aggregatedGroup) []aggregatedGroup {
	groupMap := make(map[string]*aggregatedGroup, len(existing))
	for i := range existing {
		g := existing[i]
		groupMap[g.Key] = &g
	}

	for _, g := range incoming {
		if eg, found := groupMap[g.Key]; found {
			stampSet := make(map[int]bool, len(eg.Stamps))
			for _, s := range eg.Stamps {
				stampSet[s] = true
			}
			for _, s := range g.Stamps {
				if !stampSet[s] {
					eg.Stamps = append(eg.Stamps, s)
				}
			}
			sort.Ints(eg.Stamps)
		} else {
			newGroup := g
			groupMap[g.Key] = &newGroup
		}
	}

	keys := make([]string, 0, len(groupMap))
	for k := range groupMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]aggregatedGroup, 0, len(keys))
	for _, k := range keys {
		result = append(result, *groupMap[k])
	}
	return result
}

// cleanupCache removes request_ids not in activeIDs, then removes cache entries with no references.
func cleanupCache(cache map[string]string, reqToAction map[string]string, activeIDs map[string]bool) {
	for reqID := range reqToAction {
		if !activeIDs[reqID] {
			delete(reqToAction, reqID)
		}
	}

	referenced := make(map[string]bool, len(cache))
	for _, ah := range reqToAction {
		referenced[ah] = true
	}

	for ah := range cache {
		if !referenced[ah] {
			log.Printf("[INFO] wallarm_hits_data_cache: removing orphaned cache entry %s", ah)
			delete(cache, ah)
		}
	}
}

// buildCachedRequestIDs returns a sorted comma-separated string of request_ids
// that have data in the cache (exist in reqToAction).
func buildCachedRequestIDs(reqToAction map[string]string) string {
	ids := make([]string, 0, len(reqToAction))
	for id := range reqToAction {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

// readJSONMapFromDiff reads a JSON-encoded map[string]string from a ResourceDiff.
// Used in CustomizeDiff to read existing state values.
func readJSONMapFromDiff(d *schema.ResourceDiff, key string) map[string]string {
	raw := d.Get(key)
	if raw == nil {
		return make(map[string]string)
	}
	s, ok := raw.(string)
	if !ok || s == "" || s == "{}" {
		return make(map[string]string)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		log.Printf("[WARN] wallarm_hits_data_cache: failed to parse %s from diff: %v", key, err)
		return make(map[string]string)
	}
	return m
}
