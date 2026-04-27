package resourcerule

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

// Each customizer reads one schema field and writes the corresponding pointer
// on HintUpdateV3Params. This test verifies the field-name → params-field
// round-trip for every simple With* helper. Nested-block customizers
// (WithThreshold/Reaction/EnumeratedParameters) and WithValues have their
// own dedicated tests below.

func TestUpdateCustomizers_Simple(t *testing.T) {
	type stringCase struct {
		name      string
		field     string
		value     string
		customize UpdateCustomizer
		extract   func(*wallarm.HintUpdateV3Params) *string
	}
	stringCases := []stringCase{
		{"WithMode", "mode", "block", WithMode, func(p *wallarm.HintUpdateV3Params) *string { return p.Mode }},
		{"WithAttackType", "attack_type", "sqli", WithAttackType, func(p *wallarm.HintUpdateV3Params) *string { return p.AttackType }},
		{"WithRegex", "regex", ".*", WithRegex, func(p *wallarm.HintUpdateV3Params) *string { return p.Regex }},
		{"WithLoginRegex", "login_regex", ".+", WithLoginRegex, func(p *wallarm.HintUpdateV3Params) *string { return p.LoginRegex }},
		{"WithCredStuffType", "cred_stuff_type", "custom", WithCredStuffType, func(p *wallarm.HintUpdateV3Params) *string { return p.CredStuffType }},
		{"WithParser", "parser", "json_doc", WithParser, func(p *wallarm.HintUpdateV3Params) *string { return p.Parser }},
		{"WithState", "state", "enabled", WithState, func(p *wallarm.HintUpdateV3Params) *string { return p.State }},
		{"WithTimeUnit", "time_unit", "rps", WithTimeUnit, func(p *wallarm.HintUpdateV3Params) *string { return p.TimeUnit }},
		{"WithName", "name", "X-Foo", WithName, func(p *wallarm.HintUpdateV3Params) *string { return p.Name }},
		{"WithFileType", "file_type", "images", WithFileType, func(p *wallarm.HintUpdateV3Params) *string { return p.FileType }},
	}
	for _, tc := range stringCases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				tc.field: {Type: schema.TypeString, Optional: true},
			}, map[string]interface{}{tc.field: tc.value})
			p := &wallarm.HintUpdateV3Params{}
			if err := tc.customize(d, p); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := tc.extract(p)
			if got == nil || *got != tc.value {
				t.Errorf("field %q: got %v, want %q", tc.field, got, tc.value)
			}
		})
	}

	type intCase struct {
		name      string
		field     string
		value     int
		customize UpdateCustomizer
		extract   func(*wallarm.HintUpdateV3Params) *int
	}
	intCases := []intCase{
		{"WithStamp", "stamp", 5678, WithStamp, func(p *wallarm.HintUpdateV3Params) *int { return p.Stamp }},
		{"WithSize", "size", 200, WithSize, func(p *wallarm.HintUpdateV3Params) *int { return p.Size }},
		{"WithMaxDepth", "max_depth", 20, WithMaxDepth, func(p *wallarm.HintUpdateV3Params) *int { return p.MaxDepth }},
		{"WithMaxValueSizeKb", "max_value_size_kb", 99, WithMaxValueSizeKb, func(p *wallarm.HintUpdateV3Params) *int { return p.MaxValueSizeKb }},
		{"WithMaxDocSizeKb", "max_doc_size_kb", 200, WithMaxDocSizeKb, func(p *wallarm.HintUpdateV3Params) *int { return p.MaxDocSizeKb }},
		{"WithMaxDocPerBatch", "max_doc_per_batch", 25, WithMaxDocPerBatch, func(p *wallarm.HintUpdateV3Params) *int { return p.MaxDocPerBatch }},
		{"WithOverlimitTime", "overlimit_time", 5000, WithOverlimitTime, func(p *wallarm.HintUpdateV3Params) *int { return p.OverlimitTime }},
		{"WithDelay", "delay", 250, WithDelay, func(p *wallarm.HintUpdateV3Params) *int { return p.Delay }},
		{"WithBurst", "burst", 99, WithBurst, func(p *wallarm.HintUpdateV3Params) *int { return p.Burst }},
		{"WithRate", "rate", 700, WithRate, func(p *wallarm.HintUpdateV3Params) *int { return p.Rate }},
		{"WithRspStatus", "rsp_status", 503, WithRspStatus, func(p *wallarm.HintUpdateV3Params) *int { return p.RspStatus }},
	}
	for _, tc := range intCases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				tc.field: {Type: schema.TypeInt, Optional: true},
			}, map[string]interface{}{tc.field: tc.value})
			p := &wallarm.HintUpdateV3Params{}
			if err := tc.customize(d, p); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := tc.extract(p)
			if got == nil || *got != tc.value {
				t.Errorf("field %q: got %v, want %d", tc.field, got, tc.value)
			}
		})
	}

	type boolCase struct {
		name      string
		field     string
		value     bool
		customize UpdateCustomizer
		extract   func(*wallarm.HintUpdateV3Params) *bool
	}
	boolCases := []boolCase{
		{"WithCaseSensitive", "case_sensitive", true, WithCaseSensitive, func(p *wallarm.HintUpdateV3Params) *bool { return p.CaseSensitive }},
		{"WithIntrospection", "introspection", true, WithIntrospection, func(p *wallarm.HintUpdateV3Params) *bool { return p.Introspection }},
		{"WithDebugEnabled", "debug_enabled", true, WithDebugEnabled, func(p *wallarm.HintUpdateV3Params) *bool { return p.DebugEnabled }},
	}
	for _, tc := range boolCases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				tc.field: {Type: schema.TypeBool, Optional: true},
			}, map[string]interface{}{tc.field: tc.value})
			p := &wallarm.HintUpdateV3Params{}
			if err := tc.customize(d, p); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := tc.extract(p)
			if got == nil || *got != tc.value {
				t.Errorf("field %q: got %v, want %v", tc.field, got, tc.value)
			}
		})
	}
}

// WithValues handles both TypeSet (set_response_header.values) and TypeList,
// converting via .List() / direct slice respectively. The TypeSet path
// surfaced as a panic in v2.3.7 import smoke; this test pins both shapes.
func TestWithValues_TypeSet(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"values": {Type: schema.TypeSet, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}, map[string]interface{}{"values": []interface{}{"a", "b"}})
	p := &wallarm.HintUpdateV3Params{}
	if err := WithValues(d, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Values == nil || len(*p.Values) != 2 {
		t.Fatalf("expected 2 values, got %v", p.Values)
	}
}

func TestWithValues_TypeList(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"values": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}, map[string]interface{}{"values": []interface{}{"a", "b", "c"}})
	p := &wallarm.HintUpdateV3Params{}
	if err := WithValues(d, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Values == nil || len(*p.Values) != 3 {
		t.Fatalf("expected 3 values, got %v", p.Values)
	}
}

// Nested-block customizers must propagate errors from the underlying *ToAPI
// helpers (don't silently swallow — Update path otherwise reports success
// while server-side fields stay unchanged).

func TestWithThreshold_Success(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"threshold": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"count":  {Type: schema.TypeInt, Optional: true},
				"period": {Type: schema.TypeInt, Optional: true},
			},
		}},
	}, map[string]interface{}{
		"threshold": []interface{}{map[string]interface{}{"count": 99, "period": 60}},
	})
	p := &wallarm.HintUpdateV3Params{}
	if err := WithThreshold(d, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Threshold == nil || p.Threshold.Count != 99 || p.Threshold.Period != 60 {
		t.Errorf("got %+v", p.Threshold)
	}
}

func TestWithReaction_Success(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"reaction": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"block_by_session": {Type: schema.TypeInt, Optional: true},
				"block_by_ip":      {Type: schema.TypeInt, Optional: true},
				"graylist_by_ip":   {Type: schema.TypeInt, Optional: true},
			},
		}},
	}, map[string]interface{}{
		"reaction": []interface{}{map[string]interface{}{"block_by_session": 3000, "block_by_ip": 4000}},
	})
	p := &wallarm.HintUpdateV3Params{}
	if err := WithReaction(d, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Reaction == nil || p.Reaction.BlockBySession == nil || *p.Reaction.BlockBySession != 3000 ||
		p.Reaction.BlockByIP == nil || *p.Reaction.BlockByIP != 4000 {
		t.Errorf("got %+v", p.Reaction)
	}
}
