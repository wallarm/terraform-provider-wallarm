package resourcerule

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// EnumeratedParamsCustomizeDiff fails plan when the user has populated fields
// in `enumerated_parameters` that the Wallarm API ignores for the chosen
// `mode`. The TF→API mapper silently drops these fields on PUT, which
// produces a perpetual plan diff: state is read back without them, the next
// plan re-emits the same change, the next apply drops them again.
//
// `mode = "exact"`  → only `points` applies.
// `mode = "regexp"` → only `name_regexps`, `value_regexps`,
//
//	`additional_parameters`, `plain_parameters` apply.
//
// Booleans are checked only when `true`. `Default: false` makes the unset
// case indistinguishable from an explicit false; a false bool is harmless
// (the mapper would drop it anyway), so erroring on false would break valid
// configs.
func EnumeratedParamsCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	raw, ok := d.Get("enumerated_parameters").([]interface{})
	if !ok || len(raw) == 0 {
		return nil
	}
	block, ok := raw[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return validateEnumeratedParamsBlock(block)
}

// validateEnumeratedParamsBlock is the pure-data half of
// EnumeratedParamsCustomizeDiff — separated so it can be unit-tested without
// constructing a *schema.ResourceDiff.
func validateEnumeratedParamsBlock(block map[string]interface{}) error {
	mode, _ := block["mode"].(string)

	switch mode {
	case modeExact:
		var bad []string
		if v, _ := block["name_regexps"].([]interface{}); len(v) > 0 {
			bad = append(bad, "name_regexps")
		}
		if v, _ := block["value_regexps"].([]interface{}); len(v) > 0 {
			bad = append(bad, "value_regexps")
		}
		if v, _ := block["additional_parameters"].(bool); v {
			bad = append(bad, "additional_parameters")
		}
		if v, _ := block["plain_parameters"].(bool); v {
			bad = append(bad, "plain_parameters")
		}
		if len(bad) > 0 {
			return fmt.Errorf(
				"enumerated_parameters: %s not allowed when mode = %q (only `points` applies)",
				strings.Join(bad, ", "), modeExact,
			)
		}
	case modeRegexp:
		if v, _ := block["points"].([]interface{}); len(v) > 0 {
			return fmt.Errorf(
				"enumerated_parameters: `points` not allowed when mode = %q (use name_regexps / value_regexps)",
				modeRegexp,
			)
		}
	}
	return nil
}
