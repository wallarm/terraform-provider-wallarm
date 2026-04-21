package wallarm

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	wallarm "github.com/wallarm/wallarm-go"
)

// CacheStats provides a snapshot of the cache's operational state.
type CacheStats struct {
	FullyLoaded   bool      `json:"fully_loaded"`
	HintCount     int       `json:"hint_count"`
	CacheHits     int64     `json:"cache_hits"`
	PageFetches   int64     `json:"page_fetches"`
	Passthroughs  int64     `json:"passthroughs"`
	Invalidations int64     `json:"invalidations"`
	LastFetchAt   time.Time `json:"last_fetch_at"`
}

// HintCache provides a thread-safe, lazily-paginated cache of hints keyed by hint ID.
//
// Instead of bulk-loading all hints upfront, it fetches pages on demand:
// - First Read for ID 123 → fetch page 1 (200 hints) → cache all → check for 123
// - Next Read for ID 456 → check cache → hit (was on page 1) → return
// - Next Read for ID 999 → check cache → miss → fetch page 2 → check → found
//
// This minimizes API calls: if 5 managed rules are all on page 1, only 1 API call.
type HintCache struct {
	mu          sync.Mutex
	hints       map[int]*wallarm.ActionBody
	nextOffset  int  // next page offset to fetch
	fullyLoaded bool // all pages have been fetched
	clientID    int  // set on first fetch, used for consistency

	// stats
	cacheHits     int64
	pageFetches   int64
	passthroughs  int64
	invalidations int64
	lastFetchAt   time.Time
}

// isCredentialStuffingType returns true for rule types that are served by the
// dedicated credential stuffing API (v4). With elevated permissions HintRead
// also returns these, so the cache skips them to avoid duplicates — the
// CredentialStuffingCache is the authoritative source.
func isCredentialStuffingType(t string) bool {
	return t == "credentials_point" || t == "credentials_regex"
}

// NewHintCache creates an empty HintCache.
func NewHintCache() *HintCache {
	return &HintCache{
		hints: make(map[int]*wallarm.ActionBody),
	}
}

// GetOrFetch returns a cached hint by ID. If not cached, fetches pages lazily
// until the hint is found or all pages are exhausted.
func (c *HintCache) GetOrFetch(hintID, clientID int, api wallarm.API) (*wallarm.ActionBody, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Already cached?
	if h, ok := c.hints[hintID]; ok {
		c.cacheHits++
		return h, nil
	}

	// Already fetched everything? Not found.
	if c.fullyLoaded {
		log.Printf("[DEBUG] HintCache: MISS hint_id=%d (fully loaded, %d hints cached — rule may be deleted)", hintID, len(c.hints))
		return nil, nil
	}

	// Set clientID on first fetch
	if c.clientID == 0 {
		c.clientID = clientID
	}

	// Paginate until found or exhausted
	systemFalse := false
	for {
		log.Printf("[DEBUG] HintCache: fetching page offset=%d limit=%d for hint_id=%d", c.nextOffset, HintBulkFetchLimit, hintID)

		resp, err := api.HintRead(&wallarm.HintRead{
			Limit:     HintBulkFetchLimit,
			Offset:    c.nextOffset,
			OrderBy:   "id",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				System:   &systemFalse,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("HintCache: page fetch failed at offset %d: %w", c.nextOffset, err)
		}

		c.pageFetches++
		c.lastFetchAt = time.Now()

		if resp.Body == nil || len(*resp.Body) == 0 {
			c.fullyLoaded = true
			log.Printf("[DEBUG] HintCache: fully loaded — %d hints cached, hint_id=%d not found", len(c.hints), hintID)
			return nil, nil
		}

		batch := *resp.Body
		for i := range batch {
			if isCredentialStuffingType(batch[i].Type) {
				continue
			}
			c.hints[batch[i].ID] = &batch[i]
		}

		c.nextOffset += HintBulkFetchLimit

		// Found it?
		if h, ok := c.hints[hintID]; ok {
			log.Printf("[DEBUG] HintCache: found hint_id=%d after fetching page (cache size: %d)", hintID, len(c.hints))
			return h, nil
		}

		// Last page?
		if len(batch) < HintBulkFetchLimit {
			c.fullyLoaded = true
			log.Printf("[DEBUG] HintCache: fully loaded — %d hints cached, hint_id=%d not found", len(c.hints), hintID)
			return nil, nil
		}
	}
}

// TODO: add test — mock API returning 2 pages then empty, verify fullyLoaded, credential stuffing filtered
// LoadAll fetches ALL hints into cache. Used by data.wallarm_rules which needs
// the complete set. After this call, fullyLoaded is true.
func (c *HintCache) LoadAll(clientID int, api wallarm.API) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fullyLoaded {
		return nil
	}

	// Continue from where we left off (may already have some pages cached)
	systemFalse := false
	for {
		resp, err := api.HintRead(&wallarm.HintRead{
			Limit:     HintBulkFetchLimit,
			Offset:    c.nextOffset,
			OrderBy:   "id",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				System:   &systemFalse,
			},
		})
		if err != nil {
			return fmt.Errorf("HintCache: LoadAll failed at offset %d: %w", c.nextOffset, err)
		}

		c.pageFetches++
		c.lastFetchAt = time.Now()

		if resp.Body == nil || len(*resp.Body) == 0 {
			break
		}

		batch := *resp.Body
		for i := range batch {
			if isCredentialStuffingType(batch[i].Type) {
				continue
			}
			c.hints[batch[i].ID] = &batch[i]
		}

		c.nextOffset += HintBulkFetchLimit

		if len(batch) < HintBulkFetchLimit {
			break
		}
	}

	c.fullyLoaded = true
	log.Printf("[INFO] HintCache: LoadAll complete — %d hints cached", len(c.hints))
	return nil
}

// TODO: add test — before LoadAll returns nil, after LoadAll returns sorted hints
// All returns all cached hints sorted by ID descending.
// Returns nil if not fully loaded.
func (c *HintCache) All() []wallarm.ActionBody {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fullyLoaded {
		return nil
	}
	result := make([]wallarm.ActionBody, 0, len(c.hints))
	for _, h := range c.hints {
		result = append(result, *h)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID > result[j].ID
	})
	return result
}

// TODO: add test — insert single hint, verify retrievable via GetOrFetch
// Insert adds or updates a single hint in the cache without invalidating.
// Used by HintCreate and HintUpdateV3 to keep the cache warm during batch operations.
func (c *HintCache) Insert(hint *wallarm.ActionBody) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hints == nil {
		c.hints = make(map[int]*wallarm.ActionBody)
	}
	c.hints[hint.ID] = hint
	log.Printf("[DEBUG] HintCache: INSERT hint_id=%d (cache size: %d)", hint.ID, len(c.hints))
}

// Invalidate clears the cache and resets pagination state.
func (c *HintCache) Invalidate(caller string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	prevCount := len(c.hints)
	c.hints = make(map[int]*wallarm.ActionBody)
	c.fullyLoaded = false
	c.nextOffset = 0
	c.invalidations++
	log.Printf("[INFO] HintCache: INVALIDATED by %s — cleared %d cached hints (invalidation #%d)", caller, prevCount, c.invalidations)
}

// trackPassthrough increments the passthrough counter for non-cacheable queries.
func (c *HintCache) trackPassthrough() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.passthroughs++
}

// Stats returns a snapshot of the cache's operational statistics.
func (c *HintCache) Stats() CacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return CacheStats{
		FullyLoaded:   c.fullyLoaded,
		HintCount:     len(c.hints),
		CacheHits:     c.cacheHits,
		PageFetches:   c.pageFetches,
		Passthroughs:  c.passthroughs,
		Invalidations: c.invalidations,
		LastFetchAt:   c.lastFetchAt,
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// CachedClient
// ═══════════════════════════════════════════════════════════════════════════════

// CachedClient wraps a wallarm.API and overrides HintRead with a cache-backed
// implementation. All other API methods delegate to the embedded client.
//
// HintRead is overridden so that single-ID lookups (the pattern used by every
// rule resource's Read function) are served from the lazily-paginated cache.
// Multi-ID or complex filter queries pass through to the underlying API.
//
// Mutating methods:
//   - HintCreate, HintUpdateV3: Insert into cache (no invalidation)
//   - HintDelete: Delegates to underlying API (no caching)
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

// TODO: add test — mock API, verify returns all non-credential-stuffing hints
// AllRules loads all hints into cache and returns them.
// Used by data.wallarm_rules which needs the complete set.
func (c *CachedClient) AllRules(clientID int) ([]wallarm.ActionBody, error) {
	if err := c.hintCache.LoadAll(clientID, c.API); err != nil {
		return nil, err
	}
	return c.hintCache.All(), nil
}

// HintCacheStats returns the current cache statistics.
func (c *CachedClient) HintCacheStats() CacheStats {
	return c.hintCache.Stats()
}

// LogHintCacheStats logs a summary of cache performance.
func (c *CachedClient) LogHintCacheStats() {
	s := c.hintCache.Stats()
	total := s.CacheHits + s.PageFetches + s.Passthroughs
	hitRate := float64(0)
	if s.CacheHits+s.PageFetches > 0 {
		hitRate = float64(s.CacheHits) / float64(s.CacheHits+s.PageFetches) * 100
	}
	log.Printf("[INFO] HintCache stats: %d total | %d hits (%.1f%%) | %d page fetches | %d passthroughs | %d invalidations | %d hints cached",
		total, s.CacheHits, hitRate, s.PageFetches, s.Passthroughs, s.Invalidations, s.HintCount)
}

// HintRead overrides the embedded API's HintRead. Flushes pending deletes first.
// For single-ID lookups,
// it serves from the lazily-paginated cache. For all other query patterns,
// it passes through to the underlying API.
func (c *CachedClient) HintRead(body *wallarm.HintRead) (*wallarm.HintReadResp, error) {
	if body.Filter != nil &&
		len(body.Filter.ID) == 1 &&
		len(body.Filter.Clientid) == 1 &&
		len(body.Filter.ActionID) == 0 &&
		len(body.Filter.Type) == 0 {

		clientID := body.Filter.Clientid[0]
		hintID := body.Filter.ID[0]

		hint, err := c.hintCache.GetOrFetch(hintID, clientID, c.API)
		if err != nil {
			log.Printf("[WARN] HintCache: GetOrFetch failed, falling back to direct API call: %s", err)
			return c.API.HintRead(body)
		}

		if hint == nil {
			return &wallarm.HintReadResp{
				Status: 200,
				Body:   &[]wallarm.ActionBody{},
			}, nil
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

// TODO: add test — mock API, verify response cached via Insert
// HintCreate delegates to the underlying API and inserts the new hint into cache.
func (c *CachedClient) HintCreate(body *wallarm.ActionCreate) (*wallarm.ActionCreateResp, error) {
	resp, err := c.API.HintCreate(body)
	if err != nil {
		return resp, err
	}
	if resp != nil && resp.Body != nil {
		c.hintCache.Insert(resp.Body)
	}
	return resp, nil
}

// HintDelete delegates to the underlying API and invalidates the hint cache
// on success. Without invalidation, a subsequent HintRead could return stale
// entries for the just-deleted rule — surfaces in acceptance tests as
// "dangling resource" failures in CheckDestroy after a Create path that also
// populated the cache (e.g., existingHintForAction → HintRead).
func (c *CachedClient) HintDelete(body *wallarm.HintDelete) error {
	if err := c.API.HintDelete(body); err != nil {
		return err
	}
	c.hintCache.Invalidate("HintDelete")
	return nil
}

// TODO: add test — mock API, verify response cached via Insert
// HintUpdateV3 delegates to the underlying API and updates the cache entry.
func (c *CachedClient) HintUpdateV3(ruleID int, body *wallarm.HintUpdateV3Params) (*wallarm.ActionCreateResp, error) {
	resp, err := c.API.HintUpdateV3(ruleID, body)
	if err != nil {
		return resp, err
	}
	if resp != nil && resp.Body != nil {
		c.hintCache.Insert(resp.Body)
	}
	return resp, nil
}
