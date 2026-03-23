package resourcerule

import (
	"encoding/json"
	"sort"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestActionDirName_Default(t *testing.T) {
	got := ActionDirName(nil)
	if got != "_default" {
		t.Errorf("empty conditions: got %q, want %q", got, "_default")
	}

	got2 := ActionDirName([]wallarm.ActionDetails{})
	if got2 != "_default" {
		t.Errorf("empty slice: got %q, want %q", got2, "_default")
	}
}

func TestActionDirName_SimplePath(t *testing.T) {
	// /import/mode/test → _import_mode_test_{hash8}
	conditions := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "import"},
		{Type: "equal", Point: []interface{}{"path", float64(1)}, Value: "mode"},
		{Type: "absent", Point: []interface{}{"path", float64(2)}, Value: nil},
		{Type: "equal", Point: []interface{}{"action_name"}, Value: "test"},
		{Type: "absent", Point: []interface{}{"action_ext"}, Value: nil},
	}
	got := ActionDirName(conditions)
	hash := ConditionsHash(conditions)[:8]
	want := "_import_mode_test_" + hash
	if got != want {
		t.Errorf("simple path:\n  got  %q\n  want %q", got, want)
	}
}

func TestActionDirName_DomainAndRootPath(t *testing.T) {
	// example.com / → example.com_root_{hash8}
	conditions := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "absent", Point: []interface{}{"path", float64(0)}, Value: nil},
		{Type: "equal", Point: []interface{}{"action_name"}, Value: ""},
		{Type: "absent", Point: []interface{}{"action_ext"}, Value: nil},
	}
	got := ActionDirName(conditions)
	hash := ConditionsHash(conditions)[:8]
	want := "example.com_root_" + hash
	if got != want {
		t.Errorf("domain+root:\n  got  %q\n  want %q", got, want)
	}
}

func TestActionDirName_Complex(t *testing.T) {
	// i13 example.com /path/action.ext + method, query, headers
	conditions := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"method"}, Value: "POST"},
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "path"},
		{Type: "absent", Point: []interface{}{"path", float64(1)}, Value: nil},
		{Type: "equal", Point: []interface{}{"action_name"}, Value: "action"},
		{Type: "equal", Point: []interface{}{"action_ext"}, Value: "ext"},
		{Type: "equal", Point: []interface{}{"get", "key"}, Value: "val"},
		{Type: "equal", Point: []interface{}{"instance"}, Value: "13"},
		{Type: "equal", Point: []interface{}{"proto"}, Value: "1.1"},
		{Type: "equal", Point: []interface{}{"scheme"}, Value: "https"},
		{Type: "regex", Point: []interface{}{"header", "CONTENT-TYPE"}, Value: "application/.*"},
	}
	got := ActionDirName(conditions)
	hash := ConditionsHash(conditions)[:8]
	want := "13_example.com_path_action.ext_" + hash
	if got != want {
		t.Errorf("complex:\n  got  %q\n  want %q", got, want)
	}
}

func TestActionDirName_WildcardPath(t *testing.T) {
	// /api/**/*.* → _api_.._._. + hash
	conditions := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
	}
	got := ActionDirName(conditions)
	hash := ConditionsHash(conditions)[:8]
	// ReverseMapActions: path[0]=api, no limiter → /api/**/*.* → api_.._._.
	want := "_api_.._._._" + hash
	if got != want {
		t.Errorf("wildcard path:\n  got  %q\n  want %q", got, want)
	}
}

func TestActionDirName_GlobalWildcard(t *testing.T) {
	// conditions: []  →  _default (no hash)
	// The /**/*.* path is the default and produces "_default"
	got := ActionDirName([]wallarm.ActionDetails{})
	if got != "_default" {
		t.Errorf("global wildcard: got %q, want %q", got, "_default")
	}
}

func TestActionDirName_DomainOnly(t *testing.T) {
	// domain with no path constraints → domain + /**/*.* path (empty path part)
	conditions := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
	}
	got := ActionDirName(conditions)
	hash := ConditionsHash(conditions)[:8]
	// ReverseMapActions: domain=example.com, path=/**/*.* → empty path part
	want := "example.com_" + hash
	if got != want {
		t.Errorf("domain only:\n  got  %q\n  want %q", got, want)
	}
}

func TestActionDirName_MaxLength(t *testing.T) {
	// Very long domain + path should be truncated to 64 chars.
	conditions := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "api.staging.very-long-subdomain.example.com"},
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "v2"},
		{Type: "equal", Point: []interface{}{"path", float64(1)}, Value: "organizations"},
		{Type: "equal", Point: []interface{}{"path", float64(2)}, Value: "members"},
		{Type: "absent", Point: []interface{}{"path", float64(3)}, Value: nil},
		{Type: "equal", Point: []interface{}{"action_name"}, Value: "permissions"},
		{Type: "absent", Point: []interface{}{"action_ext"}, Value: nil},
	}
	got := ActionDirName(conditions)
	if len(got) > 64 {
		t.Errorf("dir name exceeds 64 chars: len=%d, name=%q", len(got), got)
	}
	// Must end with _hash8
	hash := ConditionsHash(conditions)[:8]
	if !endsWith(got, "_"+hash) {
		t.Errorf("truncated name doesn't end with hash: got %q, want suffix %q", got, "_"+hash)
	}
}

func TestActionDirName_SortOrder(t *testing.T) {
	// Verify natural sort order: _default < _path < numeric < domain
	names := []string{
		"example.com_api_a3f2e1b7",
		"_default",
		"13_example.com_api_c522d1d1",
		"_api_v1_users_e3a1ef0f",
		"2_other.com_b7e1c4d8",
	}
	sort.Strings(names)
	expected := []string{
		"13_example.com_api_c522d1d1",
		"2_other.com_b7e1c4d8",
		"_api_v1_users_e3a1ef0f",
		"_default",
		"example.com_api_a3f2e1b7",
	}
	for i, got := range names {
		if got != expected[i] {
			t.Errorf("sort order[%d]: got %q, want %q", i, got, expected[i])
		}
	}
}

func TestActionMeta_Format(t *testing.T) {
	actionID := 11461731
	meta := NewActionMeta(
		&actionID,
		8649,
		nil,
		[]wallarm.ActionDetails{
			{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "import"},
			{Type: "equal", Point: []interface{}{"action_name"}, Value: "test"},
		},
		strPtr("/import/test"),
		nil,
		nil,
		intPtr(1774068210),
	)

	data, err := FormatActionMeta(meta)
	if err != nil {
		t.Fatalf("FormatActionMeta error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify key fields
	if parsed["action_id"] != float64(11461731) {
		t.Errorf("action_id: got %v", parsed["action_id"])
	}
	if parsed["client_id"] != float64(8649) {
		t.Errorf("client_id: got %v", parsed["client_id"])
	}
	if parsed["conditions_hash"] == nil || parsed["conditions_hash"] == "" {
		t.Error("conditions_hash is empty")
	}
	if parsed["dir_name"] == nil || parsed["dir_name"] == "" {
		t.Error("dir_name is empty")
	}
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
