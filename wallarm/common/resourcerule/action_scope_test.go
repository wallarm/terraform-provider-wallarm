package resourcerule

import (
	"testing"
)

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
		if !pointValuePoints[p] {
			t.Errorf("expected %q in pointValuePoints", p)
		}
	}

	// Header and query use the value field for matched content.
	valueField := []string{"header", "query", "path"}
	for _, p := range valueField {
		if pointValuePoints[p] {
			t.Errorf("%q should not be in pointValuePoints (value goes in the value field, not point)", p)
		}
	}
}
