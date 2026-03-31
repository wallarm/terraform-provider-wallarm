package wallarm

import (
	"sync"
	"testing"

	wl "github.com/wallarm/wallarm-go"
)

func TestIPListCache_NewIPListCache(t *testing.T) {
	cache := NewIPListCache()
	if cache == nil {
		t.Fatal("NewIPListCache returned nil")
	}
	if cache.entries == nil {
		t.Fatal("entries map is nil")
	}
	if cache.loaded == nil {
		t.Fatal("loaded map is nil")
	}
	if cache.createMu == nil {
		t.Fatal("createMu map is nil")
	}

	// Verify per-list-type mutexes exist for all three list types.
	for _, lt := range []wl.IPListType{wl.DenylistType, wl.AllowlistType, wl.GraylistType} {
		if _, ok := cache.createMu[lt]; !ok {
			t.Errorf("createMu missing entry for list type %q", lt)
		}
	}

	// Empty cache should have zero entries for any list type.
	for _, lt := range []wl.IPListType{wl.DenylistType, wl.AllowlistType, wl.GraylistType} {
		if count := cache.EntryCount(lt); count != 0 {
			t.Errorf("EntryCount(%q) = %d, want 0", lt, count)
		}
	}
}

// populateCache is a test helper that inserts entries directly into the cache's internal map.
func populateCache(cache *IPListCache, listType wl.IPListType, entries map[string]IPCacheEntry) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.entries[listType] == nil {
		cache.entries[listType] = make(map[string]IPCacheEntry)
	}
	for k, v := range entries {
		cache.entries[listType][k] = v
	}
	cache.loaded[listType] = true
}

func TestIPListCache_Lookup_Found(t *testing.T) {
	cache := NewIPListCache()
	entry := IPCacheEntry{GroupID: 100, RuleType: "subnet", RawValue: "10.0.0.1/32"}
	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{
		"10.0.0.1/32": entry,
		"10.0.0.1":    entry,
	})

	got, ok := cache.Lookup(wl.DenylistType, "10.0.0.1/32")
	if !ok {
		t.Fatal("Lookup returned false for existing entry")
	}
	if got.GroupID != 100 {
		t.Errorf("GroupID = %d, want 100", got.GroupID)
	}

	// Lookup by bare IP should also work.
	got, ok = cache.Lookup(wl.DenylistType, "10.0.0.1")
	if !ok {
		t.Fatal("Lookup by bare IP returned false")
	}
	if got.GroupID != 100 {
		t.Errorf("GroupID = %d, want 100", got.GroupID)
	}
}

func TestIPListCache_Lookup_NotFound(t *testing.T) {
	cache := NewIPListCache()
	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{
		"10.0.0.1/32": {GroupID: 100, RuleType: "subnet", RawValue: "10.0.0.1/32"},
	})

	// Different value in same list type.
	_, ok := cache.Lookup(wl.DenylistType, "192.168.1.1")
	if ok {
		t.Error("Lookup should return false for missing value")
	}

	// Same value but different list type (not populated).
	_, ok = cache.Lookup(wl.AllowlistType, "10.0.0.1/32")
	if ok {
		t.Error("Lookup should return false for unpopulated list type")
	}
}

func TestIPListCache_Lookup_EmptyCache(t *testing.T) {
	cache := NewIPListCache()

	_, ok := cache.Lookup(wl.DenylistType, "anything")
	if ok {
		t.Error("Lookup on empty cache should return false")
	}
}

func TestIPListCache_LookupMany_AllFound(t *testing.T) {
	cache := NewIPListCache()
	populateCache(cache, wl.AllowlistType, map[string]IPCacheEntry{
		"US": {GroupID: 200, RuleType: "location", RawValue: "US"},
		"UK": {GroupID: 200, RuleType: "location", RawValue: "UK"},
		"DE": {GroupID: 200, RuleType: "location", RawValue: "DE"},
	})

	found, missing := cache.LookupMany(wl.AllowlistType, []string{"US", "UK", "DE"})
	if len(missing) != 0 {
		t.Errorf("missing = %v, want empty", missing)
	}
	// All three values share GroupID 200, so dedup should yield 1 entry.
	if len(found) != 1 {
		t.Errorf("found %d entries, want 1 (deduplicated by GroupID)", len(found))
	}
	if found[0].GroupID != 200 {
		t.Errorf("GroupID = %d, want 200", found[0].GroupID)
	}
}

func TestIPListCache_LookupMany_SomeMissing(t *testing.T) {
	cache := NewIPListCache()
	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{
		"1.1.1.1/32": {GroupID: 10, RuleType: "subnet", RawValue: "1.1.1.1/32"},
		"2.2.2.2/32": {GroupID: 20, RuleType: "subnet", RawValue: "2.2.2.2/32"},
	})

	found, missing := cache.LookupMany(wl.DenylistType, []string{"1.1.1.1/32", "3.3.3.3/32", "2.2.2.2/32"})
	if len(found) != 2 {
		t.Errorf("found %d entries, want 2", len(found))
	}
	if len(missing) != 1 || missing[0] != "3.3.3.3/32" {
		t.Errorf("missing = %v, want [3.3.3.3/32]", missing)
	}
}

func TestIPListCache_LookupMany_AllMissing(t *testing.T) {
	cache := NewIPListCache()
	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{})

	found, missing := cache.LookupMany(wl.DenylistType, []string{"a", "b"})
	if len(found) != 0 {
		t.Errorf("found = %v, want empty", found)
	}
	if len(missing) != 2 {
		t.Errorf("missing has %d entries, want 2", len(missing))
	}
}

func TestIPListCache_LookupMany_EmptyCache(t *testing.T) {
	cache := NewIPListCache()

	found, missing := cache.LookupMany(wl.GraylistType, []string{"x", "y"})
	if found != nil {
		t.Errorf("found = %v, want nil", found)
	}
	if len(missing) != 2 {
		t.Errorf("missing has %d entries, want 2", len(missing))
	}
}

func TestIPListCache_LookupMany_DeduplicatesByGroupID(t *testing.T) {
	cache := NewIPListCache()
	// Subnets: each IP gets its own group ID.
	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{
		"10.0.0.1/32": {GroupID: 1, RuleType: "subnet", RawValue: "10.0.0.1/32"},
		"10.0.0.1":    {GroupID: 1, RuleType: "subnet", RawValue: "10.0.0.1/32"},
		"10.0.0.2/32": {GroupID: 2, RuleType: "subnet", RawValue: "10.0.0.2/32"},
		"10.0.0.2":    {GroupID: 2, RuleType: "subnet", RawValue: "10.0.0.2/32"},
	})

	// Look up both CIDR and bare forms — should still deduplicate.
	found, missing := cache.LookupMany(wl.DenylistType, []string{"10.0.0.1/32", "10.0.0.1", "10.0.0.2/32"})
	if len(missing) != 0 {
		t.Errorf("missing = %v, want empty", missing)
	}
	// GroupID 1 appears twice in values but should be deduped to 1 entry. GroupID 2 is 1 entry.
	if len(found) != 2 {
		t.Errorf("found %d entries, want 2 (deduped)", len(found))
	}
}

func TestIPListCache_Invalidate(t *testing.T) {
	cache := NewIPListCache()
	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{
		"1.1.1.1": {GroupID: 1, RuleType: "subnet", RawValue: "1.1.1.1/32"},
	})
	populateCache(cache, wl.AllowlistType, map[string]IPCacheEntry{
		"US": {GroupID: 2, RuleType: "location", RawValue: "US"},
	})

	// Invalidate denylist only.
	cache.Invalidate(wl.DenylistType)

	if count := cache.EntryCount(wl.DenylistType); count != 0 {
		t.Errorf("EntryCount(deny) = %d after invalidate, want 0", count)
	}

	// Allowlist should be unaffected.
	if count := cache.EntryCount(wl.AllowlistType); count != 1 {
		t.Errorf("EntryCount(allow) = %d after invalidate of deny, want 1", count)
	}

	// loaded flag should be cleared.
	cache.mu.Lock()
	if cache.loaded[wl.DenylistType] {
		t.Error("loaded[deny] should be false after invalidate")
	}
	if !cache.loaded[wl.AllowlistType] {
		t.Error("loaded[allow] should still be true")
	}
	cache.mu.Unlock()
}

func TestIPListCache_Invalidate_AlreadyEmpty(t *testing.T) {
	cache := NewIPListCache()
	// Should not panic on invalidating a never-populated list type.
	cache.Invalidate(wl.GraylistType)

	if count := cache.EntryCount(wl.GraylistType); count != 0 {
		t.Errorf("EntryCount = %d, want 0", count)
	}
}

func TestIPListCache_EntryCount(t *testing.T) {
	cache := NewIPListCache()

	if count := cache.EntryCount(wl.DenylistType); count != 0 {
		t.Errorf("empty cache EntryCount = %d, want 0", count)
	}

	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{
		"1.1.1.1/32": {GroupID: 1, RuleType: "subnet", RawValue: "1.1.1.1/32"},
		"1.1.1.1":    {GroupID: 1, RuleType: "subnet", RawValue: "1.1.1.1/32"},
		"2.2.2.2/32": {GroupID: 2, RuleType: "subnet", RawValue: "2.2.2.2/32"},
	})

	if count := cache.EntryCount(wl.DenylistType); count != 3 {
		t.Errorf("EntryCount = %d, want 3 (counts map keys, not unique groups)", count)
	}
}

func TestIPListCache_LockCreate_UnlockCreate(_ *testing.T) {
	cache := NewIPListCache()

	// Test that Lock/Unlock works without deadlock for each list type.
	for _, lt := range []wl.IPListType{wl.DenylistType, wl.AllowlistType, wl.GraylistType} {
		cache.LockCreate(lt)
		cache.UnlockCreate(lt)
	}

	// Test that different list types can be locked concurrently (no deadlock).
	cache.LockCreate(wl.DenylistType)
	cache.LockCreate(wl.AllowlistType)
	cache.UnlockCreate(wl.AllowlistType)
	cache.UnlockCreate(wl.DenylistType)
}

func TestIPListCache_LockCreate_Serialization(t *testing.T) {
	cache := NewIPListCache()

	// Verify that LockCreate actually serializes access for the same list type.
	var mu sync.Mutex
	var order []int

	cache.LockCreate(wl.DenylistType)

	done := make(chan struct{})
	go func() {
		cache.LockCreate(wl.DenylistType)
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		cache.UnlockCreate(wl.DenylistType)
		close(done)
	}()

	// Give goroutine time to block on lock.
	// Record that we were first while holding the lock.
	mu.Lock()
	order = append(order, 1)
	mu.Unlock()
	cache.UnlockCreate(wl.DenylistType)

	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("order = %v, want [1 2] (lock should serialize)", order)
	}
}

func TestIPListCache_CrossListTypeIsolation(t *testing.T) {
	cache := NewIPListCache()

	// Populate deny and allow with same key but different entries.
	populateCache(cache, wl.DenylistType, map[string]IPCacheEntry{
		"10.0.0.1": {GroupID: 100, RuleType: "subnet", RawValue: "10.0.0.1/32"},
	})
	populateCache(cache, wl.AllowlistType, map[string]IPCacheEntry{
		"10.0.0.1": {GroupID: 200, RuleType: "subnet", RawValue: "10.0.0.1/32"},
	})

	denyEntry, ok := cache.Lookup(wl.DenylistType, "10.0.0.1")
	if !ok || denyEntry.GroupID != 100 {
		t.Errorf("deny entry GroupID = %d, want 100", denyEntry.GroupID)
	}

	allowEntry, ok := cache.Lookup(wl.AllowlistType, "10.0.0.1")
	if !ok || allowEntry.GroupID != 200 {
		t.Errorf("allow entry GroupID = %d, want 200", allowEntry.GroupID)
	}

	// Invalidate deny — allow should remain.
	cache.Invalidate(wl.DenylistType)

	_, ok = cache.Lookup(wl.DenylistType, "10.0.0.1")
	if ok {
		t.Error("deny entry should be gone after invalidate")
	}

	allowEntry, ok = cache.Lookup(wl.AllowlistType, "10.0.0.1")
	if !ok || allowEntry.GroupID != 200 {
		t.Error("allow entry should survive deny invalidate")
	}
}
