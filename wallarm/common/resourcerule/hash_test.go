package resourcerule

import (
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

// DB-verified conditions_hash values from the Wallarm API.
// These test cases ensure byte-exact compatibility with Ruby's
// Action.calculate_conditions_hash implementation.

func TestConditionsHash_EmptyConditions(t *testing.T) {
	// conditions: [], conditions_hash from DB
	got := ConditionsHash(nil)
	want := "5b8b61bd5ed79de9b3d130436a1e5a63ec663e224557ccb981bbb491a891b4dc"
	if got != want {
		t.Errorf("empty conditions:\n  got  %s\n  want %s", got, want)
	}

	// Also test with empty slice (not nil)
	got2 := ConditionsHash([]wallarm.ActionDetails{})
	if got2 != want {
		t.Errorf("empty slice:\n  got  %s\n  want %s", got2, want)
	}
}

func TestConditionsHash_SimplePath(t *testing.T) {
	// endpoint_path: "/import/mode/test"
	conditions := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "import"},
		{Type: "equal", Point: []interface{}{"path", float64(1)}, Value: "mode"},
		{Type: "absent", Point: []interface{}{"path", float64(2)}, Value: nil},
		{Type: "equal", Point: []interface{}{"action_name"}, Value: "test"},
		{Type: "absent", Point: []interface{}{"action_ext"}, Value: nil},
	}
	got := ConditionsHash(conditions)
	want := "e3a1ef0f6cf7fa1ee309ec81f63d623569b123cd88e62a7498dc9af6b733591b"
	if got != want {
		t.Errorf("simple path:\n  got  %s\n  want %s", got, want)
	}
}

func TestConditionsHash_DomainAndPath(t *testing.T) {
	// endpoint_path: "/", endpoint_domain: "example.com"
	conditions := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "absent", Point: []interface{}{"path", float64(0)}, Value: nil},
		{Type: "equal", Point: []interface{}{"action_name"}, Value: ""},
		{Type: "absent", Point: []interface{}{"action_ext"}, Value: nil},
	}
	got := ConditionsHash(conditions)
	want := "661cad9d4e1f1c8cf972668d9c01d8c0062f02920a1a1647572abd60b62755b1"
	if got != want {
		t.Errorf("domain+path:\n  got  %s\n  want %s", got, want)
	}
}

func TestConditionsHash_Complex(t *testing.T) {
	// endpoint_path: "/path/action.ext", endpoint_domain: "example.com", endpoint_instance: "13"
	// 11 conditions: method, host, path, action_name, action_ext, query, instance, proto, scheme, header regex
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
	got := ConditionsHash(conditions)
	want := "c522d1d19df25efd77ff21d973127f9515db97aa29c7373aba3f3d33fe86c42c"
	if got != want {
		t.Errorf("complex:\n  got  %s\n  want %s", got, want)
	}
}

func TestConditionsHash_OrderIndependent(t *testing.T) {
	// Same conditions in different order must produce the same hash.
	a := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
	}
	b := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
	}
	if ConditionsHash(a) != ConditionsHash(b) {
		t.Error("same conditions in different order produced different hashes")
	}
}

func TestConditionsHash_DifferentConditions(t *testing.T) {
	a := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "api"},
	}
	b := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"path", float64(0)}, Value: "admin"},
	}
	if ConditionsHash(a) == ConditionsHash(b) {
		t.Error("different conditions produced the same hash")
	}
}

func TestPointHash(t *testing.T) {
	tests := []struct {
		name  string
		point []interface{}
	}{
		{"simple", []interface{}{"header", "HOST"}},
		{"path_index", []interface{}{"path", float64(0)}},
		{"single", []interface{}{"action_name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := PointHash(tt.point)
			if len(h) != 64 {
				t.Errorf("expected 64 char hex string, got %d chars: %s", len(h), h)
			}
		})
	}

	// Same point → same hash
	h1 := PointHash([]interface{}{"header", "HOST"})
	h2 := PointHash([]interface{}{"header", "HOST"})
	if h1 != h2 {
		t.Error("same point produced different hashes")
	}

	// Different point → different hash
	h3 := PointHash([]interface{}{"header", "HOST"})
	h4 := PointHash([]interface{}{"header", "COOKIE"})
	if h3 == h4 {
		t.Error("different points produced the same hash")
	}
}

func TestRawPack(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"nil", nil, "null"},
		{"string", "hello", `"hello"`},
		{"empty_string", "", `""`},
		{"string_with_quotes", `say "hi"`, `"say \"hi\""`},
		{"int", 42, "42"},
		{"float_whole", float64(3), "3"},
		{"float_decimal", float64(1.5), "1.5"},
		{"empty_array", []interface{}{}, "[]"},
		{"string_array", []interface{}{"a", "b"}, `["a","b"]`},
		{"mixed_array", []interface{}{"path", float64(0)}, `["path",0]`},
		{"nested_array", []interface{}{[]interface{}{"a", "b"}, []interface{}{"c"}}, `[["a","b"],["c"]]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rawPack(tt.in)
			if got != tt.want {
				t.Errorf("rawPack(%#v):\n  got  %s\n  want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestSerializeCondition(t *testing.T) {
	// Verify the double-encoding: point is raw_packed first (becomes a string),
	// then the outer array raw_packs it again (quotes + escapes).
	c := wallarm.ActionDetails{
		Type:  "equal",
		Point: []interface{}{"path", float64(0)},
		Value: "import",
	}
	got := serializeCondition(c)
	// Inner: rawPack(["path",0]) = `["path",0]`
	// Outer: rawPack("["path",0]") = `"[\"path\",0]"`
	want := `[["point","[\"path\",0]"],["type","equal"],["value","import"]]`
	if got != want {
		t.Errorf("serializeCondition:\n  got  %s\n  want %s", got, want)
	}

	// Absent condition with nil value
	c2 := wallarm.ActionDetails{
		Type:  "absent",
		Point: []interface{}{"action_ext"},
		Value: nil,
	}
	got2 := serializeCondition(c2)
	want2 := `[["point","[\"action_ext\"]"],["type","absent"],["value",null]]`
	if got2 != want2 {
		t.Errorf("serializeCondition (absent):\n  got  %s\n  want %s", got2, want2)
	}

	// Empty string value (e.g., action_name = "")
	c3 := wallarm.ActionDetails{
		Type:  "equal",
		Point: []interface{}{"action_name"},
		Value: "",
	}
	got3 := serializeCondition(c3)
	want3 := `[["point","[\"action_name\"]"],["type","equal"],["value",""]]`
	if got3 != want3 {
		t.Errorf("serializeCondition (empty value):\n  got  %s\n  want %s", got3, want3)
	}
}
