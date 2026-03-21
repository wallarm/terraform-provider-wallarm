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
		ForceNew:    true,
		Description: "URL path pattern. Supports wildcards: * (any segment), ** (any depth). Expands into action conditions automatically.",
	},
	"action_domain": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "Domain (HOST header) to match. Use * for any domain.",
	},
	"action_instance": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "Application instance (pool) ID.",
	},
	"action_method": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "HTTP method to match (GET, POST, etc.).",
	},
	"action_scheme": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "URL scheme to match (http, https).",
	},
	"action_proto": {
		Type:        schema.TypeString,
		Optional:    true,
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
func ScopeActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice([]string{"equal", "iequal", "regex", "absent"}, false),
				},
				"value": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Computed: true,
				},
				"point": {
					Type:     schema.TypeMap,
					Optional: true,
					ForceNew: true,
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
func ActionScopeCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	// Check if scope fields are set in config.
	actionPath := d.Get("action_path").(string)

	hasScopeFields := actionPath != "" ||
		d.Get("action_domain").(string) != "" ||
		d.Get("action_instance").(string) != "" ||
		d.Get("action_method").(string) != ""

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
		items = append(items, actionDetailToSchemaItem(a))
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

// actionDetailToSchemaItem converts an ActionDetails to the map format expected by
// ResourceDiff.SetNew for the action TypeSet. Uses map[string]interface{} for point
// (not map[string]string) and explicit string values (not empty with Computed).
func actionDetailToSchemaItem(a wallarm.ActionDetails) map[string]interface{} {
	pointKey := ActionPointKey(a)
	value := ActionValueString(a)
	condType := a.Type

	pointMap := map[string]interface{}{}

	switch pointKey {
	case "header":
		pointMap["header"] = strings.ToUpper(ActionPointSecond(a))
	case "get":
		pointMap["query"] = ActionPointSecond(a)
	case "path":
		pointMap["path"] = fmt.Sprintf("%d", ActionPointIndex(a))
	case "instance":
		pointMap["instance"] = value
		value = ""
		condType = ""
	case "action_name":
		pointMap["action_name"] = value
		value = ""
	case "action_ext":
		pointMap["action_ext"] = value
		value = ""
	case "method":
		pointMap["method"] = value
		value = ""
	case "scheme":
		pointMap["scheme"] = value
		value = ""
	case "proto":
		pointMap["proto"] = value
		value = ""
	case "uri":
		pointMap["uri"] = value
		value = ""
	}

	return map[string]interface{}{
		"type":  condType,
		"value": value,
		"point": pointMap,
	}
}
