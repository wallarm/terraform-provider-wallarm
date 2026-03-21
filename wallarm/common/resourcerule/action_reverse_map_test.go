package resourcerule

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestReverseMapActions(t *testing.T) {
	tests := []struct {
		name       string
		actions    []wallarm.ActionDetails
		wantPath   string
		wantDomain string
		wantInst   string
		wantMethod string
		wantScheme string
	}{
		{
			name: "simple /api/v1/users",
			actions: []wallarm.ActionDetails{
				{Type: "iequal", Value: "example.com", Point: []interface{}{"header", "HOST"}},
				{Type: "equal", Value: "users", Point: []interface{}{"action_name"}},
				{Type: "absent", Point: []interface{}{"action_ext"}},
				{Type: "equal", Value: "api", Point: []interface{}{"path", float64(0)}},
				{Type: "equal", Value: "v1", Point: []interface{}{"path", float64(1)}},
				{Type: "absent", Point: []interface{}{"path", float64(2)}},
			},
			wantPath: "/api/v1/users", wantDomain: "example.com",
		},
		{
			name: "with extension /api/v1/data.json",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "data", Point: []interface{}{"action_name"}},
				{Type: "equal", Value: "json", Point: []interface{}{"action_ext"}},
				{Type: "equal", Value: "api", Point: []interface{}{"path", float64(0)}},
				{Type: "equal", Value: "v1", Point: []interface{}{"path", float64(1)}},
				{Type: "absent", Point: []interface{}{"path", float64(2)}},
			},
			wantPath: "/api/v1/data.json",
		},
		{
			name: "root /",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "", Point: []interface{}{"action_name"}},
				{Type: "absent", Point: []interface{}{"action_ext"}},
				{Type: "absent", Point: []interface{}{"path", float64(0)}},
			},
			wantPath: "/",
		},
		{
			name: "wildcard /api/*/users (gap in path indices)",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "users", Point: []interface{}{"action_name"}},
				{Type: "absent", Point: []interface{}{"action_ext"}},
				{Type: "equal", Value: "api", Point: []interface{}{"path", float64(0)}},
				{Type: "absent", Point: []interface{}{"path", float64(2)}},
			},
			wantPath: "/api/*/users",
		},
		{
			name: "globstar /api/**/admin (no limiter)",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "admin", Point: []interface{}{"action_name"}},
				{Type: "absent", Point: []interface{}{"action_ext"}},
				{Type: "equal", Value: "api", Point: []interface{}{"path", float64(0)}},
			},
			wantPath: "/api/**/admin",
		},
		{
			name: "wildcard action_name /articles/*",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "articles", Point: []interface{}{"path", float64(0)}},
				{Type: "absent", Point: []interface{}{"path", float64(1)}},
				{Type: "absent", Point: []interface{}{"action_ext"}},
			},
			wantPath: "/articles/*",
		},
		{
			name: "wildcard extension /api/rfp/jobs.*",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "jobs", Point: []interface{}{"action_name"}},
				{Type: "equal", Value: "api", Point: []interface{}{"path", float64(0)}},
				{Type: "equal", Value: "rfp", Point: []interface{}{"path", float64(1)}},
				{Type: "absent", Point: []interface{}{"path", float64(2)}},
			},
			wantPath: "/api/rfp/jobs.*",
		},
		{
			name: "no path conditions -> /**/*.*",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "20", Point: []interface{}{"instance"}},
			},
			wantPath: "/**/*.*", wantInst: "20",
		},
		{
			name:     "empty conditions -> /**/*.*",
			actions:  []wallarm.ActionDetails{},
			wantPath: "/**/*.*",
		},
		{
			name: "with method and scheme",
			actions: []wallarm.ActionDetails{
				{Type: "iequal", Value: "example.com", Point: []interface{}{"header", "HOST"}},
				{Type: "equal", Value: "users", Point: []interface{}{"action_name"}},
				{Type: "absent", Point: []interface{}{"action_ext"}},
				{Type: "absent", Point: []interface{}{"path", float64(0)}},
				{Type: "equal", Value: "POST", Point: []interface{}{"method"}},
				{Type: "equal", Value: "https", Point: []interface{}{"scheme"}},
			},
			wantPath: "/users", wantDomain: "example.com", wantMethod: "POST", wantScheme: "https",
		},
		{
			name: "with instance",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "1", Point: []interface{}{"instance"}},
				{Type: "equal", Value: "endpoint", Point: []interface{}{"action_name"}},
				{Type: "absent", Point: []interface{}{"action_ext"}},
				{Type: "equal", Value: "api", Point: []interface{}{"path", float64(0)}},
				{Type: "absent", Point: []interface{}{"path", float64(1)}},
			},
			wantPath: "/api/endpoint", wantInst: "1",
		},
		{
			name: "URI fallback",
			actions: []wallarm.ActionDetails{
				{Type: "equal", Value: "/very/deep/path", Point: []interface{}{"uri"}},
			},
			wantPath: "/very/deep/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReverseMapActions(tt.actions)
			if got.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", got.Path, tt.wantPath)
			}
			if got.Domain != tt.wantDomain {
				t.Errorf("Domain = %q, want %q", got.Domain, tt.wantDomain)
			}
			if tt.wantInst != "" && got.Instance != tt.wantInst {
				t.Errorf("Instance = %q, want %q", got.Instance, tt.wantInst)
			}
			if tt.wantMethod != "" && got.Method != tt.wantMethod {
				t.Errorf("Method = %q, want %q", got.Method, tt.wantMethod)
			}
			if tt.wantScheme != "" && got.Scheme != tt.wantScheme {
				t.Errorf("Scheme = %q, want %q", got.Scheme, tt.wantScheme)
			}
		})
	}
}

// TestReverseMapRealExamples validates against 343 real API examples.
func TestReverseMapRealExamples(t *testing.T) {
	type example struct {
		Conditions []wallarm.ActionDetails `json:"conditions"`
		Path       string                  `json:"path"`
		Domain     string                  `json:"domain"`
		Instance   interface{}             `json:"instance"`
		Method     string                  `json:"method"`
		Scheme     string                  `json:"scheme"`
		Proto      string                  `json:"proto"`
	}

	data, err := os.ReadFile("../../../.claude/actions_examples.json")
	if err != nil {
		t.Skipf("Skipping real examples test: %v", err)
	}

	var examples []example
	if err := json.Unmarshal(data, &examples); err != nil {
		t.Fatalf("Failed to parse examples: %v", err)
	}

	t.Logf("Testing %d real API examples", len(examples))

	for i, ex := range examples {
		got := ReverseMapActions(ex.Conditions)

		if got.Path != ex.Path {
			t.Errorf("Example #%d: Path = %q, want %q", i, got.Path, ex.Path)
		}
		if got.Domain != ex.Domain {
			t.Errorf("Example #%d: Domain = %q, want %q", i, got.Domain, ex.Domain)
		}
		if ex.Instance != nil {
			wantInst := fmt.Sprintf("%v", ex.Instance)
			if got.Instance != wantInst {
				t.Errorf("Example #%d: Instance = %q, want %q", i, got.Instance, wantInst)
			}
		}
		if ex.Method != "" && got.Method != ex.Method {
			t.Errorf("Example #%d: Method = %q, want %q", i, got.Method, ex.Method)
		}
	}
}

// TestExpandPathToActions tests the forward mapping.
func TestExpandPathToActions(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantKeys []string // expected point keys in order
	}{
		{
			name:     "simple /api/v1/users",
			path:     "/api/v1/users",
			wantKeys: []string{"action_name", "action_ext", "path", "path", "path"},
		},
		{
			name:     "with extension /api/data.json",
			path:     "/api/data.json",
			wantKeys: []string{"action_name", "action_ext", "path", "path"},
		},
		{
			name:     "root /",
			path:     "/",
			wantKeys: []string{"action_name", "action_ext", "path"},
		},
		{
			name:     "wildcard /api/*/users",
			path:     "/api/*/users",
			wantKeys: []string{"action_name", "action_ext", "path", "path"}, // path[1] skipped for *
		},
		{
			name:     "globstar /api/**/admin",
			path:     "/api/**/admin",
			wantKeys: []string{"action_name", "action_ext", "path"}, // no limiter
		},
		{
			name:     "wildcard action /articles/*",
			path:     "/articles/*",
			wantKeys: []string{"action_ext", "path", "path"}, // no action_name, limiter
		},
		{
			name:     "wildcard ext /api/jobs.*",
			path:     "/api/jobs.*",
			wantKeys: []string{"action_name", "path", "path"}, // no action_ext
		},
		{
			name:     "global /**/*.*",
			path:     "/**/*.*",
			wantKeys: nil, // no conditions
		},
		{
			name:     "empty path",
			path:     "",
			wantKeys: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if tt.wantKeys == nil {
				if len(got) != 0 {
					t.Errorf("Expected no conditions, got %d", len(got))
				}
				return
			}
			if len(got) != len(tt.wantKeys) {
				t.Errorf("Got %d conditions, want %d", len(got), len(tt.wantKeys))
				for i, a := range got {
					t.Logf("  [%d] %s %v = %v", i, ActionPointKey(a), a.Point, a.Value)
				}
				return
			}
			for i, wantKey := range tt.wantKeys {
				if ActionPointKey(got[i]) != wantKey {
					t.Errorf("Condition[%d] key = %q, want %q", i, ActionPointKey(got[i]), wantKey)
				}
			}
		})
	}
}

// TestRoundTrip validates that forward -> reverse produces the original path.
func TestRoundTrip(t *testing.T) {
	paths := []string{
		"/",
		"/users",
		"/api/v1/users",
		"/api/v1/data.json",
		"/api/*/users",
		"/api/**/admin",
		"/articles/*",
		"/api/jobs.*",
		"/api/v1/*/data",
		"/w1/WebService/Admin/getDBLogBySQL",
		"/**/*.*",
		// Note: empty "" is equivalent to "/**/*.*" -- both produce zero conditions.
		// Round-trip: "" -> 0 conditions -> "/**/*.*". This is expected.
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			// Forward: path -> conditions
			actions := ExpandPathToActions(path, "", "", "", "", "", nil, nil)

			// Reverse: conditions -> path
			got := ReverseMapActions(actions)

			if got.Path != path {
				t.Errorf("Round-trip failed: %q -> %d conditions -> %q", path, len(actions), got.Path)
				for i, a := range actions {
					t.Logf("  [%d] type=%q key=%s point=%v value=%v", i, a.Type, ActionPointKey(a), a.Point, a.Value)
				}
			}
		})
	}
}

// TestRoundTripWithScope validates round-trip with domain/instance/method.
func TestRoundTripWithScope(t *testing.T) {
	actions := ExpandPathToActions(
		"/api/v1/users",
		"example.com",
		"17",
		"POST",
		"https",
		"1.1",
		[]QueryParam{{Key: "key", Value: "value", Type: "equal"}},
		[]HeaderParam{{Name: "X-Custom", Value: "test", Type: "equal"}},
	)

	got := ReverseMapActions(actions)

	if got.Path != "/api/v1/users" {
		t.Errorf("Path = %q, want /api/v1/users", got.Path)
	}
	if got.Domain != "example.com" {
		t.Errorf("Domain = %q, want example.com", got.Domain)
	}
	if got.Instance != "17" {
		t.Errorf("Instance = %q, want 17", got.Instance)
	}
	if got.Method != "POST" {
		t.Errorf("Method = %q, want POST", got.Method)
	}
	if got.Scheme != "https" {
		t.Errorf("Scheme = %q, want https", got.Scheme)
	}
	if got.Proto != "1.1" {
		t.Errorf("Proto = %q, want 1.1", got.Proto)
	}
	if len(got.Query) != 1 || got.Query[0].Key != "key" {
		t.Errorf("Query = %+v, want [{key value equal}]", got.Query)
	}
	if len(got.Headers) != 1 || got.Headers[0].Name != "X-CUSTOM" {
		t.Errorf("Headers = %+v, want [{X-CUSTOM test equal}]", got.Headers)
	}
}

// TestRealExamplesRoundTrip validates that for each real example,
// reverse(conditions) -> path -> forward(path) -> reverse again produces the same path.
func TestRealExamplesRoundTrip(t *testing.T) {
	type example struct {
		Conditions []wallarm.ActionDetails `json:"conditions"`
		Path       string                  `json:"path"`
		Domain     string                  `json:"domain"`
		Instance   interface{}             `json:"instance"`
		Method     string                  `json:"method"`
		Scheme     string                  `json:"scheme"`
		Proto      string                  `json:"proto"`
		Query      []QueryParam            `json:"query"`
		Headers    []HeaderParam           `json:"headers"`
	}

	data, err := os.ReadFile("../../../.claude/actions_examples.json")
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}

	var examples []example
	if err := json.Unmarshal(data, &examples); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	for i, ex := range examples {
		// Step 1: Reverse-map API conditions -> path
		rev := ReverseMapActions(ex.Conditions)
		if rev.Path != ex.Path {
			t.Errorf("#%d reverse: got %q, want %q", i, rev.Path, ex.Path)
			continue
		}

		// Step 2: Forward-map path -> conditions
		inst := ""
		if ex.Instance != nil {
			inst = fmt.Sprintf("%v", ex.Instance)
		}
		actions2 := ExpandPathToActions(rev.Path, rev.Domain, inst, rev.Method, rev.Scheme, rev.Proto, ex.Query, ex.Headers)

		// Step 3: Reverse again -> should match original path
		rev2 := ReverseMapActions(actions2)
		if rev2.Path != ex.Path {
			t.Errorf("#%d round-trip: %q -> expand -> %q (want %q)", i, ex.Path, rev2.Path, ex.Path)
		}
	}
}
