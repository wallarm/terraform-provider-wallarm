package resourcerule

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	wallarm "github.com/wallarm/wallarm-go"
)

// ActionScopeFields are the user-friendly fields for defining action conditions
// via path/domain/method/etc. instead of explicit action blocks.
// These are added to every rule resource via lo.Assign alongside commonResourceRuleFields.
//
// When scope fields are set, action blocks are computed by CustomizeDiff.
// When action blocks are set directly, scope fields are ignored.
// ConflictsWith prevents mixing both styles.
var ActionScopeFields = map[string]*schema.Schema{
	"action_path": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		Description: "URL path pattern. Supports wildcards: * (any segment), ** (any depth). Expands into action conditions automatically.",
	},
	"action_domain": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		Description: "Domain (HOST header) to match. Use * for any domain.",
	},
	"action_instance": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		Description: "Application instance (pool) ID.",
	},
	"action_method": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		Description: "HTTP method to match (GET, POST, etc.).",
	},
	"action_scheme": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		Description: "URL scheme to match (http, https).",
	},
	"action_proto": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		Description: "HTTP protocol version (1.0, 1.1, 2.0).",
	},
	"action_query": {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key":   {Type: schema.TypeString, Required: true, ForceNew: true},
				"value": {Type: schema.TypeString, Required: true, ForceNew: true},
				"type": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "equal",
					ForceNew: true,
				},
			},
		},
		Description: "Query parameter conditions.",
	},
	"action_header": {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":  {Type: schema.TypeString, Required: true, ForceNew: true},
				"value": {Type: schema.TypeString, Required: true, ForceNew: true},
				"type": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "equal",
					ForceNew: true,
				},
			},
		},
		Description: "Custom header conditions (in addition to domain -> HOST).",
	},
}

// ScopeActionSchema returns the action schema modified to be Optional+Computed,
// so it can be either set directly by the user or computed from scope fields.
// Used by rule hints where a scope change means a different hint (ForceNew).
func ScopeActionSchema() *schema.Schema {
	return scopeActionSchema(true, true)
}

// ScopeActionSchemaMutable returns the action schema without ForceNew and
// without Computed, for resources whose API supports updating conditions in
// place (e.g. API spec policy PUT) and where the user config is authoritative.
// Uses the same element shape and HashActionDetails set function so all
// existing expand/flatten helpers continue to work.
func ScopeActionSchemaMutable() *schema.Schema {
	return scopeActionSchema(false, false)
}

// iequalSiblingMatchesCaseInsensitive returns true when the action element at
// `parent` has type "iequal" and old/newVal differ only in case. The API
// downcases iequal values server-side, so mixed-case HCL is equivalent to
// lowercased state.
func iequalSiblingMatchesCaseInsensitive(d *schema.ResourceData, parent, old, newVal string) bool {
	siblingType, _ := d.Get(parent + ".type").(string)
	if siblingType != "iequal" {
		return false
	}
	return strings.EqualFold(old, newVal)
}

// suppressIequalValueCaseDiff suppresses case-only diffs on action.value
// (paired-element points: header, query — where the matched string is in
// the value field).
func suppressIequalValueCaseDiff(k, old, newVal string, d *schema.ResourceData) bool {
	parent := strings.TrimSuffix(k, ".value")
	return iequalSiblingMatchesCaseInsensitive(d, parent, old, newVal)
}

// suppressIequalPointValueCaseDiff suppresses case-only diffs on
// action.point.<key> entries:
//   - <key> in PointValuePoints (action_name, action_ext, method, instance,
//     scheme, uri, proto): suppressed only when the sibling type is iequal
//     (the API downcases iequal values server-side).
//   - <key> == "header": always suppressed. HTTP header names are
//     case-insensitive per RFC 7230; the Wallarm API uppercases them on
//     receive, so config "referer" vs state "REFERER" is the same name.
func suppressIequalPointValueCaseDiff(k, old, newVal string, d *schema.ResourceData) bool {
	idx := strings.LastIndex(k, ".point.")
	parent := k[:idx]
	pointKey := k[idx+len(".point."):]
	if pointKey == Header {
		return strings.EqualFold(old, newVal)
	}
	if !PointValuePoints[pointKey] {
		return false
	}
	return iequalSiblingMatchesCaseInsensitive(d, parent, old, newVal)
}

func scopeActionSchema(forceNew, computed bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: computed,
		ForceNew: forceNew,
		Set:      HashActionDetails,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     computed,
					ForceNew:     forceNew,
					ValidateFunc: validation.StringInSlice([]string{"equal", "iequal", "regex", "absent", ""}, false),
				},
				"value": {
					Type:             schema.TypeString,
					Optional:         true,
					ForceNew:         forceNew,
					Computed:         computed,
					DiffSuppressFunc: suppressIequalValueCaseDiff,
				},
				"point": {
					Type:             schema.TypeMap,
					Optional:         true,
					ForceNew:         forceNew,
					DiffSuppressFunc: suppressIequalPointValueCaseDiff,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

// ActionScopeCustomizeDiff is a CustomizeDiffFunc that expands scope fields
// (action_path, action_domain, etc.) into the action TypeSet when scope fields are set.
// When scope fields are present, they always take priority over action blocks
// (which may be populated from state after a previous apply).
//
// For existing resources, action is only recomputed when a scope field actually
// changes. This prevents spurious drift from TypeSet representation differences.
//
// Also validates that "uri" conditions are not mixed with "path", "action_name",
// "action_ext", or "query" conditions (mutually exclusive in the Wallarm API).
func ActionScopeCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	// Validate explicit action blocks (point keys, URI conflicts, type/value rules).
	if err := validateActionBlocks(d); err != nil {
		return err
	}

	// Check if scope fields are set in config.
	actionPath := d.Get("action_path").(string)

	hasScopeFields := actionPath != "" ||
		d.Get("action_domain").(string) != "" ||
		d.Get("action_instance").(string) != "" ||
		d.Get("action_method").(string) != "" ||
		d.Get("action_scheme").(string) != "" ||
		d.Get("action_proto").(string) != ""

	// Also check list-type scope fields (query and header blocks).
	if !hasScopeFields {
		if v, ok := d.GetOk("action_query"); ok && len(v.([]interface{})) > 0 {
			hasScopeFields = true
		}
	}
	if !hasScopeFields {
		if v, ok := d.GetOk("action_header"); ok && len(v.([]interface{})) > 0 {
			hasScopeFields = true
		}
	}

	if !hasScopeFields {
		return nil
	}

	// For existing resources: only recompute action when a scope field changed.
	if d.Id() != "" && !anyScopeFieldChanged(d) {
		return nil
	}

	// Extract scope fields.
	domain := d.Get("action_domain").(string)
	instance := d.Get("action_instance").(string)
	method := d.Get("action_method").(string)
	schemeName := d.Get("action_scheme").(string)
	proto := d.Get("action_proto").(string)

	// Extract query params.
	var queryParams []QueryParam
	if v, ok := d.GetOk("action_query"); ok {
		for _, q := range v.([]interface{}) {
			qm := q.(map[string]interface{})
			queryParams = append(queryParams, QueryParam{
				Key:   qm["key"].(string),
				Value: qm["value"].(string),
				Type:  qm["type"].(string),
			})
		}
	}

	// Extract headers.
	var headerParams []HeaderParam
	if v, ok := d.GetOk("action_header"); ok {
		for _, h := range v.([]interface{}) {
			hm := h.(map[string]interface{})
			headerParams = append(headerParams, HeaderParam{
				Name:  hm["name"].(string),
				Value: hm["value"].(string),
				Type:  hm["type"].(string),
			})
		}
	}

	// Expand to action conditions.
	actions := ExpandPathToActions(actionPath, domain, instance, method, schemeName, proto, queryParams, headerParams)

	// Build Set items directly in TF schema format.
	// ResourceDiff.SetNew is stricter than ResourceData.Set:
	// - point must be map[string]interface{} (not map[string]string)
	// - value must be explicit string (not "" with Computed, which becomes "known after apply")
	items := make([]interface{}, 0, len(actions))
	for _, a := range actions {
		items = append(items, ActionDetailToSchemaItem(a))
	}

	log.Printf("[DEBUG] ActionScopeCustomizeDiff: expanded %d action conditions from path=%q domain=%q",
		len(actions), actionPath, domain)

	return d.SetNew("action", schema.NewSet(
		schema.HashResource(ScopeActionSchema().Elem.(*schema.Resource)),
		items,
	))
}

// anyScopeFieldChanged returns true if any action_* scope field has changed.
func anyScopeFieldChanged(d *schema.ResourceDiff) bool {
	scopeFields := []string{
		"action_path", "action_domain", "action_instance",
		"action_method", "action_scheme", "action_proto",
		"action_query", "action_header",
	}
	for _, f := range scopeFields {
		if d.HasChange(f) {
			return true
		}
	}
	return false
}

// validPointKeys are all valid keys for the "point" map in action conditions.
var validPointKeys = map[string]bool{
	"header":      true,
	"method":      true,
	"path":        true,
	"action_name": true,
	"action_ext":  true,
	"query":       true,
	"proto":       true,
	"scheme":      true,
	"uri":         true,
	"instance":    true,
}

// PointValuePoints are points where the actual value goes in the point map
// and the "value" field must be "".
var PointValuePoints = map[string]bool{
	"action_name": true,
	"action_ext":  true,
	"method":      true,
	"proto":       true,
	"scheme":      true,
	"uri":         true,
	"instance":    true,
}

// uriConflictPoints are the action condition points that conflict with "uri".
var uriConflictPoints = map[string]bool{
	"path":        true,
	"action_name": true,
	"action_ext":  true,
	"query":       true,
}

// validateActionBlocks validates all explicit action blocks for:
//   - Valid point keys (no typos)
//   - Single key per point map (each condition targets one request part)
//   - URI conflict with path/action_name/action_ext/query
//   - Point-value points (action_name, method, etc.) require value = ""
//   - Header and query require non-empty value
func validateActionBlocks(d *schema.ResourceDiff) error {
	v, ok := d.GetOk("action")
	if !ok {
		return nil
	}
	actionSet, ok := v.(*schema.Set)
	if !ok {
		return nil
	}
	return validateActionSet(actionSet)
}

// validateActionSet is the unit-testable core of validateActionBlocks.
func validateActionSet(actionSet *schema.Set) error {
	if actionSet == nil || actionSet.Len() == 0 {
		return nil
	}

	var hasURI bool
	var conflictingPoints []string

	for _, item := range actionSet.List() {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		pointMap, ok := m["point"].(map[string]interface{})
		if !ok || len(pointMap) == 0 {
			// Skip empty point maps — can be TypeSet zero-value artifacts
			// from Computed action blocks during plan.
			continue
		}

		condType, _ := m["type"].(string)
		condValue, _ := m["value"].(string)

		// Single key per point map.
		if len(pointMap) > 1 {
			keys := make([]string, 0, len(pointMap))
			for k := range pointMap {
				keys = append(keys, k)
			}
			return fmt.Errorf("action block \"point\" must contain exactly one key, got %d: %v", len(pointMap), keys)
		}

		// Valid point key.
		for key := range pointMap {
			if !validPointKeys[key] {
				return fmt.Errorf("unknown action point key %q — valid keys: header, method, path, action_name, action_ext, query, proto, scheme, uri, instance", key)
			}

			// Track URI vs path/name/ext/query for conflict check.
			if key == pointKeyURI {
				hasURI = true
			}
			if uriConflictPoints[key] {
				conflictingPoints = append(conflictingPoints, key)
			}

			// Point-value points require value = "".
			if PointValuePoints[key] && condType != condTypeAbsent && condValue != "" {
				return fmt.Errorf("action condition with point %q: the value goes in the point map, \"value\" field must be empty", key)
			}

			// Header and query require non-empty value (it's the matched content).
			if (key == pointKeyHeader || key == "query") && condType != condTypeAbsent && condValue == "" {
				return fmt.Errorf("action condition with point %q requires a non-empty \"value\" (the content to match)", key)
			}
		}
	}

	// URI conflict check across all blocks.
	if hasURI && len(conflictingPoints) > 0 {
		return fmt.Errorf("action condition \"uri\" conflicts with %v — use one or the other (uri is a full URI match, while path/action_name/action_ext/query are decomposed parts)", conflictingPoints)
	}

	return nil
}

// ActionDetailToSchemaItem converts an ActionDetails to the map format expected by
// ResourceDiff.SetNew for the action TypeSet. Uses map[string]interface{} for point
// (not map[string]string) and explicit string values (not empty with Computed).
func ActionDetailToSchemaItem(a wallarm.ActionDetails) map[string]interface{} {
	pointKey := ActionPointKey(a)
	value := ActionValueString(a)
	condType := a.Type

	pointMap := map[string]interface{}{}

	switch pointKey {
	case pointKeyHeader:
		pointMap[pointKeyHeader] = strings.ToUpper(ActionPointSecond(a))
	case pointKeyGet:
		pointMap["query"] = ActionPointSecond(a)
	case pointKeyPath:
		pointMap[pointKeyPath] = fmt.Sprintf("%d", ActionPointIndex(a))
	case pointKeyInstance:
		pointMap[pointKeyInstance] = value
		value = ""
	case pointKeyActionName:
		pointMap[pointKeyActionName] = value
		value = ""
	case pointKeyActionExt:
		pointMap[pointKeyActionExt] = value
		value = ""
	case pointKeyMethod:
		pointMap[pointKeyMethod] = value
		value = ""
	case pointKeyScheme:
		pointMap[pointKeyScheme] = value
		value = ""
	case pointKeyProto:
		pointMap[pointKeyProto] = value
		value = ""
	case pointKeyURI:
		pointMap[pointKeyURI] = value
		value = ""
	}

	return map[string]interface{}{
		"type":  condType,
		"value": value,
		"point": pointMap,
	}
}
