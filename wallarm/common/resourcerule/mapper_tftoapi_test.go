package resourcerule

import (
	"testing"
)

func TestThresholdToAPI_Empty(t *testing.T) {
	got, err := ThresholdToAPI([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestThresholdToAPI_Values(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"count": 10, "period": 60},
	}
	got, err := ThresholdToAPI(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 10 || got.Period != 60 {
		t.Errorf("expected count=10 period=60, got count=%d period=%d", got.Count, got.Period)
	}
}

func TestReactionToAPI_Empty(t *testing.T) {
	got, err := ReactionToAPI([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestReactionToAPI_Values(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"block_by_session": 600, "block_by_ip": 0, "graylist_by_ip": 0},
	}
	got, err := ReactionToAPI(input)
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

func TestEnumeratedParametersToAPI_Empty(t *testing.T) {
	got, err := EnumeratedParametersToAPI([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditionsToAPI_Empty(t *testing.T) {
	got, err := AdvancedConditionsToAPI([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditionsToAPI_Values(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"field":    "ip",
			"operator": "eq",
			"value":    []interface{}{"1.2.3.4"},
		},
	}
	got, err := AdvancedConditionsToAPI(input)
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

func TestEnumeratedParametersToAPI_RegexpMode(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"mode":                  "regexp",
			"name_regexps":          []interface{}{"^user_"},
			"value_regexps":         []interface{}{"\\d+"},
			"plain_parameters":      true,
			"additional_parameters": false,
		},
	}
	got, err := EnumeratedParametersToAPI(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Mode != "regexp" {
		t.Errorf("mode: got %q, want regexp", got.Mode)
	}
	if len(got.NameRegexps) != 1 || got.NameRegexps[0] != "^user_" {
		t.Errorf("name_regexps: got %v", got.NameRegexps)
	}
	if got.PlainParameters == nil || !*got.PlainParameters {
		t.Errorf("plain_parameters: got %v", got.PlainParameters)
	}
	if got.AdditionalParameters == nil || *got.AdditionalParameters {
		t.Errorf("additional_parameters: got %v", got.AdditionalParameters)
	}
}

func TestEnumeratedParametersToAPI_RegexpModeDefaultsEmpty(t *testing.T) {
	// Empty name/value regexps lists → helper injects [""] so the API receives a list.
	input := []interface{}{
		map[string]interface{}{
			"mode":          "regexp",
			"name_regexps":  []interface{}{},
			"value_regexps": []interface{}{},
		},
	}
	got, err := EnumeratedParametersToAPI(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.NameRegexps) != 1 || got.NameRegexps[0] != "" {
		t.Errorf("expected NameRegexps=[\"\"], got %v", got.NameRegexps)
	}
	if len(got.ValueRegexp) != 1 || got.ValueRegexp[0] != "" {
		t.Errorf("expected ValueRegexp=[\"\"], got %v", got.ValueRegexp)
	}
}

func TestEnumeratedParametersToAPI_ExactMode(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"mode": "exact",
			"points": []interface{}{
				map[string]interface{}{
					"point":     []interface{}{"get", "password"},
					"sensitive": true,
				},
			},
		},
	}
	got, err := EnumeratedParametersToAPI(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Mode != "exact" {
		t.Errorf("mode: got %q, want exact", got.Mode)
	}
	if len(got.Points) != 1 {
		t.Fatalf("expected 1 points, got %d", len(got.Points))
	}
	if !got.Points[0].Sensitive {
		t.Errorf("expected sensitive=true, got false")
	}
}

func TestEnumeratedParametersToAPI_ExactModeEmptyPoints(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"mode": "exact", "points": []interface{}{}},
	}
	got, err := EnumeratedParametersToAPI(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Mode != "exact" {
		t.Errorf("mode: got %q, want exact", got.Mode)
	}
	if got.Points != nil {
		t.Errorf("expected nil points for empty input, got %v", got.Points)
	}
}

func TestArbitraryConditionsToAPI_Empty(t *testing.T) {
	got, err := ArbitraryConditionsToAPI([]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestArbitraryConditionsToAPI_SingleCondition(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"point": []interface{}{
				[]interface{}{"header", "HOST"},
			},
			"operator": "eq",
			"value":    []interface{}{"example.com"},
		},
	}
	got, err := ArbitraryConditionsToAPI(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Operator != "eq" {
		t.Errorf("operator: got %q, want eq", got[0].Operator)
	}
	if len(got[0].Value) != 1 || got[0].Value[0] != "example.com" {
		t.Errorf("value: got %v", got[0].Value)
	}
	if len(got[0].Point) != 1 || len(got[0].Point[0]) != 2 {
		t.Errorf("point shape: got %v", got[0].Point)
	}
}
