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

func TestEnumeratedParametersToTF_Nil(t *testing.T) {
	if got := EnumeratedParametersToTF(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestEnumeratedParametersToTF_RegexpMode(t *testing.T) {
	plain, additional := true, false
	got := EnumeratedParametersToTF(&wallarm.EnumeratedParameters{
		Mode:                 "regexp",
		NameRegexps:          []string{"^user_"},
		ValueRegexp:          []string{"\\d+"},
		PlainParameters:      &plain,
		AdditionalParameters: &additional,
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["mode"] != "regexp" {
		t.Errorf("mode: got %v, want regexp", m["mode"])
	}
	if names, _ := m["name_regexps"].([]string); len(names) != 1 || names[0] != "^user_" {
		t.Errorf("name_regexps: got %v", m["name_regexps"])
	}
	if m["plain_parameters"] != true {
		t.Errorf("plain_parameters: got %v", m["plain_parameters"])
	}
}

func TestEnumeratedParametersToTF_ExactMode(t *testing.T) {
	got := EnumeratedParametersToTF(&wallarm.EnumeratedParameters{
		Mode: "exact",
		Points: []*wallarm.Points{
			{Point: []interface{}{"get", "password"}, Sensitive: true},
		},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["mode"] != "exact" {
		t.Errorf("mode: got %v, want exact", m["mode"])
	}
	points, ok := m["points"].([]interface{})
	if !ok || len(points) != 1 {
		t.Fatalf("points: got %v", m["points"])
	}
	p := points[0].(map[string]interface{})
	if p["sensitive"] != true {
		t.Errorf("sensitive: got %v, want true", p["sensitive"])
	}
}

func TestArbitraryConditionsToTF_Nil(t *testing.T) {
	if got := ArbitraryConditionsToTF(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestArbitraryConditionsToTF_SingleCondition(t *testing.T) {
	got := ArbitraryConditionsToTF([]wallarm.ArbitraryConditionResp{
		{Point: []interface{}{"header", "HOST"}, Operator: "eq", Value: []string{"example.com"}},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["operator"] != "eq" {
		t.Errorf("operator: got %v, want eq", m["operator"])
	}
	if values, _ := m["value"].([]string); len(values) != 1 || values[0] != "example.com" {
		t.Errorf("value: got %v", m["value"])
	}
}
