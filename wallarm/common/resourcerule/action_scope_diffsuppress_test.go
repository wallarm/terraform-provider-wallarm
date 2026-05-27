package resourcerule

import (
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestSuppressIequalValueCaseDiff(t *testing.T) {
	cases := []struct {
		name    string
		setItem map[string]any
		old     string
		newVal  string
		want    bool
	}{
		{
			name:    "iequal: case-only diff suppressed",
			setItem: map[string]any{"type": "iequal", "value": "example.com"},
			old:     "example.com",
			newVal:  "Example.COM",
			want:    true,
		},
		{
			name:    "iequal: substantive diff NOT suppressed",
			setItem: map[string]any{"type": "iequal", "value": "foo.com"},
			old:     "foo.com",
			newVal:  "bar.com",
			want:    false,
		},
		{
			name:    "equal: case-only diff NOT suppressed",
			setItem: map[string]any{"type": "equal", "value": "Foo"},
			old:     "Foo",
			newVal:  "foo",
			want:    false,
		},
		{
			name:    "regex: case-only diff NOT suppressed",
			setItem: map[string]any{"type": "regex", "value": ".*"},
			old:     ".*",
			newVal:  ".+",
			want:    false,
		},
	}

	res := &schema.Resource{
		Schema: map[string]*schema.Schema{"action": ScopeActionSchema()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, res.Schema, map[string]any{
				"action": []any{tc.setItem},
			})
			set := d.Get("action").(*schema.Set)
			hash := set.F(set.List()[0])
			path := "action." + strconv.Itoa(hash) + ".value"

			got := suppressIequalValueCaseDiff(path, tc.old, tc.newVal, d)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSuppressIequalPointValueCaseDiff(t *testing.T) {
	cases := []struct {
		name     string
		setItem  map[string]any
		pointKey string
		old      string
		newVal   string
		want     bool
	}{
		{
			name:     "iequal + action_name: case-only diff suppressed",
			setItem:  map[string]any{"type": "iequal", "value": "", "point": map[string]any{"action_name": "test"}},
			pointKey: "action_name",
			old:      "test",
			newVal:   "TEST",
			want:     true,
		},
		{
			name:     "iequal + method: case-only diff suppressed",
			setItem:  map[string]any{"type": "iequal", "value": "", "point": map[string]any{"method": "get"}},
			pointKey: "method",
			old:      "get",
			newVal:   "GET",
			want:     true,
		},
		{
			name:     "iequal + instance: case-only diff suppressed",
			setItem:  map[string]any{"type": "iequal", "value": "", "point": map[string]any{"instance": "pool"}},
			pointKey: "instance",
			old:      "pool",
			newVal:   "POOL",
			want:     true,
		},
		{
			name:     "header: case-only diff suppressed (HTTP names are case-insensitive)",
			setItem:  map[string]any{"type": "iequal", "value": "example.com", "point": map[string]any{"header": "HOST"}},
			pointKey: "header",
			old:      "HOST",
			newVal:   "host",
			want:     true,
		},
		{
			name:     "header: case-only diff suppressed even with type=equal (header is always case-insensitive)",
			setItem:  map[string]any{"type": "equal", "value": "x", "point": map[string]any{"header": "X-Foo"}},
			pointKey: "header",
			old:      "X-Foo",
			newVal:   "X-FOO",
			want:     true,
		},
		{
			name:     "header: substantive diff NOT suppressed",
			setItem:  map[string]any{"type": "iequal", "value": "example.com", "point": map[string]any{"header": "HOST"}},
			pointKey: "header",
			old:      "HOST",
			newVal:   "REFERER",
			want:     false,
		},
		{
			name:     "equal + action_name: case-only diff NOT suppressed",
			setItem:  map[string]any{"type": "equal", "value": "", "point": map[string]any{"action_name": "Foo"}},
			pointKey: "action_name",
			old:      "Foo",
			newVal:   "foo",
			want:     false,
		},
		{
			name:     "iequal + action_name: substantive diff NOT suppressed",
			setItem:  map[string]any{"type": "iequal", "value": "", "point": map[string]any{"action_name": "foo"}},
			pointKey: "action_name",
			old:      "foo",
			newVal:   "bar",
			want:     false,
		},
	}

	res := &schema.Resource{
		Schema: map[string]*schema.Schema{"action": ScopeActionSchema()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, res.Schema, map[string]any{
				"action": []any{tc.setItem},
			})
			set := d.Get("action").(*schema.Set)
			hash := set.F(set.List()[0])
			path := "action." + strconv.Itoa(hash) + ".point." + tc.pointKey

			got := suppressIequalPointValueCaseDiff(path, tc.old, tc.newVal, d)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
