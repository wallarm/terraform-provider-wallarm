package resourcerule

import (
	"strings"
	"testing"
)

func TestValidateEnumeratedParamsBlock(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		block       map[string]interface{}
		wantErrSubs []string // substrings that must all be present in the error
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
		},
		{
			name: "exact: empty everything",
			block: map[string]interface{}{
				"mode": "exact",
			},
		},
		{
			name: "exact: bool defaults to false → no error",
			block: map[string]interface{}{
				"mode":                  "exact",
				"additional_parameters": false,
				"plain_parameters":      false,
			},
		},
		// --- exact mode rejections ---
		{
			name: "exact: name_regexps populated → error",
			block: map[string]interface{}{
				"mode":         "exact",
				"name_regexps": []interface{}{"foo"},
			},
			wantErrSubs: []string{"name_regexps", "exact"},
		},
		{
			name: "exact: value_regexps populated → error",
			block: map[string]interface{}{
				"mode":          "exact",
				"value_regexps": []interface{}{"bar"},
			},
			wantErrSubs: []string{"value_regexps", "exact"},
		},
		{
			name: "exact: additional_parameters=true → error",
			block: map[string]interface{}{
				"mode":                  "exact",
				"additional_parameters": true,
			},
			wantErrSubs: []string{"additional_parameters", "exact"},
		},
		{
			name: "exact: plain_parameters=true → error",
			block: map[string]interface{}{
				"mode":             "exact",
				"plain_parameters": true,
			},
			wantErrSubs: []string{"plain_parameters", "exact"},
		},
		{
			name: "exact: multiple violations listed",
			block: map[string]interface{}{
				"mode":                  "exact",
				"name_regexps":          []interface{}{"foo"},
				"additional_parameters": true,
				"plain_parameters":      true,
			},
			wantErrSubs: []string{"name_regexps", "additional_parameters", "plain_parameters"},
		},
		// --- regexp mode happy paths ---
		{
			name: "regexp: only name_regexps",
			block: map[string]interface{}{
				"mode":         "regexp",
				"name_regexps": []interface{}{"foo"},
			},
		},
		{
			name: "regexp: full payload",
			block: map[string]interface{}{
				"mode":                  "regexp",
				"name_regexps":          []interface{}{"foo"},
				"value_regexps":         []interface{}{"bar"},
				"additional_parameters": true,
				"plain_parameters":      false,
			},
		},
		// --- regexp mode rejection ---
		{
			name: "regexp: points populated → error",
			block: map[string]interface{}{
				"mode": "regexp",
				"points": []interface{}{
					map[string]interface{}{"point": []interface{}{"header", "REFERER"}, "sensitive": false},
				},
			},
			wantErrSubs: []string{"points", "regexp"},
		},
		// --- unknown / empty mode is no-op (validation handled by schema ValidateFunc) ---
		{
			name: "unknown mode: no-op (caught elsewhere)",
			block: map[string]interface{}{
				"mode":         "weird",
				"name_regexps": []interface{}{"x"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEnumeratedParamsBlock(tc.block)
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
