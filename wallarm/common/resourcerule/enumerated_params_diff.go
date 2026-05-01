package resourcerule

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// EnumeratedParamsCustomizeDiff fails plan when the user has populated fields
// in `enumerated_parameters` that the Wallarm API ignores for the chosen
// `mode`. The TF→API mapper silently drops mismatched fields on PUT, which
// produces a perpetual plan diff: state is read back without them, the next
// plan re-emits the same change, the next apply drops them again.
//
// `mode = "exact"`  → only `points` applies. `additional_parameters` and
// `plain_parameters` are denied even when set to false: with `Default: false`
// the SDK fills the schema, so checking `d.Get` cannot distinguish "user
// wrote false" from "user omitted, default applied". `d.GetRawConfig()` is
// the authoritative view of the user's HCL — null at a path means the user
// did not write it. We deny on presence, regardless of value.
//
// `mode = "regexp"` → `name_regexps` and `value_regexps` are required (each
// at least one element, may be `[""]`). The Wallarm API rejects regexp-mode
// hints without these keys; previously the mapper substituted `[""]`
// silently, which round-tripped from the API as `[null]` in state and
// produced a perpetual diff. Forcing explicit user-supplied values keeps
// HCL and state aligned. `points` is rejected (regexp mode ignores it).
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
// EnumeratedParamsCustomizeDiff — separated so it can be unit-tested with
// hand-built `block` maps, no *schema.ResourceDiff.
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
		// Bool fields: error only when true. The schema is Optional+Computed
		// (no Default), so omitted values stay at the SDK's bool zero
		// (`false`); auto-generated configs after import may surface explicit
		// `false`, and erroring on those would block valid round-trips. The
		// mapper drops these fields from the wire body in exact mode
		// regardless of value, so a literal `false` in HCL is harmless.
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
		var missing []string
		if v, _ := block["name_regexps"].([]interface{}); len(v) == 0 {
			missing = append(missing, "name_regexps")
		}
		if v, _ := block["value_regexps"].([]interface{}); len(v) == 0 {
			missing = append(missing, "value_regexps")
		}
		if len(missing) > 0 {
			return fmt.Errorf(
				"enumerated_parameters: %s required when mode = %q — set [\"\"] to opt out of that filter",
				strings.Join(missing, ", "), modeRegexp,
			)
		}
	}
	return nil
}
