package wallarm

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	wallarm "github.com/wallarm/wallarm-go"
)

// allRuleTypes is the full list of IP list rule types.
var allRuleTypes = []string{ruleTypeSubnet, "location", "datacenter", "proxy_type"}

// IPCacheEntry stores the API group ID and metadata for a single IP list value.
type IPCacheEntry struct {
	GroupID  int
	RuleType string
	RawValue string // API value (e.g. "1.2.3.4/32")
}

// IPListCache provides a shared, thread-safe map of IP list values to their API group IDs.
// It is populated by per-rule-type bulk fetches from the Wallarm API and used by all
// IP list resources to resolve their config values to group IDs.
//
// Create operations are serialized per list type via createMu to prevent race conditions
// when multiple resources of the same list type are created concurrently.
type IPListCache struct {
	mu       sync.Mutex
	entries  map[wallarm.IPListType]map[string]IPCacheEntry
	loaded   map[wallarm.IPListType]bool
	createMu map[wallarm.IPListType]*sync.Mutex
}

// NewIPListCache creates an empty cache.
func NewIPListCache() *IPListCache {
	return &IPListCache{
		entries: make(map[wallarm.IPListType]map[string]IPCacheEntry),
		loaded:  make(map[wallarm.IPListType]bool),
		createMu: map[wallarm.IPListType]*sync.Mutex{
			wallarm.DenylistType:  {},
			wallarm.AllowlistType: {},
			wallarm.GraylistType:  {},
		},
	}
}

// LockCreate acquires the per-list-type Create mutex.
func (c *IPListCache) LockCreate(listType wallarm.IPListType) {
	c.createMu[listType].Lock()
}

// UnlockCreate releases the per-list-type Create mutex.
func (c *IPListCache) UnlockCreate(listType wallarm.IPListType) {
	c.createMu[listType].Unlock()
}

// EnsureLoaded populates the cache for the given list type (all rule types) if not already loaded.
func (c *IPListCache) EnsureLoaded(client wallarm.API, listType wallarm.IPListType, clientID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded[listType] {
		return nil
	}
	return c.loadRuleTypesLocked(client, listType, clientID, allRuleTypes)
}

// Refresh forces a re-fetch of all rule types for the given list type.
func (c *IPListCache) Refresh(client wallarm.API, listType wallarm.IPListType, clientID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.loadRuleTypesLocked(client, listType, clientID, allRuleTypes)
}

// RefreshRuleTypes forces a re-fetch of specific rule types for the given list type.
// Only the specified rule types are re-fetched; other cached data is preserved.
func (c *IPListCache) RefreshRuleTypes(client wallarm.API, listType wallarm.IPListType, clientID int, ruleTypes []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.loadRuleTypesLocked(client, listType, clientID, ruleTypes)
}

// RefreshUntilFound fetches specific rule types and resolves config values.
//
// Retry logic:
//   - Retry if NONE of this resource's values are found (API hasn't indexed them yet)
//   - Once at least some values appear, stop retrying and use IPListSearch for stragglers
//   - 5xx errors are handled at the wallarm-go transport level (auto-retry)
func (c *IPListCache) RefreshUntilFound(
	client wallarm.API,
	listType wallarm.IPListType,
	clientID int,
	values []string,
	ruleTypes []string,
	maxRetries int,
	retryDelay time.Duration,
) (found []IPCacheEntry, missing []string) {
	// Step 1: fetch specific rule types with retries until at least some of our values appear.
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[DEBUG] IPListCache: retry %d/%d — none of %d values found yet in list=%s types=%v, waiting %s",
				attempt, maxRetries, len(values), string(listType), ruleTypes, retryDelay)
			time.Sleep(retryDelay)
		}

		if err := c.RefreshRuleTypes(client, listType, clientID, ruleTypes); err != nil {
			log.Printf("[WARN] IPListCache: refresh failed: %v", err)
			continue
		}

		found, missing = c.LookupMany(listType, values)
		if len(found) > 0 {
			break
		}
	}

	if len(missing) == 0 {
		return found, nil
	}

	// Step 2: targeted search for missing values (pagination boundary or propagation delay).
	log.Printf("[DEBUG] IPListCache: %d/%d values missing after bulk fetch, searching individually (list=%s)",
		len(missing), len(values), string(listType))

	c.mu.Lock()
	m := c.entries[listType]
	var stillMissing []string
	for _, val := range missing {
		// Determine rule type for search — use the first one from ruleTypes.
		rt := ruleTypes[0]
		groups, err := client.IPListSearch(listType, clientID, rt, val)
		if err != nil {
			log.Printf("[WARN] IPListCache: search failed for %s: %v", val, err)
			stillMissing = append(stillMissing, val)
			continue
		}
		if len(groups) == 0 {
			stillMissing = append(stillMissing, val)
			continue
		}
		for _, group := range groups {
			entry := IPCacheEntry{
				GroupID:  group.ID,
				RuleType: group.RuleType,
				RawValue: group.Values[0],
			}
			if m != nil {
				m[val] = entry
				for _, v := range group.Values {
					m[v] = entry
					bareIP, _, _ := strings.Cut(v, "/")
					if bareIP != v {
						m[bareIP] = entry
					}
				}
			}
			found = append(found, entry)
		}
	}
	c.mu.Unlock()
	missing = stillMissing

	if len(missing) > 0 {
		log.Printf("[WARN] IPListCache: %d values not found in API (list=%s, missing=%v)",
			len(missing), string(listType), missing)
	}
	return found, missing
}

// Lookup returns the cache entry for a single value, or false if not found.
func (c *IPListCache) Lookup(listType wallarm.IPListType, value string) (IPCacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.entries[listType]
	if !ok {
		return IPCacheEntry{}, false
	}
	entry, found := m[value]
	return entry, found
}

// LookupMany resolves a batch of values against the cache.
// Returns found entries and a list of values not in the cache.
func (c *IPListCache) LookupMany(listType wallarm.IPListType, values []string) (found []IPCacheEntry, missing []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := c.entries[listType]
	if m == nil {
		return nil, values
	}

	seen := make(map[int]bool)
	for _, val := range values {
		if entry, ok := m[val]; ok {
			if !seen[entry.GroupID] {
				seen[entry.GroupID] = true
				found = append(found, entry)
			}
		} else {
			missing = append(missing, val)
		}
	}
	return found, missing
}

// Invalidate clears the cache for a list type so the next access triggers a fresh fetch.
func (c *IPListCache) Invalidate(listType wallarm.IPListType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, listType)
	c.loaded[listType] = false
	log.Printf("[DEBUG] IPListCache: invalidated cache for list type %s", string(listType))
}

// loadRuleTypesLocked fetches specific rule types from the API and merges into the cache map.
// Must be called while holding c.mu.
func (c *IPListCache) loadRuleTypesLocked(client wallarm.API, listType wallarm.IPListType, clientID int, ruleTypes []string) error {
	startTime := time.Now()

	groups, err := client.IPListReadByRuleType(listType, clientID, ruleTypes, IPListPageSize)
	if err != nil {
		return err
	}

	// Ensure the map exists for this list type.
	if c.entries[listType] == nil {
		c.entries[listType] = make(map[string]IPCacheEntry, len(groups)*2)
	}
	m := c.entries[listType]

	// Clear existing entries for the fetched rule types before re-populating.
	for key, entry := range m {
		for _, rt := range ruleTypes {
			if entry.RuleType == rt {
				delete(m, key)
				break
			}
		}
	}

	// Build per-rule-type counters for logging.
	typeCounts := make(map[string]int)
	for _, group := range groups {
		entry := IPCacheEntry{
			GroupID:  group.ID,
			RuleType: group.RuleType,
		}
		typeCounts[group.RuleType]++
		for _, val := range group.Values {
			entry.RawValue = val
			m[val] = entry
			if group.RuleType == ruleTypeSubnet {
				bareIP, _, _ := strings.Cut(val, "/")
				if bareIP != val {
					m[bareIP] = entry
				}
			}
		}
	}

	c.loaded[listType] = true

	// Format per-rule-type breakdown.
	breakdown := make([]string, 0, len(ruleTypes))
	for _, rt := range ruleTypes {
		breakdown = append(breakdown, fmt.Sprintf("%s=%d", rt, typeCounts[rt]))
	}

	log.Printf("[INFO] IPListCache: loaded %d groups (%d map entries) for list=%s [%s] in %s",
		len(groups), len(m), string(listType), strings.Join(breakdown, ", "), time.Since(startTime).Round(time.Millisecond))

	// Dump full map at DEBUG level (visible with TF_LOG=DEBUG).
	for key, entry := range m {
		log.Printf("[DEBUG] IPListCache:   %q → group_id=%d rule_type=%s raw=%s",
			key, entry.GroupID, entry.RuleType, entry.RawValue)
	}

	return nil
}

// EntryCount returns the number of map entries for a list type.
func (c *IPListCache) EntryCount(listType wallarm.IPListType) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries[listType])
}

// DumpEntries logs every cache entry at DEBUG level for the given list type.
// Call with TF_LOG=DEBUG to see output.
func (c *IPListCache) DumpEntries(listType wallarm.IPListType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := c.entries[listType]
	if m == nil {
		log.Printf("[DEBUG] IPListCache dump: list=%s — empty (not loaded)", string(listType))
		return
	}

	log.Printf("[DEBUG] IPListCache dump: list=%s — %d map entries", string(listType), len(m))
	for key, entry := range m {
		log.Printf("[DEBUG]   %q → group_id=%d rule_type=%s raw=%s",
			key, entry.GroupID, entry.RuleType, entry.RawValue)
	}
}
