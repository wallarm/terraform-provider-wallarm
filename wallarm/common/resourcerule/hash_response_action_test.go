package resourcerule

import (
	"testing"
)

// TestHashResponseActionDetails_Baseline captures the current hash values and
// side effects of HashResponseActionDetails for every point type. This is the
// regression safety net — if any refactoring changes a hash or side effect,
// this test catches it.
func TestHashResponseActionDetails_Baseline(t *testing.T) {
	tests := []struct {
		name string
		// input is the API-format map (as produced by ActionDetailsToMap)
		input map[string]interface{}
		// wantHash is the expected hash value
		wantHash int
		// After hashing, the map is mutated. These are the expected values.
		wantType  string
		wantValue string
		wantPoint interface{} // map[string]string after transform
	}{
		{
			name: "header",
			input: map[string]interface{}{
				"type":  "iequal",
				"value": "example.com",
				"point": []interface{}{"header", "HOST"},
			},
			wantHash:  HashString("iequal-example.com-map[header:HOST]-"),
			wantType:  "iequal",
			wantValue: "example.com",
			wantPoint: map[string]string{"header": "HOST"},
		},
		{
			name: "instance",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "9",
				"point": []interface{}{"instance"},
			},
			wantHash:  HashString("equal-9-map[instance:9]-"),
			wantType:  "",
			wantValue: "",
			wantPoint: map[string]string{"instance": "9"},
		},
		{
			name: "path",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "api",
				"point": []interface{}{"path", float64(0)},
			},
			wantHash:  HashString("equal-api-map[path:0]-"),
			wantType:  "equal",
			wantValue: "api",
			wantPoint: map[string]string{"path": "0"},
		},
		{
			name: "action_name",
			input: map[string]interface{}{
				"type":  "iequal",
				"value": "login",
				"point": []interface{}{"action_name"},
			},
			wantHash:  HashString("iequal-login-map[action_name:login]-"),
			wantType:  "iequal",
			wantValue: "",
			wantPoint: map[string]string{"action_name": "login"},
		},
		{
			name: "action_ext",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "php",
				"point": []interface{}{"action_ext"},
			},
			wantHash:  HashString("equal-php-map[action_ext:php]-"),
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"action_ext": "php"},
		},
		{
			name: "method",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "GET",
				"point": []interface{}{"method"},
			},
			wantHash:  HashString("equal-GET-map[method:GET]-"),
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"method": "GET"},
		},
		{
			name: "scheme",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "https",
				"point": []interface{}{"scheme"},
			},
			wantHash:  HashString("equal-https-map[scheme:https]-"),
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"scheme": "https"},
		},
		{
			name: "proto",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "1.1",
				"point": []interface{}{"proto"},
			},
			wantHash:  HashString("equal-1.1-map[proto:1.1]-"),
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"proto": "1.1"},
		},
		{
			name: "uri",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "/api/v1",
				"point": []interface{}{"uri"},
			},
			wantHash:  HashString("equal-/api/v1-map[uri:/api/v1]-"),
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"uri": "/api/v1"},
		},
		{
			name: "query (get → query)",
			input: map[string]interface{}{
				"type":  "equal",
				"value": "test",
				"point": []interface{}{"get", "search"},
			},
			wantHash:  HashString("equal-test-map[query:search]-"),
			wantType:  "equal",
			wantValue: "test",
			wantPoint: map[string]string{"query": "search"},
		},
		{
			name: "absent action_ext",
			input: map[string]interface{}{
				"type":  "absent",
				"value": "",
				"point": []interface{}{"action_ext"},
			},
			wantHash:  HashString("absent--map[action_ext:]-"),
			wantType:  "absent",
			wantValue: "",
			wantPoint: map[string]string{"action_ext": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashResponseActionDetails(tt.input)
			if got != tt.wantHash {
				t.Errorf("hash: got %d, want %d", got, tt.wantHash)
			}

			// Verify side effects on the map.
			if tt.input["type"].(string) != tt.wantType {
				t.Errorf("type after hash: got %q, want %q", tt.input["type"], tt.wantType)
			}
			if tt.input["value"].(string) != tt.wantValue {
				t.Errorf("value after hash: got %q, want %q", tt.input["value"], tt.wantValue)
			}
			gotPoint := tt.input["point"].(map[string]string)
			wantPoint := tt.wantPoint.(map[string]string)
			if len(gotPoint) != len(wantPoint) {
				t.Errorf("point length: got %d, want %d", len(gotPoint), len(wantPoint))
			}
			for k, v := range wantPoint {
				if gotPoint[k] != v {
					t.Errorf("point[%q]: got %q, want %q", k, gotPoint[k], v)
				}
			}
		})
	}
}

// TestTransformAPIActionToSchema verifies that TransformAPIActionToSchema produces
// the same mutations as HashResponseActionDetails side effects for every point type.
func TestTransformAPIActionToSchema(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantType  string
		wantValue string
		wantPoint map[string]string
	}{
		{
			name:      "header",
			input:     map[string]interface{}{"type": "iequal", "value": "example.com", "point": []interface{}{"header", "HOST"}},
			wantType:  "iequal",
			wantValue: "example.com",
			wantPoint: map[string]string{"header": "HOST"},
		},
		{
			name:      "instance — type preserved, value moved to point",
			input:     map[string]interface{}{"type": "equal", "value": "9", "point": []interface{}{"instance"}},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"instance": "9"},
		},
		{
			name:      "path",
			input:     map[string]interface{}{"type": "equal", "value": "api", "point": []interface{}{"path", float64(0)}},
			wantType:  "equal",
			wantValue: "api",
			wantPoint: map[string]string{"path": "0"},
		},
		{
			name:      "action_name",
			input:     map[string]interface{}{"type": "iequal", "value": "login", "point": []interface{}{"action_name"}},
			wantType:  "iequal",
			wantValue: "",
			wantPoint: map[string]string{"action_name": "login"},
		},
		{
			name:      "action_ext",
			input:     map[string]interface{}{"type": "equal", "value": "php", "point": []interface{}{"action_ext"}},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"action_ext": "php"},
		},
		{
			name:      "method",
			input:     map[string]interface{}{"type": "equal", "value": "GET", "point": []interface{}{"method"}},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"method": "GET"},
		},
		{
			name:      "scheme",
			input:     map[string]interface{}{"type": "equal", "value": "https", "point": []interface{}{"scheme"}},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"scheme": "https"},
		},
		{
			name:      "proto",
			input:     map[string]interface{}{"type": "equal", "value": "1.1", "point": []interface{}{"proto"}},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"proto": "1.1"},
		},
		{
			name:      "uri",
			input:     map[string]interface{}{"type": "equal", "value": "/api/v1", "point": []interface{}{"uri"}},
			wantType:  "equal",
			wantValue: "",
			wantPoint: map[string]string{"uri": "/api/v1"},
		},
		{
			name:      "query (get → query)",
			input:     map[string]interface{}{"type": "equal", "value": "test", "point": []interface{}{"get", "search"}},
			wantType:  "equal",
			wantValue: "test",
			wantPoint: map[string]string{"query": "search"},
		},
		{
			name:      "absent action_ext",
			input:     map[string]interface{}{"type": "absent", "value": "", "point": []interface{}{"action_ext"}},
			wantType:  "absent",
			wantValue: "",
			wantPoint: map[string]string{"action_ext": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TransformAPIActionToSchema(tt.input)

			if tt.input["type"].(string) != tt.wantType {
				t.Errorf("type: got %q, want %q", tt.input["type"], tt.wantType)
			}
			if tt.input["value"].(string) != tt.wantValue {
				t.Errorf("value: got %q, want %q", tt.input["value"], tt.wantValue)
			}
			gotPoint := tt.input["point"].(map[string]string)
			if len(gotPoint) != len(tt.wantPoint) {
				t.Errorf("point length: got %d, want %d", len(gotPoint), len(tt.wantPoint))
			}
			for k, v := range tt.wantPoint {
				if gotPoint[k] != v {
					t.Errorf("point[%q]: got %q, want %q", k, gotPoint[k], v)
				}
			}
		})
	}
}

// TestTransformAPIActionToSchema_AlreadyTransformed verifies that calling
// TransformAPIActionToSchema on already-transformed data is a no-op.
func TestTransformAPIActionToSchema_AlreadyTransformed(t *testing.T) {
	input := map[string]interface{}{
		"type":  "",
		"value": "",
		"point": map[string]interface{}{"instance": "9"},
	}

	// Should not panic or modify.
	TransformAPIActionToSchema(input)

	if input["type"].(string) != "" {
		t.Errorf("type should remain empty, got %q", input["type"])
	}
	if input["value"].(string) != "" {
		t.Errorf("value should remain empty, got %q", input["value"])
	}
}

// TestHashActionDetails_ConsistentAcrossFormats verifies that HashActionDetails
// produces the same hash for the same logical condition regardless of whether
// the input is in API format ([]interface{}) or config format (map).
func TestHashActionDetails_ConsistentAcrossFormats(t *testing.T) {
	tests := []struct {
		name      string
		apiFormat map[string]interface{}
		cfgFormat map[string]interface{}
	}{
		{
			name: "header",
			apiFormat: map[string]interface{}{
				"type": "iequal", "value": "example.com",
				"point": []interface{}{"header", "HOST"},
			},
			cfgFormat: map[string]interface{}{
				"type": "iequal", "value": "example.com",
				"point": map[string]interface{}{"header": "HOST"},
			},
		},
		{
			name: "instance — type equal vs empty both hash the same",
			apiFormat: map[string]interface{}{
				"type": "equal", "value": "9",
				"point": []interface{}{"instance"},
			},
			cfgFormat: map[string]interface{}{
				"type": "", "value": "9",
				"point": map[string]interface{}{"instance": "9"},
			},
		},
		{
			name: "instance — type equal in both formats",
			apiFormat: map[string]interface{}{
				"type": "equal", "value": "9",
				"point": []interface{}{"instance"},
			},
			cfgFormat: map[string]interface{}{
				"type": "equal", "value": "9",
				"point": map[string]interface{}{"instance": "9"},
			},
		},
		{
			name: "path",
			apiFormat: map[string]interface{}{
				"type": "equal", "value": "api",
				"point": []interface{}{"path", float64(0)},
			},
			cfgFormat: map[string]interface{}{
				"type": "equal", "value": "api",
				"point": map[string]interface{}{"path": "0"},
			},
		},
		{
			name: "action_name",
			apiFormat: map[string]interface{}{
				"type": "iequal", "value": "login",
				"point": []interface{}{"action_name"},
			},
			cfgFormat: map[string]interface{}{
				"type": "iequal", "value": "login",
				"point": map[string]interface{}{"action_name": "login"},
			},
		},
		{
			name: "query (get → query)",
			apiFormat: map[string]interface{}{
				"type": "equal", "value": "test",
				"point": []interface{}{"get", "search"},
			},
			cfgFormat: map[string]interface{}{
				"type": "equal", "value": "test",
				"point": map[string]interface{}{"query": "search"},
			},
		},
		{
			name: "method",
			apiFormat: map[string]interface{}{
				"type": "equal", "value": "GET",
				"point": []interface{}{"method"},
			},
			cfgFormat: map[string]interface{}{
				"type": "equal", "value": "GET",
				"point": map[string]interface{}{"method": "GET"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiHash := HashActionDetails(tt.apiFormat)
			cfgHash := HashActionDetails(tt.cfgFormat)
			if apiHash != cfgHash {
				t.Errorf("hash mismatch: API format=%d, config format=%d", apiHash, cfgHash)
			}
		})
	}
}

// TestHashActionDetails_InstanceTypeChange verifies that changing instance type
// from "equal" to "regex" produces a different hash, triggering ForceNew.
func TestHashActionDetails_InstanceTypeChange(t *testing.T) {
	equalInstance := map[string]interface{}{
		"type": "equal", "value": "9",
		"point": map[string]interface{}{"instance": "9"},
	}
	emptyInstance := map[string]interface{}{
		"type": "", "value": "9",
		"point": map[string]interface{}{"instance": "9"},
	}
	regexInstance := map[string]interface{}{
		"type": "regex", "value": "9",
		"point": map[string]interface{}{"instance": "9"},
	}

	equalHash := HashActionDetails(equalInstance)
	emptyHash := HashActionDetails(emptyInstance)
	regexHash := HashActionDetails(regexInstance)

	// "" and "equal" should hash the same (both are the default).
	if equalHash != emptyHash {
		t.Errorf("equal and empty should hash the same: equal=%d, empty=%d", equalHash, emptyHash)
	}

	// "regex" should hash differently (actual type change).
	if regexHash == equalHash {
		t.Errorf("regex should hash differently from equal: both=%d", regexHash)
	}
}

// TestHashActionDetails_NoSideEffects verifies HashActionDetails does not
// mutate the input map.
func TestHashActionDetails_NoSideEffects(t *testing.T) {
	input := map[string]interface{}{
		"type":  "equal",
		"value": "9",
		"point": []interface{}{"instance"},
	}

	HashActionDetails(input)

	// Verify no mutations.
	if input["type"].(string) != "equal" {
		t.Errorf("type was mutated to %q", input["type"])
	}
	if input["value"].(string) != "9" {
		t.Errorf("value was mutated to %q", input["value"])
	}
	p := input["point"].([]interface{})
	if len(p) != 1 || p[0].(string) != "instance" {
		t.Errorf("point was mutated to %v", input["point"])
	}
}

// TestHashActionDetails_IequalValueCaseInsensitive: iequal values share a
// hash regardless of case because the API downcases them server-side.
func TestHashActionDetails_IequalValueCaseInsensitive(t *testing.T) {
	mixed := map[string]interface{}{
		"type":  "iequal",
		"value": "Example.COM",
		"point": map[string]interface{}{"header": "HOST"},
	}
	lower := map[string]interface{}{
		"type":  "iequal",
		"value": "example.com",
		"point": map[string]interface{}{"header": "HOST"},
	}
	if HashActionDetails(mixed) != HashActionDetails(lower) {
		t.Errorf("iequal: expected case-insensitive hashes\n  mixed=%d\n  lower=%d",
			HashActionDetails(mixed), HashActionDetails(lower))
	}
}

// TestHashActionDetails_IequalValueBearingPointCaseInsensitive: for point
// types where the matched value lives inside the point map (action_name,
// method, instance, etc.), iequal-typed conditions hash equal regardless
// of case in the point-map value.
func TestHashActionDetails_IequalValueBearingPointCaseInsensitive(t *testing.T) {
	for _, key := range []string{"action_name", "action_ext", "method", "instance", "scheme", "uri", "proto"} {
		mixed := map[string]interface{}{
			"type":  "iequal",
			"value": "",
			"point": map[string]interface{}{key: "TEST"},
		}
		lower := map[string]interface{}{
			"type":  "iequal",
			"value": "",
			"point": map[string]interface{}{key: "test"},
		}
		if HashActionDetails(mixed) != HashActionDetails(lower) {
			t.Errorf("%s: expected case-insensitive hashes for iequal\n  mixed=%d\n  lower=%d",
				key, HashActionDetails(mixed), HashActionDetails(lower))
		}
	}
}

// TestLowercaseValueBearingPointEntries verifies the map[string]string
// variant (the iface variant is exercised via TestHashActionDetails_*).
func TestLowercaseValueBearingPointEntries(t *testing.T) {
	in := map[string]string{
		"action_name": "TEST",
		"method":      "GET",
		"header":      "HOST", // not value-bearing — stays as-is
	}
	out := lowercaseValueBearingPointEntries(in)
	if out["action_name"] != "test" {
		t.Errorf("action_name: got %q, want test", out["action_name"])
	}
	if out["method"] != "get" {
		t.Errorf("method: got %q, want get", out["method"])
	}
	if out["header"] != "HOST" {
		t.Errorf("header (not value-bearing): got %q, want HOST", out["header"])
	}
}

// TestHashActionDetails_NonIequalValueCaseSensitive: equal and regex stay
// case-sensitive so a real value change still produces a different hash.
func TestHashActionDetails_NonIequalValueCaseSensitive(t *testing.T) {
	for _, condType := range []string{"equal", "regex"} {
		mixed := map[string]interface{}{
			"type":  condType,
			"value": "Example.COM",
			"point": map[string]interface{}{"header": "HOST"},
		}
		lower := map[string]interface{}{
			"type":  condType,
			"value": "example.com",
			"point": map[string]interface{}{"header": "HOST"},
		}
		if HashActionDetails(mixed) == HashActionDetails(lower) {
			t.Errorf("%s: expected case-sensitive hashes, got equal", condType)
		}
	}
}

// TestHashResponseActionDetails_ConfigFormatPanics documents that the current
// HashResponseActionDetails panics when given config-format data (point as
// map[string]interface{} instead of []interface{}). This is the limitation
// we're fixing: it can only handle API-format data.
func TestHashResponseActionDetails_ConfigFormatPanics(t *testing.T) {
	input := map[string]interface{}{
		"type":  "",
		"value": "",
		"point": map[string]interface{}{"instance": "9"},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on config-format input, but got none")
		}
	}()

	HashResponseActionDetails(input)
}
