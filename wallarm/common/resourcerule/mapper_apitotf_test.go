package resourcerule

import (
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestThresholdToTF_Nil(t *testing.T) {
	got := ThresholdToTF(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestThresholdToTF_Values(t *testing.T) {
	got := ThresholdToTF(&wallarm.Threshold{Count: 10, Period: 60})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["count"] != 10 || m["period"] != 60 {
		t.Errorf("expected count=10 period=60, got %v", m)
	}
}

func TestReactionToTF_Nil(t *testing.T) {
	got := ReactionToTF(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestReactionToTF_Values(t *testing.T) {
	bbs := 600
	got := ReactionToTF(&wallarm.Reaction{BlockBySession: &bbs})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["block_by_session"] != &bbs {
		t.Errorf("expected block_by_session=&600, got %v", m["block_by_session"])
	}
}

func TestSliceAnyToSliceString(t *testing.T) {
	got := SliceAnyToSliceString([]any{"hello", "world"})
	if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
		t.Errorf("expected [hello world], got %v", got)
	}
}

func TestSliceAnyToSliceString_Nil(t *testing.T) {
	got := SliceAnyToSliceString(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditionsToTF_Nil(t *testing.T) {
	got := AdvancedConditionsToTF(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditionsToTF_Values(t *testing.T) {
	got := AdvancedConditionsToTF([]wallarm.AdvancedCondition{
		{Field: "ip", Operator: "eq", Value: []string{"1.2.3.4"}},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["field"] != "ip" || m["operator"] != "eq" {
		t.Errorf("wrong fields: %v", m)
	}
}
