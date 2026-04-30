package resourcerule

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
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
	return validateEnumeratedParamsBlock(block, d.GetRawConfig())
}

// validateEnumeratedParamsBlock is the pure-data half of
// EnumeratedParamsCustomizeDiff — separated so it can be unit-tested with
// hand-built `block` maps and `cty.Value` fixtures, no *schema.ResourceDiff.
//
// `rawCfg` is the resource-level cty.Value from d.GetRawConfig(); used to
// detect bool fields the user actually wrote in HCL (vs. fields the SDK
// filled from `Default: false`). When unavailable in tests, pass cty.NilVal.
func validateEnumeratedParamsBlock(block map[string]interface{}, rawCfg cty.Value) error {
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
		if isEnumeratedParamFieldSetInRawConfig(rawCfg, "additional_parameters") {
			bad = append(bad, "additional_parameters")
		}
		if isEnumeratedParamFieldSetInRawConfig(rawCfg, "plain_parameters") {
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

// isEnumeratedParamFieldSetInRawConfig returns true when the user's literal
// HCL contains a non-null value for `enumerated_parameters[0].<field>`. The
// SDK's `Default: false` on the booleans means d.Get cannot tell whether the
// user wrote the field — only GetRawConfig sees the unfilled form. Returns
// false when rawCfg is unavailable (e.g. cty.NilVal in unit tests),
// effectively allowing the field; tests that need the strict semantics must
// pass a populated rawCfg.
func isEnumeratedParamFieldSetInRawConfig(rawCfg cty.Value, field string) bool {
	if rawCfg == cty.NilVal || !rawCfg.IsKnown() || rawCfg.IsNull() {
		return false
	}
	ty := rawCfg.Type()
	if !ty.IsObjectType() || !ty.HasAttribute("enumerated_parameters") {
		return false
	}
	ep := rawCfg.GetAttr("enumerated_parameters")
	if ep.IsNull() || !ep.IsKnown() {
		return false
	}
	if !ep.Type().IsListType() && !ep.Type().IsSetType() && !ep.Type().IsTupleType() {
		return false
	}
	elems := ep.AsValueSlice()
	if len(elems) == 0 {
		return false
	}
	first := elems[0]
	if first.IsNull() || !first.IsKnown() {
		return false
	}
	if !first.Type().IsObjectType() || !first.Type().HasAttribute(field) {
		return false
	}
	v := first.GetAttr(field)
	return !v.IsNull() && v.IsKnown()
}
