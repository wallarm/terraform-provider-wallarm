package wallarm

import (
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestFlattenActionConditions(t *testing.T) {
	conditions := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
		{Type: "absent", Point: []interface{}{"path", float64(1)}, Value: nil},
		{Type: "equal", Point: []interface{}{"instance"}, Value: "13"},
		{Type: "equal", Point: []interface{}{"action_name"}, Value: "login"},
	}

	result := flattenActionConditions(conditions)

	if len(result) != 5 {
		t.Fatalf("expected 5 conditions, got %d", len(result))
	}

	// Header: point flattened as strings.
	r0 := result[0]
	if r0["type"] != "iequal" {
		t.Errorf("[0] type: got %v, want iequal", r0["type"])
	}
	if r0["value"] != "example.com" {
		t.Errorf("[0] value: got %v, want example.com", r0["value"])
	}
	point0, ok := r0["point"].([]string)
	if !ok || len(point0) != 2 || point0[0] != "header" || point0[1] != "HOST" {
		t.Errorf("[0] point: got %v, want [header HOST]", r0["point"])
	}

	// Path with float64 index → string.
	r1 := result[1]
	point1, ok := r1["point"].([]string)
	if !ok || len(point1) != 2 || point1[1] != "0" {
		t.Errorf("[1] point: got %v, want [path 0]", r1["point"])
	}

	// Absent path — nil value → empty string.
	r2 := result[2]
	if r2["type"] != "absent" {
		t.Errorf("[2] type: got %v, want absent", r2["type"])
	}
	if r2["value"] != "" {
		t.Errorf("[2] value: got %q, want empty string", r2["value"])
	}

	// Instance — single-element point.
	r3 := result[3]
	point3, ok := r3["point"].([]string)
	if !ok || len(point3) != 1 || point3[0] != "instance" {
		t.Errorf("[3] point: got %v, want [instance]", r3["point"])
	}
	if r3["value"] != "13" {
		t.Errorf("[3] value: got %v, want 13", r3["value"])
	}

	// Action name — single-element point.
	r4 := result[4]
	point4, ok := r4["point"].([]string)
	if !ok || len(point4) != 1 || point4[0] != "action_name" {
		t.Errorf("[4] point: got %v, want [action_name]", r4["point"])
	}
}

func TestFlattenActionConditions_Empty(t *testing.T) {
	result := flattenActionConditions(nil)
	if len(result) != 0 {
		t.Errorf("expected 0 results for nil input, got %d", len(result))
	}
}

func TestFlattenActionConditions_RoundTrip(t *testing.T) {
	// Flatten then convert back via schemaActionToDetails should preserve semantics.
	original := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
	}

	flat := flattenActionConditions(original)

	// Convert flat format to schema format (as buildActionFromHit would produce).
	schemaActions := make([]map[string]interface{}, 0, len(flat))
	for _, f := range flat {
		point := f["point"].([]string)
		pointMap := make(map[string]interface{})
		if len(point) >= 2 {
			pointMap[point[0]] = point[1]
		} else if len(point) == 1 {
			pointMap[point[0]] = f["value"]
		}
		schemaActions = append(schemaActions, map[string]interface{}{
			"type":  f["type"],
			"value": f["value"],
			"point": pointMap,
		})
	}

	// Convert back to ActionDetails.
	roundTripped := schemaActionToDetails(schemaActions)

	if len(roundTripped) != len(original) {
		t.Fatalf("round-trip length: got %d, want %d", len(roundTripped), len(original))
	}
	for i, rt := range roundTripped {
		if rt.Type != original[i].Type {
			t.Errorf("[%d] type: got %q, want %q", i, rt.Type, original[i].Type)
		}
	}
}
