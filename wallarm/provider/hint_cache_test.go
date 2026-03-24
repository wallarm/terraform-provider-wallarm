package wallarm

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

// mockAPI implements wallarm.API by embedding nil and only overriding HintRead.
// We use a concrete mock that tracks call counts to verify caching behavior.
type mockHintAPI struct {
	wallarm.API // embed to satisfy interface; only HintRead is called
	hints       []wallarm.ActionBody
	callCount   atomic.Int32
	failOnCall  int // if > 0, fail on the Nth call
}

func (m *mockHintAPI) HintRead(body *wallarm.HintRead) (*wallarm.HintReadResp, error) {
	n := int(m.callCount.Add(1))
	if m.failOnCall > 0 && n == m.failOnCall {
		return nil, fmt.Errorf("simulated API error on call %d", n)
	}

	// Paginate based on offset/limit
	offset := body.Offset
	limit := body.Limit
	if offset >= len(m.hints) {
		return &wallarm.HintReadResp{Status: 200, Body: &[]wallarm.ActionBody{}}, nil
	}
	end := offset + limit
	if end > len(m.hints) {
		end = len(m.hints)
	}

	// If filtering by specific IDs, return only matching hints
	if body.Filter != nil && len(body.Filter.ID) > 0 {
		idSet := make(map[int]bool)
		for _, id := range body.Filter.ID {
			idSet[id] = true
		}
		var filtered []wallarm.ActionBody
		for _, h := range m.hints {
			if idSet[h.ID] {
				filtered = append(filtered, h)
			}
		}
		return &wallarm.HintReadResp{Status: 200, Body: &filtered}, nil
	}

	page := m.hints[offset:end]
	return &wallarm.HintReadResp{Status: 200, Body: &page}, nil
}

func makeHints(n int) []wallarm.ActionBody {
	hints := make([]wallarm.ActionBody, n)
	for i := 0; i < n; i++ {
		hints[i] = wallarm.ActionBody{
			ID:       1000 + i,
			ActionID: 2000 + i,
			Clientid: 1,
			Type:     "disable_stamp",
		}
	}
	return hints
}

func TestHintCache_BulkLoadReducesAPICalls(t *testing.T) {
	hints := makeHints(250)
	mock := &mockHintAPI{hints: hints}
	cached := NewCachedClient(mock)

	// Read 250 hints individually — should trigger 1 bulk load (2 pages of 200)
	// then all reads served from cache.
	for _, h := range hints {
		body := &wallarm.HintRead{
			Limit:     APIListLimit,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{1},
				ID:       []int{h.ID},
			},
		}
		resp, err := cached.HintRead(body)
		if err != nil {
			t.Fatalf("HintRead for ID %d failed: %v", h.ID, err)
		}
		if resp.Body == nil || len(*resp.Body) != 1 {
			t.Fatalf("expected 1 hint for ID %d, got %d", h.ID, len(*resp.Body))
		}
		if (*resp.Body)[0].ID != h.ID {
			t.Fatalf("expected hint ID %d, got %d", h.ID, (*resp.Body)[0].ID)
		}
	}

	// With 250 hints and batch size 200, bulk load should make 2 API calls
	// (200 + 50). All 250 individual reads come from cache.
	calls := int(mock.callCount.Load())
	if calls != 2 {
		t.Errorf("expected 2 API calls for bulk load, got %d", calls)
	}

	// Verify stats
	stats := cached.HintCacheStats()
	// 250 hints read, but 2 were found during page fetches (not cache hits):
	// first read triggers page 1 (200 hints), ID 1200 triggers page 2 (50 hints)
	if stats.CacheHits != 248 {
		t.Errorf("expected 248 cache hits, got %d", stats.CacheHits)
	}
	// 2 page fetches: page 1 (200 hints at offset 0), page 2 (50 hints at offset 200)
	if stats.PageFetches != 2 {
		t.Errorf("expected 2 page fetches, got %d", stats.PageFetches)
	}
	if stats.HintCount != 250 {
		t.Errorf("expected 250 hints in cache, got %d", stats.HintCount)
	}
	// Note: cache may not be FullyLoaded because page 2 found the hint
	// and returned early before checking if it was the last page.
}

func TestHintCache_ConcurrentReads(t *testing.T) {
	hints := makeHints(150)
	mock := &mockHintAPI{hints: hints}
	cached := NewCachedClient(mock)

	var wg sync.WaitGroup
	errors := make(chan error, len(hints))

	// Simulate Terraform's parallel reads (default parallelism=10)
	for _, h := range hints {
		wg.Add(1)
		go func(hintID int) {
			defer wg.Done()
			body := &wallarm.HintRead{
				Limit:     APIListLimit,
				Offset:    0,
				OrderBy:   "updated_at",
				OrderDesc: true,
				Filter: &wallarm.HintFilter{
					Clientid: []int{1},
					ID:       []int{hintID},
				},
			}
			resp, err := cached.HintRead(body)
			if err != nil {
				errors <- fmt.Errorf("HintRead for ID %d failed: %v", hintID, err)
				return
			}
			if resp.Body == nil || len(*resp.Body) != 1 {
				errors <- fmt.Errorf("expected 1 hint for ID %d", hintID)
			}
		}(h.ID)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	// Only the bulk load calls should have been made (1 page of 200 covers all 150)
	calls := int(mock.callCount.Load())
	if calls != 1 {
		t.Errorf("expected 1 API call for bulk load, got %d", calls)
	}
}

func TestHintCache_InvalidateOnMutation(t *testing.T) {
	hints := makeHints(50)
	mock := &mockHintAPI{hints: hints}
	cached := NewCachedClient(mock)

	// Initial read — triggers bulk load (1 API call for 50 hints)
	body := &wallarm.HintRead{
		Limit: APIListLimit, Offset: 0, OrderBy: "updated_at", OrderDesc: true,
		Filter: &wallarm.HintFilter{Clientid: []int{1}, ID: []int{1000}},
	}
	_, err := cached.HintRead(body)
	if err != nil {
		t.Fatalf("initial read failed: %v", err)
	}
	callsAfterLoad := int(mock.callCount.Load())

	// Invalidate via HintDelete (we expect a panic since mock doesn't implement
	// HintDelete, so test Invalidate directly)
	cached.hintCache.Invalidate("test")

	// Next read should trigger another bulk load
	_, err = cached.HintRead(body)
	if err != nil {
		t.Fatalf("post-invalidation read failed: %v", err)
	}

	callsAfterReload := int(mock.callCount.Load())
	reloadCalls := callsAfterReload - callsAfterLoad
	if reloadCalls != 1 {
		t.Errorf("expected 1 API call after invalidation, got %d", reloadCalls)
	}
}

func TestHintCache_CacheMissFallsBackToAPI(t *testing.T) {
	hints := makeHints(10)
	mock := &mockHintAPI{hints: hints}
	cached := NewCachedClient(mock)

	// Load cache with IDs 1000-1009
	body := &wallarm.HintRead{
		Limit: APIListLimit, Offset: 0, OrderBy: "updated_at", OrderDesc: true,
		Filter: &wallarm.HintFilter{Clientid: []int{1}, ID: []int{1000}},
	}
	_, err := cached.HintRead(body)
	if err != nil {
		t.Fatalf("initial read failed: %v", err)
	}

	// Read a hint ID that doesn't exist in cache — should fall back to direct API
	body.Filter.ID = []int{9999}
	resp, err := cached.HintRead(body)
	if err != nil {
		t.Fatalf("cache miss read failed: %v", err)
	}
	// The mock returns empty for ID 9999 since it doesn't exist
	if resp.Body != nil && len(*resp.Body) > 0 {
		t.Errorf("expected empty response for non-existent hint, got %d results", len(*resp.Body))
	}
}

func TestHintCache_NonCacheableQueryPassesThrough(t *testing.T) {
	hints := makeHints(10)
	mock := &mockHintAPI{hints: hints}
	cached := NewCachedClient(mock)

	// Query with multiple IDs — should NOT use cache, pass through directly
	body := &wallarm.HintRead{
		Limit: APIListLimit, Offset: 0, OrderBy: "updated_at", OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{1},
			ID:       []int{1000, 1001, 1002},
		},
	}
	resp, err := cached.HintRead(body)
	if err != nil {
		t.Fatalf("multi-ID read failed: %v", err)
	}
	if resp.Body == nil || len(*resp.Body) != 3 {
		t.Fatalf("expected 3 hints from passthrough, got %d", len(*resp.Body))
	}

	// Query with type filter — should NOT use cache
	body2 := &wallarm.HintRead{
		Limit: APIListLimit, Offset: 0, OrderBy: "updated_at", OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{1},
			ID:       []int{1000},
			Type:     []string{"disable_stamp"},
		},
	}
	_, err = cached.HintRead(body2)
	if err != nil {
		t.Fatalf("type-filtered read failed: %v", err)
	}

	// Both should have been direct API calls (no bulk load)
	calls := int(mock.callCount.Load())
	if calls != 2 {
		t.Errorf("expected 2 direct API calls for non-cacheable queries, got %d", calls)
	}

	// Verify passthrough stats
	stats := cached.HintCacheStats()
	if stats.Passthroughs != 2 {
		t.Errorf("expected 2 passthroughs, got %d", stats.Passthroughs)
	}
	if stats.CacheHits != 0 {
		t.Errorf("expected 0 cache hits for passthrough queries, got %d", stats.CacheHits)
	}
}

func TestHintCache_BulkLoadFailureFallsBack(t *testing.T) {
	hints := makeHints(10)
	mock := &mockHintAPI{hints: hints, failOnCall: 1} // fail on first call (bulk load)
	cached := NewCachedClient(mock)

	// The bulk load will fail, but the fallback direct API call should work
	// (failOnCall=1 only fails the first call)
	body := &wallarm.HintRead{
		Limit: APIListLimit, Offset: 0, OrderBy: "updated_at", OrderDesc: true,
		Filter: &wallarm.HintFilter{Clientid: []int{1}, ID: []int{1000}},
	}
	resp, err := cached.HintRead(body)
	if err != nil {
		t.Fatalf("fallback read should have succeeded: %v", err)
	}
	if resp.Body == nil || len(*resp.Body) != 1 {
		t.Fatalf("expected 1 hint from fallback, got %d", len(*resp.Body))
	}
}

func TestHintCache_StatsTrackInvalidationCycle(t *testing.T) {
	hints := makeHints(50)
	mock := &mockHintAPI{hints: hints}
	cached := NewCachedClient(mock)

	// Phase 1: Read 10 hints — triggers bulk load, 10 cache hits
	for i := 0; i < 10; i++ {
		body := &wallarm.HintRead{
			Limit: APIListLimit, Offset: 0, OrderBy: "updated_at", OrderDesc: true,
			Filter: &wallarm.HintFilter{Clientid: []int{1}, ID: []int{hints[i].ID}},
		}
		if _, err := cached.HintRead(body); err != nil {
			t.Fatalf("read %d failed: %v", i, err)
		}
	}

	s1 := cached.HintCacheStats()
	// 1 page fetch loads all 50 hints (50 < 200 batch size → fully loaded)
	if s1.PageFetches != 1 {
		t.Errorf("phase 1: expected 1 page fetch, got %d", s1.PageFetches)
	}
	// First read triggers page fetch (not a cache hit), remaining 9 are cache hits
	if s1.CacheHits != 9 {
		t.Errorf("phase 1: expected 9 cache hits, got %d", s1.CacheHits)
	}
	if s1.Invalidations != 0 {
		t.Errorf("phase 1: expected 0 invalidations, got %d", s1.Invalidations)
	}

	// Phase 2: Invalidate (simulating a create)
	cached.hintCache.Invalidate("test")
	s2 := cached.HintCacheStats()
	if s2.Invalidations != 1 {
		t.Errorf("phase 2: expected 1 invalidation, got %d", s2.Invalidations)
	}
	if s2.FullyLoaded {
		t.Error("phase 2: cache should not be loaded after invalidation")
	}

	// Phase 3: Read 5 more — triggers second bulk load, 5 more cache hits
	for i := 0; i < 5; i++ {
		body := &wallarm.HintRead{
			Limit: APIListLimit, Offset: 0, OrderBy: "updated_at", OrderDesc: true,
			Filter: &wallarm.HintFilter{Clientid: []int{1}, ID: []int{hints[i].ID}},
		}
		if _, err := cached.HintRead(body); err != nil {
			t.Fatalf("post-invalidation read %d failed: %v", i, err)
		}
	}

	s3 := cached.HintCacheStats()
	// 2 total page fetches: 1 in phase 1, 1 after invalidation in phase 3
	if s3.PageFetches != 2 {
		t.Errorf("phase 3: expected 2 page fetches, got %d", s3.PageFetches)
	}
	// 9 cache hits in phase 1 + 4 cache hits in phase 3 = 13 total
	// (first read in each phase triggers page fetch, not counted as cache hit)
	if s3.CacheHits != 13 {
		t.Errorf("phase 3: expected 13 total cache hits, got %d", s3.CacheHits)
	}

	// Verify LogHintCacheStats doesn't panic
	cached.LogHintCacheStats()
}
