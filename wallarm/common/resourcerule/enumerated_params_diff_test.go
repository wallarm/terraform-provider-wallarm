package resourcerule

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

// buildEPRawCfg constructs a cty.Value of the same shape that
// d.GetRawConfig() returns for a resource with an `enumerated_parameters`
// list block: a top-level object with a list-of-objects attribute. Only the
// attributes referenced by the validator (`additional_parameters`,
// `plain_parameters`) are populated; pass cty.NullVal for "user did not set
// this field in HCL".
func buildEPRawCfg(addParam, plainParam cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"enumerated_parameters": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"additional_parameters": addParam,
				"plain_parameters":      plainParam,
			}),
		}),
	})
}

func TestValidateEnumeratedParamsBlock(t *testing.T) {
	t.Parallel()

	// Convenience constructors for the rawCfg fixtures.
	noBoolsSet := buildEPRawCfg(cty.NullVal(cty.Bool), cty.NullVal(cty.Bool))
	additionalFalseSet := buildEPRawCfg(cty.False, cty.NullVal(cty.Bool))
	additionalTrueSet := buildEPRawCfg(cty.True, cty.NullVal(cty.Bool))
	plainFalseSet := buildEPRawCfg(cty.NullVal(cty.Bool), cty.False)
	bothSet := buildEPRawCfg(cty.True, cty.False)

	cases := []struct {
		name        string
		block       map[string]interface{}
		rawCfg      cty.Value
		wantErrSubs []string
	}{
		// --- exact mode happy paths ---
		{
			name: "exact: only points",
			block: map[string]interface{}{
				"mode": "exact",
				"points": []interface{}{
					map[string]interface{}{"point": []interface{}{"header", "REFERER"}, "sensitive": true},
				},
			},
			rawCfg: noBoolsSet,
		},
		{
			name: "exact: empty everything",
			block: map[string]interface{}{
				"mode": "exact",
			},
			rawCfg: noBoolsSet,
		},
		{
			name: "exact: bool default false applied (user did not write field) → no error",
			block: map[string]interface{}{
				"mode":                  "exact",
				"additional_parameters": false,
				"plain_parameters":      false,
			},
			// Both cty.NullVal — user omitted from HCL; SDK Default filled the d.Get value.
			rawCfg: noBoolsSet,
		},
		// --- exact mode list rejections ---
		{
			name: "exact: name_regexps populated → error",
			block: map[string]interface{}{
				"mode":         "exact",
				"name_regexps": []interface{}{"foo"},
			},
			rawCfg:      noBoolsSet,
			wantErrSubs: []string{"name_regexps", "exact"},
		},
		{
			name: "exact: value_regexps populated → error",
			block: map[string]interface{}{
				"mode":          "exact",
				"value_regexps": []interface{}{"bar"},
			},
			rawCfg:      noBoolsSet,
			wantErrSubs: []string{"value_regexps", "exact"},
		},
		// --- exact mode strict bool denial (the v2.3.8 tightening) ---
		{
			name: "exact: additional_parameters=true in HCL → error",
			block: map[string]interface{}{
				"mode":                  "exact",
				"additional_parameters": true,
			},
			rawCfg:      additionalTrueSet,
			wantErrSubs: []string{"additional_parameters", "exact"},
		},
		{
			name: "exact: additional_parameters=false in HCL → error (strict)",
			block: map[string]interface{}{
				"mode":                  "exact",
				"additional_parameters": false,
			},
			rawCfg:      additionalFalseSet,
			wantErrSubs: []string{"additional_parameters", "exact"},
		},
		{
			name: "exact: plain_parameters=false in HCL → error (strict)",
			block: map[string]interface{}{
				"mode":             "exact",
				"plain_parameters": false,
			},
			rawCfg:      plainFalseSet,
			wantErrSubs: []string{"plain_parameters", "exact"},
		},
		{
			name: "exact: both bools written + name_regexps → all listed",
			block: map[string]interface{}{
				"mode":                  "exact",
				"name_regexps":          []interface{}{"foo"},
				"additional_parameters": true,
				"plain_parameters":      false,
			},
			rawCfg:      bothSet,
			wantErrSubs: []string{"name_regexps", "additional_parameters", "plain_parameters"},
		},
		// --- regexp mode happy paths ---
		{
			name: "regexp: name_regexps + value_regexps populated → ok",
			block: map[string]interface{}{
				"mode":          "regexp",
				"name_regexps":  []interface{}{"foo"},
				"value_regexps": []interface{}{"bar"},
			},
			rawCfg: noBoolsSet,
		},
		{
			name: "regexp: full payload (user opted out one filter with [\"\"])",
			block: map[string]interface{}{
				"mode":                  "regexp",
				"name_regexps":          []interface{}{"foo"},
				"value_regexps":         []interface{}{""},
				"additional_parameters": true,
				"plain_parameters":      false,
			},
			rawCfg: bothSet,
		},
		// --- regexp mode rejections ---
		{
			name: "regexp: points populated → error",
			block: map[string]interface{}{
				"mode":          "regexp",
				"name_regexps":  []interface{}{"foo"},
				"value_regexps": []interface{}{"bar"},
				"points": []interface{}{
					map[string]interface{}{"point": []interface{}{"header", "REFERER"}, "sensitive": false},
				},
			},
			rawCfg:      noBoolsSet,
			wantErrSubs: []string{"points", "regexp"},
		},
		{
			name: "regexp: value_regexps missing → error",
			block: map[string]interface{}{
				"mode":         "regexp",
				"name_regexps": []interface{}{"foo"},
			},
			rawCfg:      noBoolsSet,
			wantErrSubs: []string{"value_regexps", "regexp"},
		},
		{
			name: "regexp: name_regexps missing → error",
			block: map[string]interface{}{
				"mode":          "regexp",
				"value_regexps": []interface{}{"bar"},
			},
			rawCfg:      noBoolsSet,
			wantErrSubs: []string{"name_regexps", "regexp"},
		},
		{
			name: "regexp: both lists missing → both listed in error",
			block: map[string]interface{}{
				"mode": "regexp",
			},
			rawCfg:      noBoolsSet,
			wantErrSubs: []string{"name_regexps", "value_regexps", "regexp"},
		},
		// --- unknown mode is no-op (validation handled by schema ValidateFunc) ---
		{
			name: "unknown mode: no-op (caught elsewhere)",
			block: map[string]interface{}{
				"mode":         "weird",
				"name_regexps": []interface{}{"x"},
			},
			rawCfg: noBoolsSet,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEnumeratedParamsBlock(tc.block, tc.rawCfg)
			if len(tc.wantErrSubs) == 0 {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %v, got nil", tc.wantErrSubs)
			}
			for _, sub := range tc.wantErrSubs {
				if !strings.Contains(err.Error(), sub) {
					t.Errorf("expected error to contain %q, got %q", sub, err.Error())
				}
			}
		})
	}
}
