package resourcerule

import (
	"strings"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestValidateActionSet_Valid(t *testing.T) {
	set := newActionSet(
		map[string]interface{}{"type": "iequal", "value": "example.com", "point": map[string]interface{}{"header": "HOST"}},
		map[string]interface{}{"type": "equal", "value": "api", "point": map[string]interface{}{"path": "0"}},
	)
	if err := validateActionSet(set); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidateActionSet_Nil(t *testing.T) {
	if err := validateActionSet(nil); err != nil {
		t.Errorf("expected nil for nil set, got %v", err)
	}
}

func TestValidateActionSet_UnknownPointKey(t *testing.T) {
	set := newActionSet(map[string]interface{}{
		"type": "equal", "value": "x", "point": map[string]interface{}{"headers": "HOST"},
	})
	err := validateActionSet(set)
	if err == nil || !strings.Contains(err.Error(), "unknown action point key") {
		t.Errorf("expected unknown-key error, got %v", err)
	}
}

func TestValidateActionSet_MultipleKeys(t *testing.T) {
	set := newActionSet(map[string]interface{}{
		"type": "equal", "value": "x", "point": map[string]interface{}{"header": "HOST", "path": "0"},
	})
	err := validateActionSet(set)
	if err == nil || !strings.Contains(err.Error(), "exactly one key") {
		t.Errorf("expected multi-key error, got %v", err)
	}
}

func TestValidateActionSet_URIConflict(t *testing.T) {
	// uri is a point-value type: the value lives in the point map.
	set := newActionSet(
		map[string]interface{}{"type": "equal", "value": "", "point": map[string]interface{}{"uri": "/api/v1"}},
		map[string]interface{}{"type": "equal", "value": "api", "point": map[string]interface{}{"path": "0"}},
	)
	err := validateActionSet(set)
	if err == nil || !strings.Contains(err.Error(), "conflicts with") {
		t.Errorf("expected uri-conflict error, got %v", err)
	}
}

func TestValidateActionSet_PointValueWithNonEmptyValue(t *testing.T) {
	// action_name is a PointValuePoint — value must live in the point map, not the value field.
	set := newActionSet(map[string]interface{}{
		"type": "equal", "value": "login", "point": map[string]interface{}{"action_name": "something"},
	})
	err := validateActionSet(set)
	if err == nil || !strings.Contains(err.Error(), "value goes in the point map") {
		t.Errorf("expected point-value error, got %v", err)
	}
}

func TestValidateActionSet_HeaderEmptyValue(t *testing.T) {
	set := newActionSet(map[string]interface{}{
		"type": "equal", "value": "", "point": map[string]interface{}{"header": "HOST"},
	})
	err := validateActionSet(set)
	if err == nil || !strings.Contains(err.Error(), "non-empty") {
		t.Errorf("expected empty-value error, got %v", err)
	}
}

func TestValidateActionSet_AbsentSkipsValueChecks(t *testing.T) {
	// type=absent allows both point-value-with-value AND empty-value for header.
	set := newActionSet(map[string]interface{}{
		"type": "absent", "value": "", "point": map[string]interface{}{"header": "HOST"},
	})
	if err := validateActionSet(set); err != nil {
		t.Errorf("absent should bypass value checks, got %v", err)
	}
}

func TestValidPointKeys(t *testing.T) {
	expected := []string{
		"header", "method", "path", "action_name", "action_ext",
		"query", "proto", "scheme", "uri", "instance",
	}
	for _, k := range expected {
		if !validPointKeys[k] {
			t.Errorf("expected %q in validPointKeys", k)
		}
	}

	// Typos and invalid keys should not be valid.
	invalid := []string{
		"headers", "pth", "action_path", "get", "host",
		"url", "ext", "name", "protocol",
	}
	for _, k := range invalid {
		if validPointKeys[k] {
			t.Errorf("%q should not be in validPointKeys", k)
		}
	}
}

func TestURIConflictPoints(t *testing.T) {
	conflicting := []string{"path", "action_name", "action_ext", "query"}
	for _, p := range conflicting {
		if !uriConflictPoints[p] {
			t.Errorf("expected %q in uriConflictPoints", p)
		}
	}

	// These should NOT conflict with uri.
	nonConflicting := []string{"header", "method", "scheme", "proto", "instance", "uri"}
	for _, p := range nonConflicting {
		if uriConflictPoints[p] {
			t.Errorf("%q should not be in uriConflictPoints", p)
		}
	}
}

func TestPointValuePoints(t *testing.T) {
	// Points where value goes in the point map, not the value field.
	pointValue := []string{"action_name", "action_ext", "method", "proto", "scheme", "uri", "instance"}
	for _, p := range pointValue {
		if !PointValuePoints[p] {
			t.Errorf("expected %q in PointValuePoints", p)
		}
	}

	// Header and query use the value field for matched content.
	valueField := []string{"header", "query", "path"}
	for _, p := range valueField {
		if PointValuePoints[p] {
			t.Errorf("%q should not be in PointValuePoints (value goes in the value field, not point)", p)
		}
	}
}

func TestActionDetailToSchemaItem(t *testing.T) {
	tests := []struct {
		name      string
		input     wallarm.ActionDetails
		wantType  string
		wantValue string
		wantPoint map[string]interface{}
	}{
		{
			name:      "header",
			input:     wallarm.ActionDetails{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
			wantType:  "iequal",
			wantValue: "example.com",
			wantPoint: map[string]interface{}{"header": "HOST"},
		},
		{
			name:      "instance — type preserved, value moves to point",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"instance"}, Value: "13"},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]interface{}{"instance": "13"},
		},
		{
			name:      "action_name — value moves to point",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"action_name"}, Value: "login"},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]interface{}{"action_name": "login"},
		},
		{
			name:      "action_ext — value moves to point",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"action_ext"}, Value: "json"},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]interface{}{"action_ext": "json"},
		},
		{
			name:      "path with integer index",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
			wantType:  "equal",
			wantValue: "api",
			wantPoint: map[string]interface{}{"path": "0"},
		},
		{
			name:      "absent path",
			input:     wallarm.ActionDetails{Type: "absent", Point: []interface{}{"path", float64(1)}, Value: nil},
			wantType:  "absent",
			wantValue: "",
			wantPoint: map[string]interface{}{"path": "1"},
		},
		{
			name:      "query (get → query mapping)",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"get", "search"}, Value: "test"},
			wantType:  "equal",
			wantValue: "test",
			wantPoint: map[string]interface{}{"query": "search"},
		},
		{
			name:      "method — value moves to point",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"method"}, Value: "GET"},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]interface{}{"method": "GET"},
		},
		{
			name:      "scheme — value moves to point",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"scheme"}, Value: "https"},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]interface{}{"scheme": "https"},
		},
		{
			name:      "proto — value moves to point",
			input:     wallarm.ActionDetails{Type: "equal", Point: []interface{}{"proto"}, Value: "1.1"},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]interface{}{"proto": "1.1"},
		},
		{
			name:      "nil value → empty string",
			input:     wallarm.ActionDetails{Type: "absent", Point: []interface{}{"action_ext"}, Value: nil},
			wantType:  "absent",
			wantValue: "",
			wantPoint: map[string]interface{}{"action_ext": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ActionDetailToSchemaItem(tt.input)

			gotType := result["type"].(string)
			if gotType != tt.wantType {
				t.Errorf("type: got %q, want %q", gotType, tt.wantType)
			}

			gotValue := result["value"].(string)
			if gotValue != tt.wantValue {
				t.Errorf("value: got %q, want %q", gotValue, tt.wantValue)
			}

			gotPoint := result["point"].(map[string]interface{})
			if len(gotPoint) != len(tt.wantPoint) {
				t.Errorf("point length: got %d, want %d", len(gotPoint), len(tt.wantPoint))
			}
			for k, wantV := range tt.wantPoint {
				gotV, ok := gotPoint[k]
				if !ok {
					t.Errorf("point missing key %q", k)
				} else if gotV != wantV {
					t.Errorf("point[%q]: got %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}
