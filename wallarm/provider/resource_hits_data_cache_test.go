package wallarm

import (
	"encoding/json"
	"testing"
)

// ─── Unit tests ─────────────────────────────────────────────────────────────

func TestMergeAggregatedGroups_NewAction(t *testing.T) {
	cache := map[string]string{}
	reqToAction := map[string]string{}

	entry := `{"action_hash":"abc12345","action":[{"type":"iequal","value":"example.com","point":{"header":"HOST"}}],"groups":[{"key":"f1e2d3c4_xss","point":[["header","User-Agent"]],"stamps":[7994],"attack_type":"xss"}]}`

	if err := mergeNewEntry(cache, reqToAction, "req1", entry); err != nil {
		t.Fatalf("mergeNewEntry failed: %v", err)
	}

	if len(cache) != 1 {
		t.Fatalf("expected 1 cache entry, got %d", len(cache))
	}
	if reqToAction["req1"] != "abc12345" {
		t.Errorf("expected reqToAction[req1] = abc12345, got %q", reqToAction["req1"])
	}

	var parsed aggregatedOutput
	if err := json.Unmarshal([]byte(cache["abc12345"]), &parsed); err != nil {
		t.Fatalf("failed to parse cached JSON: %v", err)
	}
	if len(parsed.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(parsed.Groups))
	}
	if parsed.Groups[0].Stamps[0] != 7994 {
		t.Errorf("expected stamp 7994, got %d", parsed.Groups[0].Stamps[0])
	}
}

func TestMergeAggregatedGroups_MergeStamps(t *testing.T) {
	cache := map[string]string{}
	reqToAction := map[string]string{}

	entry1 := `{"action_hash":"abc12345","action":[{"type":"iequal","value":"example.com","point":{"header":"HOST"}}],"groups":[{"key":"f1e2d3c4_xss","point":[["header","User-Agent"]],"stamps":[7994],"attack_type":"xss"}]}`
	entry2 := `{"action_hash":"abc12345","action":[{"type":"iequal","value":"example.com","point":{"header":"HOST"}}],"groups":[{"key":"f1e2d3c4_xss","point":[["header","User-Agent"]],"stamps":[7994,8001],"attack_type":"xss"},{"key":"a5b6c7d8_sqli","point":[["get","search"]],"stamps":[1234],"attack_type":"sqli"}]}`

	if err := mergeNewEntry(cache, reqToAction, "req1", entry1); err != nil {
		t.Fatalf("mergeNewEntry(req1) failed: %v", err)
	}
	if err := mergeNewEntry(cache, reqToAction, "req2", entry2); err != nil {
		t.Fatalf("mergeNewEntry(req2) failed: %v", err)
	}

	if len(cache) != 1 {
		t.Fatalf("expected 1 cache entry (same action), got %d", len(cache))
	}

	var parsed aggregatedOutput
	if err := json.Unmarshal([]byte(cache["abc12345"]), &parsed); err != nil {
		t.Fatalf("failed to parse cached JSON: %v", err)
	}
	if len(parsed.Groups) != 2 {
		t.Fatalf("expected 2 groups after merge, got %d", len(parsed.Groups))
	}

	for _, g := range parsed.Groups {
		if g.Key == "f1e2d3c4_xss" {
			if len(g.Stamps) != 2 {
				t.Errorf("expected 2 stamps in xss group, got %d: %v", len(g.Stamps), g.Stamps)
			}
			return
		}
	}
	t.Error("xss group not found after merge")
}

func TestCleanupOrphanedCacheEntries(t *testing.T) {
	cache := map[string]string{
		"abc12345": `{"action_hash":"abc12345","action":[],"groups":[]}`,
		"def67890": `{"action_hash":"def67890","action":[],"groups":[]}`,
	}
	reqToAction := map[string]string{
		"req1": "abc12345",
		"req2": "abc12345",
		"req3": "def67890",
	}
	activeRequestIDs := map[string]bool{"req1": true}

	cleanupCache(cache, reqToAction, activeRequestIDs)

	if _, ok := reqToAction["req2"]; ok {
		t.Error("expected req2 removed from reqToAction")
	}
	if _, ok := reqToAction["req3"]; ok {
		t.Error("expected req3 removed from reqToAction")
	}
	if _, ok := cache["abc12345"]; !ok {
		t.Error("expected abc12345 preserved (req1 still references it)")
	}
	if _, ok := cache["def67890"]; ok {
		t.Error("expected def67890 removed (no references)")
	}
}

func TestBuildCachedRequestIDs(t *testing.T) {
	reqToAction := map[string]string{
		"req1": "abc",
		"req3": "def",
	}
	got := buildCachedRequestIDs(reqToAction)
	want := "req1,req3"
	if got != want {
		t.Errorf("buildCachedRequestIDs = %q, want %q", got, want)
	}
}

func TestBuildCachedRequestIDs_Empty(t *testing.T) {
	reqToAction := map[string]string{}
	got := buildCachedRequestIDs(reqToAction)
	if got != "" {
		t.Errorf("buildCachedRequestIDs = %q, want empty", got)
	}
}

func TestMergeAndCleanup_EmptyNewEntries(t *testing.T) {
	cache := map[string]string{}
	reqToAction := map[string]string{}
	activeIDs := map[string]bool{"abc123": true, "def456": true}

	cleanupCache(cache, reqToAction, activeIDs)

	if len(cache) != 0 {
		t.Errorf("expected empty cache, got %d entries", len(cache))
	}
	if len(reqToAction) != 0 {
		t.Errorf("expected empty reqToAction, got %d entries", len(reqToAction))
	}
}

func TestMergeAndCleanup_WithNewEntries(t *testing.T) {
	cache := map[string]string{}
	reqToAction := map[string]string{}

	entry := `{"action_hash":"abc12345","action":[],"groups":[{"key":"f1_xss","point":[],"stamps":[1],"attack_type":"xss"}]}`
	if err := mergeNewEntry(cache, reqToAction, "req1", entry); err != nil {
		t.Fatalf("mergeNewEntry failed: %v", err)
	}

	activeIDs := map[string]bool{"req1": true, "req2": true}
	cleanupCache(cache, reqToAction, activeIDs)

	if reqToAction["req1"] != "abc12345" {
		t.Errorf("reqToAction[req1] = %q, want abc12345", reqToAction["req1"])
	}
	if len(cache) != 1 {
		t.Errorf("expected 1 cache entry, got %d", len(cache))
	}
}

func TestMergeGroups_EmptyStamps(t *testing.T) {
	existing := []aggregatedGroup{
		{Key: "f1_xss", Point: [][]string{{"header", "User-Agent"}}, Stamps: []int{}, AttackType: "xss"},
	}
	incoming := []aggregatedGroup{
		{Key: "f1_xss", Point: [][]string{{"header", "User-Agent"}}, Stamps: []int{}, AttackType: "xss"},
	}

	merged := mergeGroups(existing, incoming)
	if len(merged) != 1 {
		t.Fatalf("expected 1 group, got %d", len(merged))
	}
	if len(merged[0].Stamps) != 0 {
		t.Errorf("expected empty stamps, got %v", merged[0].Stamps)
	}
}
