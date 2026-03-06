package wallarm

import (
	"fmt"
	"log"
	"sync"
	"time"

	wallarm "github.com/wallarm/wallarm-go"
)

const (
	// defaultBulkFetchLimit is the number of hints fetched per API call during bulk loading.
	defaultBulkFetchLimit = 100
	// maxBulkFetchPages caps the number of paginated requests to prevent runaway fetches.
	maxBulkFetchPages = 200
)

// CacheStats provides a snapshot of the cache's operational state.
type CacheStats struct {
	Loaded        bool          `json:"loaded"`
	HintCount     int           `json:"hint_count"`
	CacheHits     int64         `json:"cache_hits"`
	CacheMisses   int64         `json:"cache_misses"`
	Passthroughs  int64         `json:"passthroughs"`
	BulkLoads     int64         `json:"bulk_loads"`
	Invalidations int64         `json:"invalidations"`
	APICallsSaved int64         `json:"api_calls_saved"`
	LastLoadTime  time.Duration `json:"last_load_time"`
	LastLoadAt    time.Time     `json:"last_load_at"`
}

// HintCache provides a thread-safe, bulk-loaded cache of hints keyed by hint ID.
// It is designed to reduce API calls during terraform plan/refresh, where Terraform
// calls ReadContext on every rule resource individually.
//
// The cache is populated lazily on first access for a given clientID, fetching all
// hints in paginated batches. Subsequent reads are served from memory.
// Any mutating operation (create/update/delete) invalidates the cache so the next
// read cycle re-fetches fresh data.
type HintCache struct {
	mu     sync.Mutex
	hints  map[int]*wallarm.ActionBody // keyed by hint ID
	loaded bool

	// stats counters (updated under mu)
	cacheHits     int64
	cacheMisses   int64
	passthroughs  int64
	bulkLoads     int64
	invalidations int64
	lastLoadTime  time.Duration
	lastLoadAt    time.Time
}

// NewHintCache creates an empty, unloaded HintCache.
func NewHintCache() *HintCache {
	return &HintCache{
		hints: make(map[int]*wallarm.ActionBody),
	}
}

// ensureLoaded populates the cache if it hasn't been loaded yet.
// Only the first goroutine to arrive performs the bulk fetch; others block
// briefly on the mutex and then find loaded=true.
func (c *HintCache) ensureLoaded(clientID int, api wallarm.API) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded {
		return nil
	}

	if err := c.bulkLoad(clientID, api); err != nil {
		return err
	}
	c.loaded = true
	return nil
}

// bulkLoad fetches all hints for a client in paginated batches.
// Must be called while holding c.mu.
func (c *HintCache) bulkLoad(clientID int, api wallarm.API) error {
	c.hints = make(map[int]*wallarm.ActionBody)
	offset := 0
	totalFetched := 0
	startTime := time.Now()

	log.Printf("[INFO] HintCache: starting bulk load for client %d (batch size: %d, system=false)", clientID, defaultBulkFetchLimit)
	systemFalse := false

	for page := 0; page < maxBulkFetchPages; page++ {
		resp, err := api.HintRead(&wallarm.HintRead{
			Limit:     defaultBulkFetchLimit,
			Offset:    offset,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				System:   &systemFalse,
			},
		})
		if err != nil {
			return fmt.Errorf("HintCache bulk load failed at offset %d: %w", offset, err)
		}

		if resp.Body == nil || len(*resp.Body) == 0 {
			break
		}

		batch := *resp.Body
		for i := range batch {
			c.hints[batch[i].ID] = &batch[i]
		}
		totalFetched += len(batch)

		if len(batch) < defaultBulkFetchLimit {
			break // last page
		}
		offset += defaultBulkFetchLimit
	}

	loadDuration := time.Since(startTime)
	apiCalls := (offset / defaultBulkFetchLimit) + 1

	c.bulkLoads++
	c.lastLoadTime = loadDuration
	c.lastLoadAt = time.Now()

	log.Printf("[INFO] HintCache: loaded %d hints for client %d in %s (%d API calls) — subsequent single-hint reads will be served from cache",
		totalFetched, clientID, loadDuration.Round(time.Millisecond), apiCalls)

	return nil
}

// Get returns a cached hint by ID. Returns nil, false if not found.
// Tracks hit/miss stats and logs at DEBUG level.
func (c *HintCache) Get(hintID int) (*wallarm.ActionBody, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	h, ok := c.hints[hintID]
	if ok {
		c.cacheHits++
		log.Printf("[DEBUG] HintCache: HIT hint_id=%d (hits=%d misses=%d saved=%d)", hintID, c.cacheHits, c.cacheMisses, c.cacheHits)
	} else {
		c.cacheMisses++
		log.Printf("[DEBUG] HintCache: MISS hint_id=%d — falling back to API (hits=%d misses=%d)", hintID, c.cacheHits, c.cacheMisses)
	}
	return h, ok
}

// Invalidate clears the cache so the next read triggers a fresh bulk load.
// caller identifies which method triggered the invalidation for debugging.
func (c *HintCache) Invalidate(caller string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	prevCount := len(c.hints)
	c.hints = make(map[int]*wallarm.ActionBody)
	c.loaded = false
	c.invalidations++
	log.Printf("[INFO] HintCache: INVALIDATED by %s — cleared %d cached hints (invalidation #%d, total hits before clear: %d)", caller, prevCount, c.invalidations, c.cacheHits)
}

// trackPassthrough increments the passthrough counter for non-cacheable queries.
func (c *HintCache) trackPassthrough() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.passthroughs++
}

// All returns a copy of all cached hints. Returns nil if the cache is not loaded.
func (c *HintCache) All() []wallarm.ActionBody {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loaded {
		return nil
	}
	result := make([]wallarm.ActionBody, 0, len(c.hints))
	for _, h := range c.hints {
		result = append(result, *h)
	}
	return result
}

// Stats returns a snapshot of the cache's operational statistics.
func (c *HintCache) Stats() CacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return CacheStats{
		Loaded:        c.loaded,
		HintCount:     len(c.hints),
		CacheHits:     c.cacheHits,
		CacheMisses:   c.cacheMisses,
		Passthroughs:  c.passthroughs,
		BulkLoads:     c.bulkLoads,
		Invalidations: c.invalidations,
		APICallsSaved: c.cacheHits,
		LastLoadTime:  c.lastLoadTime,
		LastLoadAt:    c.lastLoadAt,
	}
}

// CachedClient wraps a wallarm.API and overrides HintRead with a cache-backed
// implementation. All other API methods delegate to the embedded client.
//
// HintRead is overridden so that single-ID lookups (the pattern used by every
// rule resource's Read function) are served from a bulk-loaded cache.
// Multi-ID or complex filter queries pass through to the underlying API.
//
// Mutating methods (HintCreate, HintDelete, HintUpdateV3, RuleDelete) invalidate
// the cache AFTER the mutation succeeds, so:
//   - Failed mutations don't unnecessarily clear the cache
//   - Post-mutation reads (e.g. Create calling Read at the end) get fresh data
type CachedClient struct {
	wallarm.API
	hintCache *HintCache
}

// NewCachedClient wraps an existing wallarm.API with hint caching.
func NewCachedClient(api wallarm.API) *CachedClient {
	return &CachedClient{
		API:       api,
		hintCache: NewHintCache(),
	}
}

// HintCacheStats returns the current cache statistics.
func (c *CachedClient) HintCacheStats() CacheStats {
	return c.hintCache.Stats()
}

// LogHintCacheStats logs a summary of cache performance.
func (c *CachedClient) LogHintCacheStats() {
	s := c.hintCache.Stats()
	total := s.CacheHits + s.CacheMisses + s.Passthroughs
	hitRate := float64(0)
	if s.CacheHits+s.CacheMisses > 0 {
		hitRate = float64(s.CacheHits) / float64(s.CacheHits+s.CacheMisses) * 100
	}
	log.Printf("[INFO] HintCache stats: %d total reads | %d hits (%.1f%%) | %d misses | %d passthroughs | %d bulk loads | %d invalidations | %d API calls saved | %d hints cached",
		total, s.CacheHits, hitRate, s.CacheMisses, s.Passthroughs, s.BulkLoads, s.Invalidations, s.APICallsSaved, s.HintCount)
}

// HintRead overrides the embedded API's HintRead. For single-ID lookups
// (the common Read pattern), it serves from the bulk cache. For all other
// query patterns, it passes through to the underlying API.
func (c *CachedClient) HintRead(body *wallarm.HintRead) (*wallarm.HintReadResp, error) {
	// Only use cache for single-ID, single-client lookups — the pattern used
	// by ResourceRuleWallarmRead: Filter{Clientid: []int{cid}, ID: []int{hid}}
	if body.Filter != nil &&
		len(body.Filter.ID) == 1 &&
		len(body.Filter.Clientid) == 1 &&
		len(body.Filter.ActionID) == 0 &&
		len(body.Filter.Type) == 0 {

		clientID := body.Filter.Clientid[0]
		hintID := body.Filter.ID[0]

		if err := c.hintCache.ensureLoaded(clientID, c.API); err != nil {
			log.Printf("[WARN] HintCache: bulk load failed, falling back to direct API call: %s", err)
			return c.API.HintRead(body)
		}

		hint, ok := c.hintCache.Get(hintID)
		if !ok {
			// Not in cache — could be newly created after cache load.
			return c.API.HintRead(body)
		}

		return &wallarm.HintReadResp{
			Status: 200,
			Body:   &[]wallarm.ActionBody{*hint},
		}, nil
	}

	// Non-cacheable query pattern — pass through to real API
	c.hintCache.trackPassthrough()
	log.Printf("[DEBUG] HintCache: PASSTHROUGH — query has multi-ID/ActionID/Type filter, bypassing cache")
	return c.API.HintRead(body)
}

// HintCreate delegates to the underlying API and invalidates the cache AFTER success.
func (c *CachedClient) HintCreate(body *wallarm.ActionCreate) (*wallarm.ActionCreateResp, error) {
	resp, err := c.API.HintCreate(body)
	if err != nil {
		return resp, err
	}
	c.hintCache.Invalidate("HintCreate")
	return resp, nil
}

// HintDelete delegates to the underlying API and invalidates the cache AFTER success.
func (c *CachedClient) HintDelete(body *wallarm.HintDelete) error {
	err := c.API.HintDelete(body)
	if err != nil {
		return err
	}
	c.hintCache.Invalidate("HintDelete")
	return nil
}

// RuleDelete delegates to the underlying API and invalidates the cache AFTER success.
func (c *CachedClient) RuleDelete(actionID int) error {
	err := c.API.RuleDelete(actionID)
	if err != nil {
		return err
	}
	c.hintCache.Invalidate("RuleDelete")
	return nil
}

// HintUpdateV3 delegates to the underlying API and invalidates the cache AFTER success.
func (c *CachedClient) HintUpdateV3(ruleID int, body *wallarm.HintUpdateV3Params) (*wallarm.ActionCreateResp, error) {
	resp, err := c.API.HintUpdateV3(ruleID, body)
	if err != nil {
		return resp, err
	}
	c.hintCache.Invalidate("HintUpdateV3")
	return resp, nil
}
