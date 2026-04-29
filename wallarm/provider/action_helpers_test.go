package wallarm

import (
	"testing"

	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

func TestImportAsExistsError(t *testing.T) {
	err := ImportAsExistsError("wallarm_rule_mode", "1/2/3")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

// makeAction returns a single-condition iequal HOST action whose
// ConditionsHash is deterministic for the given host string.
func makeAction(host string) []wallarm.ActionDetails {
	return []wallarm.ActionDetails{
		{
			Type:  "iequal",
			Value: host,
			Point: []interface{}{"header", "HOST"},
		},
	}
}

// fillerActions produces n actions with distinct conditions that will not
// collide with any host passed to makeAction.
func fillerActions(n, idStart int) []wallarm.ActionEntry {
	out := make([]wallarm.ActionEntry, n)
	for i := range n {
		out[i] = wallarm.ActionEntry{
			ID:       idStart + i,
			Clientid: 1,
			Conditions: []wallarm.ActionDetails{
				{Type: "iequal", Value: "filler", Point: []interface{}{"path", float64(i)}},
			},
		}
	}
	return out
}

// TestFindActionByConditionsHash_Page2 verifies that pagination finds a match
// on page 2 — single-page lookup (the pre-fix behavior) silently misses it.
// Failure mode: tenants with >APIListLimit actions of a given hint_type
// could create same-scope duplicates that existingHintForAction never detects.
func TestFindActionByConditionsHash_Page2(t *testing.T) {
	host := "wanted.example.com"
	wantedConditions := makeAction(host)
	wantHash := resourcerule.ConditionsHash(wantedConditions)

	// Page 1: APIListLimit fillers. Page 2: 1 filler + 1 matching action.
	page1 := fillerActions(APIListLimit, 100)
	page2 := append(fillerActions(1, 100+APIListLimit), wallarm.ActionEntry{
		ID:         99999,
		Clientid:   1,
		Conditions: wantedConditions,
	})
	mock := &mockHintAPI{actions: append(page1, page2...)}

	got, err := findActionByConditionsHash(mock, 1, "wallarm_mode", wantHash, len(wantedConditions))
	if err != nil {
		t.Fatalf("findActionByConditionsHash: %v", err)
	}
	if got == nil {
		t.Fatalf("expected match (action 99999) on page 2; got nil — pagination missing")
	}
	if got.ID != 99999 {
		t.Fatalf("expected ID 99999, got %d", got.ID)
	}
	if mock.actionCallCount.Load() < 2 {
		t.Fatalf("expected at least 2 ActionList calls (pagination), got %d", mock.actionCallCount.Load())
	}
}

// TestFindActionByConditionsHash_NoMatch verifies the no-match path: pagination
// stops after a short page and returns nil.
func TestFindActionByConditionsHash_NoMatch(t *testing.T) {
	wantHash := resourcerule.ConditionsHash(makeAction("never-matches.example.com"))
	mock := &mockHintAPI{actions: fillerActions(50, 100)} // <APIListLimit, single short page

	got, err := findActionByConditionsHash(mock, 1, "wallarm_mode", wantHash, 1)
	if err != nil {
		t.Fatalf("findActionByConditionsHash: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil (no match), got ID=%d", got.ID)
	}
	if mock.actionCallCount.Load() != 1 {
		t.Fatalf("expected exactly 1 ActionList call (short page), got %d", mock.actionCallCount.Load())
	}
}

// TestFindActionByConditionsHash_Page1Match verifies the common-case fast path:
// match on page 1 returns immediately without paginating further.
func TestFindActionByConditionsHash_Page1Match(t *testing.T) {
	host := "first.example.com"
	wantedConditions := makeAction(host)
	wantHash := resourcerule.ConditionsHash(wantedConditions)

	// Match is at index 5 of page 1; pagination MUST NOT continue past page 1.
	page1 := fillerActions(10, 100)
	page1[5] = wallarm.ActionEntry{ID: 12345, Clientid: 1, Conditions: wantedConditions}
	mock := &mockHintAPI{actions: page1}

	got, err := findActionByConditionsHash(mock, 1, "wallarm_mode", wantHash, len(wantedConditions))
	if err != nil {
		t.Fatalf("findActionByConditionsHash: %v", err)
	}
	if got == nil || got.ID != 12345 {
		t.Fatalf("expected ID 12345 from page 1, got %v", got)
	}
	if mock.actionCallCount.Load() != 1 {
		t.Fatalf("expected exactly 1 ActionList call (matched on page 1), got %d", mock.actionCallCount.Load())
	}
}
