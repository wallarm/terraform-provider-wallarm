package wallarm

import (
	"fmt"
	"sync/atomic"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

// mockCredStuffAPI tracks calls to CredentialStuffingConfigsRead.
type mockCredStuffAPI struct {
	wallarm.API
	configs   []wallarm.ActionBody
	callCount atomic.Int32
}

func (m *mockCredStuffAPI) CredentialStuffingConfigsRead(clientID int) ([]wallarm.ActionBody, error) {
	m.callCount.Add(1)
	return m.configs, nil
}

func makeCredStuffConfigs(ids ...int) []wallarm.ActionBody {
	configs := make([]wallarm.ActionBody, len(ids))
	for i, id := range ids {
		configs[i] = wallarm.ActionBody{
			ID:       id,
			Clientid: 1,
			Type:     "credential_stuffing_regex",
		}
	}
	return configs
}

func TestCredentialStuffingCache_GetOrFetch_CachesAfterFirstCall(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100, 200, 300)}
	cache := NewCredentialStuffingCache()

	// First call — fetches from API.
	rule, err := cache.GetOrFetch(mock, 1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID != 100 {
		t.Fatalf("expected rule ID 100, got %d", rule.ID)
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected 1 API call, got %d", mock.callCount.Load())
	}

	// Second call for different ID — cache hit, no API call.
	rule, err = cache.GetOrFetch(mock, 1, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID != 200 {
		t.Fatalf("expected rule ID 200, got %d", rule.ID)
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected still 1 API call, got %d", mock.callCount.Load())
	}

	// Third call for same ID — still cache hit.
	_, err = cache.GetOrFetch(mock, 1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected still 1 API call, got %d", mock.callCount.Load())
	}
}

func TestCredentialStuffingCache_GetOrFetch_NotFound(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100)}
	cache := NewCredentialStuffingCache()

	_, err := cache.GetOrFetch(mock, 1, 999)
	if err == nil {
		t.Fatal("expected error for missing rule, got nil")
	}
	if _, ok := err.(*ruleNotFoundError); !ok {
		t.Fatalf("expected ruleNotFoundError, got %T: %v", err, err)
	}
}

func TestCredentialStuffingCache_GetOrFetch_NotFoundAfterCacheLoaded(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100, 200)}
	cache := NewCredentialStuffingCache()

	// Load cache.
	_, _ = cache.GetOrFetch(mock, 1, 100)

	// Lookup missing ID — should return error without API call.
	_, err := cache.GetOrFetch(mock, 1, 999)
	if err == nil {
		t.Fatal("expected error for missing rule, got nil")
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected 1 API call (no re-fetch for miss), got %d", mock.callCount.Load())
	}
}

func TestCredentialStuffingCache_Invalidate_ForcesRefetch(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100)}
	cache := NewCredentialStuffingCache()

	_, _ = cache.GetOrFetch(mock, 1, 100)
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected 1 API call, got %d", mock.callCount.Load())
	}

	cache.Invalidate()

	_, _ = cache.GetOrFetch(mock, 1, 100)
	if mock.callCount.Load() != 2 {
		t.Fatalf("expected 2 API calls after invalidation, got %d", mock.callCount.Load())
	}
}

func TestCredentialStuffingCache_DifferentClientID_Refetches(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100)}
	cache := NewCredentialStuffingCache()

	_, _ = cache.GetOrFetch(mock, 1, 100)
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected 1 API call, got %d", mock.callCount.Load())
	}

	// Different client ID — must re-fetch.
	_, _ = cache.GetOrFetch(mock, 2, 100)
	if mock.callCount.Load() != 2 {
		t.Fatalf("expected 2 API calls for different client, got %d", mock.callCount.Load())
	}
}

func TestCredentialStuffingCache_LoadAll_CachesForSubsequentGetOrFetch(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100, 200, 300)}
	cache := NewCredentialStuffingCache()

	// LoadAll fetches from API.
	configs, err := cache.LoadAll(mock, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 3 {
		t.Fatalf("expected 3 configs, got %d", len(configs))
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected 1 API call, got %d", mock.callCount.Load())
	}

	// GetOrFetch should use cache — no API call.
	rule, err := cache.GetOrFetch(mock, 1, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID != 200 {
		t.Fatalf("expected rule ID 200, got %d", rule.ID)
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected still 1 API call, got %d", mock.callCount.Load())
	}
}

func TestCredentialStuffingCache_LoadAll_UsesCacheOnSecondCall(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100, 200)}
	cache := NewCredentialStuffingCache()

	_, _ = cache.LoadAll(mock, 1)
	_, _ = cache.LoadAll(mock, 1)

	if mock.callCount.Load() != 1 {
		t.Fatalf("expected 1 API call (second LoadAll should use cache), got %d", mock.callCount.Load())
	}
}

func TestCredentialStuffingCache_LoadAll_RefetchesAfterInvalidate(t *testing.T) {
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(100)}
	cache := NewCredentialStuffingCache()

	_, _ = cache.LoadAll(mock, 1)
	cache.Invalidate()
	_, _ = cache.LoadAll(mock, 1)

	if mock.callCount.Load() != 2 {
		t.Fatalf("expected 2 API calls after invalidation, got %d", mock.callCount.Load())
	}
}

func TestCredentialStuffingCache_MultipleResources_SingleAPICall(t *testing.T) {
	// Simulate terraform plan with 4 credential stuffing resources.
	mock := &mockCredStuffAPI{configs: makeCredStuffConfigs(10, 20, 30, 40)}
	cache := NewCredentialStuffingCache()

	ids := []int{10, 20, 30, 40}
	for _, id := range ids {
		rule, err := cache.GetOrFetch(mock, 1, id)
		if err != nil {
			t.Fatalf("unexpected error for rule %d: %v", id, err)
		}
		if rule.ID != id {
			t.Fatalf("expected rule ID %d, got %d", id, rule.ID)
		}
	}

	if calls := mock.callCount.Load(); calls != 1 {
		t.Fatalf("expected 1 API call for 4 resources, got %d", calls)
	}
}

func TestCredentialStuffingCache_CreateThenRead_RequiresInvalidate(t *testing.T) {
	// Simulates the bug: cache is loaded with rule 100 (from another resource's Read).
	// Then HintCreate adds rule 200 (via a different API), but the cache doesn't know about it.
	// Without Invalidate, GetOrFetch returns "not found" because rule 200 isn't in the stale cache.
	// With Invalidate, GetOrFetch re-fetches and finds the new rule.

	// Phase 1: cache loaded with only rule 100.
	initialConfigs := makeCredStuffConfigs(100)
	mock := &mockCredStuffAPIDynamic{configs: initialConfigs}
	cache := NewCredentialStuffingCache()

	rule, err := cache.GetOrFetch(mock, 1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID != 100 {
		t.Fatalf("expected rule ID 100, got %d", rule.ID)
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected 1 API call, got %d", mock.callCount.Load())
	}

	// Phase 2: HintCreate adds rule 200 — simulate by updating the mock's data.
	// The cache still has the old snapshot (only rule 100).
	mock.configs = makeCredStuffConfigs(100, 200)

	// Without invalidation, rule 200 is NOT found.
	_, err = cache.GetOrFetch(mock, 1, 200)
	if err == nil {
		t.Fatal("expected error without invalidation, got nil")
	}
	if _, ok := err.(*ruleNotFoundError); !ok {
		t.Fatalf("expected ruleNotFoundError, got %T: %v", err, err)
	}
	// No extra API call — served from stale cache.
	if mock.callCount.Load() != 1 {
		t.Fatalf("expected still 1 API call (stale cache), got %d", mock.callCount.Load())
	}

	// Phase 3: Invalidate (what Create should do), then Read succeeds.
	cache.Invalidate()

	rule, err = cache.GetOrFetch(mock, 1, 200)
	if err != nil {
		t.Fatalf("unexpected error after invalidation: %v", err)
	}
	if rule.ID != 200 {
		t.Fatalf("expected rule ID 200, got %d", rule.ID)
	}
	if mock.callCount.Load() != 2 {
		t.Fatalf("expected 2 API calls (re-fetch after invalidation), got %d", mock.callCount.Load())
	}
}

// mockCredStuffAPIDynamic allows changing configs between calls to simulate Create adding new rules.
type mockCredStuffAPIDynamic struct {
	wallarm.API
	configs   []wallarm.ActionBody
	callCount atomic.Int32
}

func (m *mockCredStuffAPIDynamic) CredentialStuffingConfigsRead(clientID int) ([]wallarm.ActionBody, error) {
	m.callCount.Add(1)
	return m.configs, nil
}

func TestCredentialStuffingCache_APIError_DoesNotCache(t *testing.T) {
	mock := &mockCredStuffAPIWithError{err: fmt.Errorf("API unavailable")}
	cache := NewCredentialStuffingCache()

	_, err := cache.GetOrFetch(mock, 1, 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Cache should not be loaded — next call should retry.
	if cache.loaded {
		t.Fatal("cache should not be marked loaded after API error")
	}
}

type mockCredStuffAPIWithError struct {
	wallarm.API
	err error
}

func (m *mockCredStuffAPIWithError) CredentialStuffingConfigsRead(clientID int) ([]wallarm.ActionBody, error) {
	return nil, m.err
}
