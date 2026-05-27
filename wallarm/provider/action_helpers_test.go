package wallarm

import (
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

// newRuleResourceDataForTest builds a minimal *ResourceData with the action
// TypeSet schema populated from the given action map. Used by
// existingHintForAction tests to avoid pulling in any specific rule resource's
// full schema.
func newRuleResourceDataForTest(t *testing.T, actionMap map[string]any, clientID int) *schema.ResourceData {
	t.Helper()
	res := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action":    resourcerule.ScopeActionSchema(),
			"client_id": {Type: schema.TypeInt, Optional: true},
		},
	}
	return schema.TestResourceDataRaw(t, res.Schema, map[string]any{
		"action":    []any{actionMap},
		"client_id": clientID,
	})
}

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
			Point: []any{"header", "HOST"},
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
				{Type: "iequal", Value: "filler", Point: []any{"path", float64(i)}},
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

// degenerateActionListAPI always returns Limit-sized pages regardless of
// Offset, simulating a misbehaving API that would cause the unbounded
// pagination loop in findActionByConditionsHash to spin forever.
type degenerateActionListAPI struct {
	wallarm.API
	actionCallCount atomic.Int32
}

func (m *degenerateActionListAPI) ActionList(params *wallarm.ActionListParams) (*wallarm.ActionListResponse, error) {
	m.actionCallCount.Add(1)
	return &wallarm.ActionListResponse{Status: 200, Body: fillerActions(params.Limit, 100)}, nil
}

// TestFindActionByConditionsHash_PageCap guards against the unbounded loop
// where a degenerate API serves always-full pages without progress. The page
// cap must trip with an error instead of hanging the apply.
func TestFindActionByConditionsHash_PageCap(t *testing.T) {
	mock := &degenerateActionListAPI{}
	got, err := findActionByConditionsHash(mock, 1, "wallarm_mode", "no-match-possible", 1)
	if err == nil {
		t.Fatalf("expected error from page cap, got nil (got=%v)", got)
	}
	if got != nil {
		t.Fatalf("expected nil match on cap-exceeded, got ID=%d", got.ID)
	}
	if c := int(mock.actionCallCount.Load()); c != findActionByConditionsHashPageCap {
		t.Fatalf("expected exactly %d ActionList calls (page cap), got %d", findActionByConditionsHashPageCap, c)
	}
}

// TestExistingHintForAction_Match exercises the end-to-end path:
// schema.ResourceData → ExpandSetToActionDetailsList → ConditionsHash →
// findActionByConditionsHash → HintRead by ActionID+Type → matched rule returned.
func TestExistingHintForAction_Match(t *testing.T) {
	d := newRuleResourceDataForTest(t, map[string]any{
		"type":  "iequal",
		"value": "test.example.com",
		"point": map[string]any{"header": "HOST"},
	}, 1)

	// Use the same expansion the production code will use, so hashes match.
	wantedConditions, err := resourcerule.ExpandSetToActionDetailsList(d.Get("action").(*schema.Set))
	if err != nil {
		t.Fatalf("ExpandSetToActionDetailsList: %v", err)
	}

	mock := &mockHintAPI{
		actions: []wallarm.ActionEntry{{ID: 42, Clientid: 1, Conditions: wantedConditions}},
		hints:   []wallarm.ActionBody{{ID: 100, ActionID: 42, Clientid: 1, Type: "wallarm_mode"}},
	}
	meta := &ProviderMeta{Client: mock, DefaultClientID: 1}

	actionID, rule, exists, err := existingHintForAction(d, meta, "wallarm_mode")
	if err != nil {
		t.Fatalf("existingHintForAction: %v", err)
	}
	if !exists {
		t.Fatal("expected match to exist")
	}
	if actionID != 42 {
		t.Errorf("expected actionID=42, got %d", actionID)
	}
	if rule == nil || rule.ID != 100 {
		t.Errorf("expected rule ID=100, got %+v", rule)
	}
}

// TestExistingHintForAction_NoMatch verifies the empty-tenant path: ActionList
// returns nothing → (0, nil, false, nil), no HintRead call made.
func TestExistingHintForAction_NoMatch(t *testing.T) {
	d := newRuleResourceDataForTest(t, map[string]any{
		"type":  "iequal",
		"value": "no-match.example.com",
		"point": map[string]any{"header": "HOST"},
	}, 1)

	mock := &mockHintAPI{} // no actions, no hints
	meta := &ProviderMeta{Client: mock, DefaultClientID: 1}

	actionID, rule, exists, err := existingHintForAction(d, meta, "wallarm_mode")
	if err != nil {
		t.Fatalf("existingHintForAction: %v", err)
	}
	if exists || actionID != 0 || rule != nil {
		t.Errorf("expected (0, nil, false, nil), got (%d, %+v, %v)", actionID, rule, exists)
	}
	if mock.callCount.Load() != 0 {
		t.Errorf("expected zero HintRead calls when no action matches, got %d", mock.callCount.Load())
	}
}

// TestExistingHintForAction_ActionMatchesButNoHint exercises the partial-state
// path: action exists with matching conditions, but no hint of the wanted type
// is attached. Returns (0, nil, false, nil) — the action match alone isn't enough.
func TestExistingHintForAction_ActionMatchesButNoHint(t *testing.T) {
	d := newRuleResourceDataForTest(t, map[string]any{
		"type":  "iequal",
		"value": "action-only.example.com",
		"point": map[string]any{"header": "HOST"},
	}, 1)
	wantedConditions, err := resourcerule.ExpandSetToActionDetailsList(d.Get("action").(*schema.Set))
	if err != nil {
		t.Fatalf("ExpandSetToActionDetailsList: %v", err)
	}

	mock := &mockHintAPI{
		actions: []wallarm.ActionEntry{{ID: 42, Clientid: 1, Conditions: wantedConditions}},
		// hints: empty — no rule attached to action 42
	}
	meta := &ProviderMeta{Client: mock, DefaultClientID: 1}

	actionID, rule, exists, err := existingHintForAction(d, meta, "wallarm_mode")
	if err != nil {
		t.Fatalf("existingHintForAction: %v", err)
	}
	if exists || actionID != 0 || rule != nil {
		t.Errorf("expected (0, nil, false, nil) when action matches but no hint, got (%d, %+v, %v)", actionID, rule, exists)
	}
}
