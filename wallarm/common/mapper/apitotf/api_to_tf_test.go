package apitotf

import (
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestThreshold_Nil(t *testing.T) {
	got := Threshold(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestThreshold_Values(t *testing.T) {
	got := Threshold(&wallarm.Threshold{Count: 10, Period: 60})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["count"] != 10 || m["period"] != 60 {
		t.Errorf("expected count=10 period=60, got %v", m)
	}
}

func TestReaction_Nil(t *testing.T) {
	got := Reaction(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestReaction_Values(t *testing.T) {
	bbs := 600
	got := Reaction(&wallarm.Reaction{BlockBySession: &bbs})
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

func TestAdvancedConditions_Nil(t *testing.T) {
	got := AdvancedConditions(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditions_Values(t *testing.T) {
	got := AdvancedConditions([]wallarm.AdvancedCondition{
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
