package tftoapi

import (
	"testing"
)

func TestThreshold_Empty(t *testing.T) {
	got, err := Threshold([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestThreshold_Values(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"count": 10, "period": 60},
	}
	got, err := Threshold(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 10 || got.Period != 60 {
		t.Errorf("expected count=10 period=60, got count=%d period=%d", got.Count, got.Period)
	}
}

func TestReaction_Empty(t *testing.T) {
	got, err := Reaction([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestReaction_Values(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"block_by_session": 600, "block_by_ip": 0, "graylist_by_ip": 0},
	}
	got, err := Reaction(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.BlockBySession == nil || *got.BlockBySession != 600 {
		t.Errorf("expected block_by_session=600, got %v", got.BlockBySession)
	}
	if got.BlockByIP != nil {
		t.Errorf("expected block_by_ip=nil (zero value), got %v", got.BlockByIP)
	}
}

func TestEnumeratedParameters_Empty(t *testing.T) {
	got, err := EnumeratedParameters([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditions_Empty(t *testing.T) {
	got, err := AdvancedConditions([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditions_Values(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"field":    "ip",
			"operator": "eq",
			"value":    []interface{}{"1.2.3.4"},
		},
	}
	got, err := AdvancedConditions(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].Field != "ip" || got[0].Operator != "eq" {
		t.Errorf("wrong fields: %+v", got[0])
	}
	if len(got[0].Value) != 1 || got[0].Value[0] != "1.2.3.4" {
		t.Errorf("wrong value: %v", got[0].Value)
	}
}
