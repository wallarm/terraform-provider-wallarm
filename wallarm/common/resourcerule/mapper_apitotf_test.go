package resourcerule

import (
	"fmt"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestThresholdToTF_Nil(t *testing.T) {
	got := ThresholdToTF(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestThresholdToTF_Values(t *testing.T) {
	got := ThresholdToTF(&wallarm.Threshold{Count: 10, Period: 60})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["count"] != 10 || m["period"] != 60 {
		t.Errorf("expected count=10 period=60, got %v", m)
	}
}

func TestReactionToTF_Nil(t *testing.T) {
	got := ReactionToTF(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestReactionToTF_Values(t *testing.T) {
	bbs := 600
	got := ReactionToTF(&wallarm.Reaction{BlockBySession: &bbs})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["block_by_session"] != 600 {
		t.Errorf("expected block_by_session=600, got %v", m["block_by_session"])
	}
}

// TestReactionToTF_OmitsAbsentKeys: when the API returns only some reaction
// keys (e.g. block_by_ip set, others absent), ReactionToTF must not seed the
// missing keys into the state map. Otherwise terraform import + -generate-
// config-out emits stray block_by_session=0 / graylist_by_ip=0 lines that the
// IntBetween(600, 315569520) validator rejects at plan time.
func TestReactionToTF_OmitsAbsentKeys(t *testing.T) {
	bbip := 3600
	got := ReactionToTF(&wallarm.Reaction{BlockByIP: &bbip})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["block_by_ip"] != 3600 {
		t.Errorf("expected block_by_ip=3600, got %v", m["block_by_ip"])
	}
	if _, ok := m["block_by_session"]; ok {
		t.Errorf("expected block_by_session absent, got %v", m["block_by_session"])
	}
	if _, ok := m["graylist_by_ip"]; ok {
		t.Errorf("expected graylist_by_ip absent, got %v", m["graylist_by_ip"])
	}
}

func TestSliceAnyToSliceString(t *testing.T) {
	got := SliceAnyToSliceString([]any{"hello", "world"})
	if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
		t.Errorf("expected [hello world], got %v", got)
	}
}

func TestSliceAnyToSliceString_Nil(t *testing.T) {
	got := SliceAnyToSliceString(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditionsToTF_Nil(t *testing.T) {
	got := AdvancedConditionsToTF(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAdvancedConditionsToTF_Values(t *testing.T) {
	got := AdvancedConditionsToTF([]wallarm.AdvancedCondition{
		{Field: "ip", Operator: "eq", Value: []string{"1.2.3.4"}},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["field"] != "ip" || m["operator"] != "eq" {
		t.Errorf("wrong fields: %v", m)
	}
}

func TestEnumeratedParametersToTF_Nil(t *testing.T) {
	if got := EnumeratedParametersToTF(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestEnumeratedParametersToTF_RegexpMode(t *testing.T) {
	plain, additional := true, false
	got := EnumeratedParametersToTF(&wallarm.EnumeratedParameters{
		Mode:                 "regexp",
		NameRegexps:          []string{"^user_"},
		ValueRegexp:          []string{"\\d+"},
		PlainParameters:      &plain,
		AdditionalParameters: &additional,
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["mode"] != "regexp" {
		t.Errorf("mode: got %v, want regexp", m["mode"])
	}
	if names, _ := m["name_regexps"].([]string); len(names) != 1 || names[0] != "^user_" {
		t.Errorf("name_regexps: got %v", m["name_regexps"])
	}
	if m["plain_parameters"] != true {
		t.Errorf("plain_parameters: got %v", m["plain_parameters"])
	}
}

func TestEnumeratedParametersToTF_ExactMode(t *testing.T) {
	got := EnumeratedParametersToTF(&wallarm.EnumeratedParameters{
		Mode: "exact",
		Points: []*wallarm.Points{
			{Point: []interface{}{"get", "password"}, Sensitive: true},
		},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["mode"] != "exact" {
		t.Errorf("mode: got %v, want exact", m["mode"])
	}
	points, ok := m["points"].([]interface{})
	if !ok || len(points) != 1 {
		t.Fatalf("points: got %v", m["points"])
	}
	p := points[0].(map[string]interface{})
	if p["sensitive"] != true {
		t.Errorf("sensitive: got %v, want true", p["sensitive"])
	}
}

func TestArbitraryConditionsToTF_Nil(t *testing.T) {
	if got := ArbitraryConditionsToTF(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestArbitraryConditionsToTF_SingleCondition(t *testing.T) {
	got := ArbitraryConditionsToTF([]wallarm.ArbitraryConditionResp{
		{Point: []interface{}{"header", "HOST"}, Operator: "eq", Value: []string{"example.com"}},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	m := got[0].(map[string]interface{})
	if m["operator"] != "eq" {
		t.Errorf("operator: got %v, want eq", m["operator"])
	}
	if values, _ := m["value"].([]string); len(values) != 1 || values[0] != "example.com" {
		t.Errorf("value: got %v", m["value"])
	}
	// Paired point ("header", "HOST") wraps to one sub-array of two elements.
	point := m["point"].([]interface{})
	if len(point) != 1 {
		t.Fatalf("expected 1 sub-array, got %d: %v", len(point), point)
	}
	inner := point[0].([]interface{})
	if len(inner) != 2 || inner[0] != "header" || inner[1] != "HOST" {
		t.Errorf("expected [[\"header\",\"HOST\"]], got %v", point)
	}
}

// TestArbitraryConditionsToTF_MultiStepPointChain is a regression test for
// the v2.3.8 round-trip bug: API returns `point` as a flat array, the
// mapper used to wrap it as a single sub-array (e.g. [["post","json_doc",
// "hash","user_id"]]), but the user's HCL expresses the same thing as a
// 2D chain ([["post"], ["json_doc"], ["hash","user_id"]]). Plan saw a
// force-replacement diff every cycle. ArbitraryConditionsToTF now uses
// WrapPointElements to chunk the flat list per the paired/simple element
// rules, matching what HCL writes.
func TestArbitraryConditionsToTF_MultiStepPointChain(t *testing.T) {
	cases := []struct {
		name string
		flat []interface{}
		want [][]string
	}{
		{
			name: "post -> json_doc -> hash:user_id",
			flat: []interface{}{"post", "json_doc", "hash", "user_id"},
			want: [][]string{{"post"}, {"json_doc"}, {"hash", "user_id"}},
		},
		{
			// XML chain: post (simple) -> xml (simple) -> xml_tag:foo (paired) -> xml_attr:bar (paired)
			// Guards the v2.3.9 follow-up: v2.3.8 only had post/json/hash coverage.
			name: "post -> xml -> xml_tag:foo -> xml_attr:bar",
			flat: []interface{}{"post", "xml", "xml_tag", "foo", "xml_attr", "bar"},
			want: [][]string{{"post"}, {"xml"}, {"xml_tag", "foo"}, {"xml_attr", "bar"}},
		},
		{
			// Protobuf chain: post -> grpc:1 (paired w/ index) -> protobuf:field (paired)
			name: "post -> grpc:1 -> protobuf:field",
			flat: []interface{}{"post", "grpc", float64(1), "protobuf", "field"},
			want: [][]string{{"post"}, {"grpc", "1"}, {"protobuf", "field"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ArbitraryConditionsToTF([]wallarm.ArbitraryConditionResp{
				{Point: tc.flat, Operator: "imatch", Value: []string{"[0-9]+"}},
			})
			point := got[0].(map[string]interface{})["point"].([]interface{})
			if len(point) != len(tc.want) {
				t.Fatalf("expected %d sub-arrays, got %d: %v", len(tc.want), len(point), point)
			}
			for i, sub := range tc.want {
				inner := point[i].([]interface{})
				if len(inner) != len(sub) {
					t.Errorf("sub-array %d: expected %d elements, got %d (%v)", i, len(sub), len(inner), inner)
					continue
				}
				for j, s := range sub {
					if fmt.Sprint(inner[j]) != s {
						t.Errorf("sub-array %d[%d]: got %v (%T), want %s", i, j, inner[j], inner[j], s)
					}
				}
			}
		})
	}
}
