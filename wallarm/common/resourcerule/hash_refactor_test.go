package resourcerule

import (
	"encoding/json"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

// TestHashEquivalence verifies that transform-then-hash produces the same result
// as the current HashResponseActionDetails for all point types.
func TestHashEquivalence(t *testing.T) {
	tests := []struct {
		name  string
		input wallarm.ActionDetails
	}{
		{
			name:  "header HOST",
			input: wallarm.ActionDetails{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		},
		{
			name:  "instance",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"instance"}, Value: "13"},
		},
		{
			name:  "action_name",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"action_name"}, Value: "users"},
		},
		{
			name:  "action_ext",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"action_ext"}, Value: "json"},
		},
		{
			name:  "action_ext absent",
			input: wallarm.ActionDetails{Type: "absent", Point: []interface{}{"action_ext"}, Value: ""},
		},
		{
			name:  "path 0",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
		},
		{
			name:  "path 1",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"path", float64(1)}, Value: "v1"},
		},
		{
			name:  "path absent",
			input: wallarm.ActionDetails{Type: "absent", Point: []interface{}{"path", float64(2)}, Value: nil},
		},
		{
			name:  "method",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"method"}, Value: "POST"},
		},
		{
			name:  "scheme",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"scheme"}, Value: "https"},
		},
		{
			name:  "proto",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"proto"}, Value: "1.1"},
		},
		{
			name:  "query",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"get", "search"}, Value: "test"},
		},
		{
			name:  "uri",
			input: wallarm.ActionDetails{Type: "equal", Point: []interface{}{"uri"}, Value: "/very/deep/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Method 1: Current — HashResponseActionDetails (hash + transform combined)
			jsonData, _ := json.Marshal(tt.input)
			var map1 map[string]interface{}
			json.Unmarshal(jsonData, &map1)
			// Ensure "value" key exists (ActionDetailsToMap does this)
			if _, ok := map1["value"]; !ok {
				map1["value"] = ""
			}
			if map1["value"] == nil {
				map1["value"] = ""
			}
			hash1 := HashResponseActionDetails(map1)

			// Method 2: New — Transform first, then hash
			var map2 map[string]interface{}
			json.Unmarshal(jsonData, &map2)
			if _, ok := map2["value"]; !ok {
				map2["value"] = ""
			}
			if map2["value"] == nil {
				map2["value"] = ""
			}
			hash2 := HashActionDetails(map2)
			TransformAPIActionToSchema(map2)

			if hash1 != hash2 {
				t.Errorf("hash mismatch:\n  current:   %d (from HashResponseActionDetails)\n  new:       %d (from Transform+Hash)\n  map1: %v\n  map2: %v",
					hash1, hash2, map1, map2)
			} else {
				t.Logf("OK: hash=%d", hash1)
			}

			// Also verify the transformed maps are identical
			json1, _ := json.Marshal(map1)
			json2, _ := json.Marshal(map2)
			if string(json1) != string(json2) {
				t.Errorf("transformed maps differ:\n  current: %s\n  new:     %s", json1, json2)
			}
		})
	}
}
