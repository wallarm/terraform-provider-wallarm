package wallarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

// String constants used throughout data_source_hits.go.
const (
	hitsPointKeyActionName = "action_name"
	hitsPointKeyHeader     = "header"
	hitsPointKeyInstance   = "instance"
	hitsCondTypeAbsent     = "absent"
	hitsPathMultiple       = "[multiple]"
)

var defaultAllowedAttackTypes = []string{
	"xss", "sqli", "rce", "ptrav", "crlf", "redir",
	"nosqli", "ldapi", "scanner", "mass_assignment", "ssrf",
	"ssi", "mail_injection", "ssti", "xxe", "invalid_xml",
}

func dataSourceWallarmHits() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWallarmHitsRead,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"request_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique request identifier to fetch all related hits",
			},

			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "request",
				ValidateFunc: validation.StringInSlice([]string{"request", "attack"}, false),
				Description:  "Fetch mode: 'request' fetches hits for the request_id only; 'attack' expands to all related hits by attack_id",
			},

			"attack_types": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Allowed attack types for filtering. In attack mode, controls which types to fetch from the API. In all modes, only hits matching these types produce rules. Defaults to the standard FP-relevant types.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"rule_types": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Rule types to generate. Defaults to both disable_stamp and disable_attack_type.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(validRuleTypes, false),
				},
			},

			"include_instance": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Include instance (pool ID) in action conditions. When true (default), rules are scoped to the hit's application instance.",
			},

			"time": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    2,
				Description: "Time range as [from, to] unix timestamps. Defaults to [6 months ago, now]",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},

			// Uses the exact same schema as all rule resources so the output
			// can be passed directly into any wallarm_rule_* action argument.
			"action": resourcerule.ScopeActionSchema(),

			"action_hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA256 of the sorted action conditions (Ruby-compatible ConditionsHash)",
			},

			"action_dir_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Computed directory name for organizing rule files by action scope",
			},

			"action_conditions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"point": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Description: "Action conditions in API format (type/point/value list). Used for .action.yaml generation.",
			},

			"aggregated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON-encoded compact representation: {action_hash, action, groups}. Groups aggregate stamps and attack_types per detection point. Use this for caching instead of rules to avoid duplicating action data.",
			},

			"hits_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of hits fetched.",
			},
			"hits": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"statuscode": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"stamps": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"stamps_hash": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"point": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"point_wrapped": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeList,
								Elem: &schema.Schema{Type: schema.TypeString},
							},
						},
						"point_hash": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "SHA256 hash of the detection point (Ruby-compatible).",
						},
						"poolid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"attack_id": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"block_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"request_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"known_attack": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"node_uuid": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceWallarmHitsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	requestID := d.Get("request_id").(string)
	mode := d.Get("mode").(string)
	attackTypes := resolveAttackTypes(d)
	ruleTypes := resolveRuleTypes(d)

	timeRange := buildTimeRange(d)

	// Phase 1: Fetch direct hits by request_id.
	directHits, err := fetchDirectHits(client, clientID, requestID, timeRange)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set stable resource ID.
	resourceID := fmt.Sprintf("hits_%d_%s", clientID, requestID)
	if mode == "attack" {
		resourceID += "_attack"
	}
	d.SetId(resourceID)

	if len(directHits) == 0 {
		return setEmptyHitsState(d)
	}

	// Validate all direct hits share the same action.
	refDomain := directHits[0].Domain
	refPath := directHits[0].Path
	refPoolID := directHits[0].PoolID
	for _, h := range directHits[1:] {
		if h.Domain != refDomain || h.Path != refPath || h.PoolID != refPoolID {
			return diag.FromErr(fmt.Errorf(
				"inconsistent hit data for request_id %s: expected domain=%s path=%s poolid=%d, got domain=%s path=%s poolid=%d",
				requestID, refDomain, refPath, refPoolID, h.Domain, h.Path, h.PoolID,
			))
		}
	}

	// Phase 2 & 3: In attack mode, expand to related hits.
	allHits := directHits
	if mode == "attack" {
		relatedHits, err := fetchRelatedHitsByAttackIDs(client, clientID, directHits, attackTypes, timeRange, refDomain, refPath, refPoolID)
		if err != nil {
			return diag.FromErr(err)
		}
		allHits = mergeHits(directHits, relatedHits)
	}

	// Phase 4: Build action conditions, compute hash and dir name.
	includeInstance := d.Get("include_instance").(bool)
	action := buildActionFromHit(refDomain, refPath, refPoolID, includeInstance)
	actionDetails := schemaActionToDetails(action)
	actionHash := resourcerule.ConditionsHash(actionDetails)
	actionDirName := resourcerule.ActionDirName(actionDetails)

	// Phase 5c: Validate action conditions against API (ActionReadByHitID).
	if len(directHits[0].ID) >= 2 {
		apiResp, err := client.ActionReadByHitID(directHits[0].ID)
		if err != nil {
			log.Printf("[WARN] wallarm_hits: failed to validate action via ActionReadByHitID: %v", err)
		} else {
			apiHash := resourcerule.ConditionsHash(apiResp.Body.Conditions)

			if apiHash != actionHash {
				var msg strings.Builder
				fmt.Fprintf(&msg, "wallarm_hits: action conditions mismatch for hit %v\n", directHits[0].ID)
				fmt.Fprintf(&msg, "  provider hash=%s, API hash=%s\n", actionHash[:16], apiHash[:16])
				fmt.Fprintf(&msg, "  provider conditions (%d):\n", len(actionDetails))
				for i, c := range actionDetails {
					fmt.Fprintf(&msg, "    [%d] type=%q point=%v value=%v\n", i, c.Type, c.Point, c.Value)
				}
				fmt.Fprintf(&msg, "  API conditions (%d):\n", len(apiResp.Body.Conditions))
				for i, c := range apiResp.Body.Conditions {
					fmt.Fprintf(&msg, "    [%d] type=%q point=%v value=%v\n", i, c.Type, c.Point, c.Value)
				}
				return diag.Errorf("%s", msg.String())
			}
			log.Printf("[DEBUG] wallarm_hits: action hash validated against API: %s", actionHash[:8])
		}
	}

	actionSet := actionToSchemaSet(action)
	hitsForSchema := hitsToSchemaList(allHits)

	// Group hits by point for aggregated output.
	groups, schemaActions := groupHitsForRules(allHits, actionDetails, attackTypes)

	if err := d.Set("action", actionSet); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action: %s", err))
	}
	if err := d.Set("action_hash", actionHash); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action_hash: %s", err))
	}
	if err := d.Set("action_conditions", flattenActionConditions(actionDetails)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action_conditions: %s", err))
	}
	if err := d.Set("action_dir_name", actionDirName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action_dir_name: %s", err))
	}

	// Build aggregated output: compact representation for caching.
	aggregatedJSON, err := buildAggregatedJSON(actionHash, schemaActions, groups, ruleTypes)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aggregated", aggregatedJSON); err != nil {
		return diag.FromErr(fmt.Errorf("error setting aggregated: %s", err))
	}

	d.Set("hits_count", len(allHits))
	if err := d.Set("hits", hitsForSchema); err != nil {
		return diag.FromErr(fmt.Errorf("error setting hits: %s", err))
	}

	return nil
}

// buildTimeRange extracts the time range from schema or defaults to 6 months.
func buildTimeRange(d *schema.ResourceData) [][]interface{} {
	if v, ok := d.GetOk("time"); ok {
		tl := v.([]interface{})
		if len(tl) == 2 {
			return [][]interface{}{{tl[0], tl[1]}}
		}
	}
	sixMonthsAgo := time.Now().AddDate(0, -6, 0).Unix()
	now := time.Now().Unix()
	return [][]interface{}{{sixMonthsAgo, now}}
}

// resolveAttackTypes returns the attack types to filter by, using schema override or defaults.
func resolveAttackTypes(d *schema.ResourceData) []string {
	if v, ok := d.GetOk("attack_types"); ok {
		items := v.([]interface{})
		types := make([]string, 0, len(items))
		for _, item := range items {
			types = append(types, item.(string))
		}
		if len(types) > 0 {
			return types
		}
	}
	return defaultAllowedAttackTypes
}

// fetchDirectHits fetches hits by request_id with standard noise filters.
func fetchDirectHits(client wallarm.API, clientID int, requestID string, timeRange [][]interface{}) ([]*wallarm.Hit, error) {
	resp, err := client.HitRead(&wallarm.HitReadRequest{
		Filter: &wallarm.HitFilter{
			ClientID:        clientID,
			RequestID:       requestID,
			State:           nil,
			NotType:         []string{"warn", "infoleak"},
			Time:            timeRange,
			NotState:        "falsepositive",
			SecurityIssueID: nil,
			NotExperimental: true,
			NotAasmEvent:    true,
		},
		Limit:     HitFetchBatchSize,
		Offset:    0,
		OrderBy:   "time",
		OrderDesc: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error reading hits for request_id %s: %w", requestID, err)
	}
	return resp, nil
}

// fetchRelatedHitsByAttackIDs collects attack_ids from direct hits, then fetches
// all related hits in batches, filtering by allowed attack types and matching action.
func fetchRelatedHitsByAttackIDs(
	client wallarm.API,
	clientID int,
	directHits []*wallarm.Hit,
	attackTypes []string,
	timeRange [][]interface{},
	refDomain, refPath string,
	refPoolID int,
) ([]*wallarm.Hit, error) {
	// Collect unique attack IDs. Each hit's AttackID is ["index_name", "actual_id"].
	// The API filter expects [][]string: [["index","id1"],["index","id2"]].
	// Deduplicate by the actual ID (last element).
	seen := make(map[string]bool)
	attackIDs := make([][]string, 0, len(directHits))
	for _, h := range directHits {
		if len(h.AttackID) < 2 {
			continue
		}
		actualID := h.AttackID[len(h.AttackID)-1]
		if actualID == "" || seen[actualID] {
			continue
		}
		seen[actualID] = true
		attackIDs = append(attackIDs, h.AttackID)
	}

	if len(attackIDs) == 0 {
		log.Printf("[DEBUG] No attack_ids found in direct hits, skipping related hits fetch")
		return nil, nil
	}

	log.Printf("[INFO] Fetching related hits for %d attack_ids", len(seen))

	// Fetch in batches.
	var allRelated []*wallarm.Hit
	offset := 0
	discarded := 0

	for {
		resp, err := client.HitRead(&wallarm.HitReadRequest{
			Filter: &wallarm.HitFilter{
				ClientID:          clientID,
				AttackID:          attackIDs,
				Type:              attackTypes,
				State:             nil,
				Time:              timeRange,
				NotState:          "falsepositive",
				SecurityIssueID:   nil,
				NotExperimental:   true,
				NotAasmEvent:      true,
				NotWallarmScanner: true,
			},
			Limit:     HitFetchBatchSize,
			Offset:    offset,
			OrderBy:   "time",
			OrderDesc: true,
		})
		if err != nil {
			return nil, fmt.Errorf("error fetching related hits at offset %d: %w", offset, err)
		}

		if len(resp) == 0 {
			break
		}

		// Filter by matching action (domain + path + poolid).
		// When refPath is hitsPathMultiple, the attack spans multiple paths — match on domain + poolid only.
		for _, h := range resp {
			pathMatch := refPath == hitsPathMultiple || h.Path == refPath
			if h.Domain == refDomain && pathMatch && h.PoolID == refPoolID {
				if refPath == hitsPathMultiple && h.Path != hitsPathMultiple {
					log.Printf("[DEBUG] Including related hit %v with path=%s (refPath=[multiple], domain+poolid match)", h.ID, h.Path)
				}
				allRelated = append(allRelated, h)
			} else {
				discarded++
				log.Printf("[DEBUG] Discarding related hit %v: action mismatch (domain=%s path=%s poolid=%d)",
					h.ID, h.Domain, h.Path, h.PoolID)
			}
		}

		if len(resp) < HitFetchBatchSize {
			break
		}
		offset += HitFetchBatchSize
	}

	log.Printf("[INFO] Fetched %d related hits (%d discarded for action mismatch)", len(allRelated), discarded)
	return allRelated, nil
}

// mergeHits combines direct and related hits, deduplicating by hit ID.
func mergeHits(direct, related []*wallarm.Hit) []*wallarm.Hit {
	seen := make(map[string]bool, len(direct))
	for _, h := range direct {
		seen[hitKey(h)] = true
	}

	merged := make([]*wallarm.Hit, len(direct), len(direct)+len(related))
	copy(merged, direct)

	for _, h := range related {
		key := hitKey(h)
		if !seen[key] {
			seen[key] = true
			merged = append(merged, h)
		}
	}

	return merged
}

// hitKey returns a string key for deduplication based on hit ID components.
func hitKey(h *wallarm.Hit) string {
	return strings.Join(h.ID, "/")
}

// setEmptyHitsState sets empty values for all computed fields.
func setEmptyHitsState(d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	if err := d.Set("action", schema.NewSet(resourcerule.HashActionDetails, []interface{}{})); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	d.Set("action_hash", "")
	d.Set("action_dir_name", "")
	if err := d.Set("action_conditions", []interface{}{}); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("aggregated", `{"action_hash":"","action":[],"groups":[]}`); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	d.Set("hits_count", 0)
	if err := d.Set("hits", []interface{}{}); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	return diags
}

// schemaActionToDetails converts []map[string]interface{} (schema format) to []ActionDetails
// for use with ConditionsHash. Handles the schema convention where point-value types
// (action_name, method, etc.) store the value in the point map, not the value field.
func schemaActionToDetails(action []map[string]interface{}) []wallarm.ActionDetails {
	details := make([]wallarm.ActionDetails, 0, len(action))
	for _, m := range action {
		condType, _ := m["type"].(string)
		condValue, _ := m["value"].(string)
		pointMap, _ := m["point"].(map[string]interface{})

		var point []interface{}
		var value interface{}

		for key, val := range pointMap {
			valStr, _ := val.(string)
			switch key {
			case hitsPointKeyHeader, "query":
				point = []interface{}{key, valStr}
				value = condValue
			case "path":
				idx, _ := strconv.Atoi(valStr)
				point = []interface{}{key, float64(idx)}
				if condType == hitsCondTypeAbsent {
					value = nil
				} else {
					value = condValue
				}
			case hitsPointKeyInstance, hitsPointKeyActionName, "action_ext", "method", "proto", "scheme", "uri":
				point = []interface{}{key}
				if condType == hitsCondTypeAbsent {
					value = nil
				} else {
					value = valStr
				}
			}
		}

		details = append(details, wallarm.ActionDetails{
			Type:  condType,
			Point: point,
			Value: value,
		})
	}
	return details
}

// actionToSchemaSet converts action conditions to a schema.Set.
func actionToSchemaSet(action []map[string]interface{}) *schema.Set {
	ifaces := make([]interface{}, len(action))
	for i, a := range action {
		ifaces[i] = a
	}
	return schema.NewSet(resourcerule.HashActionDetails, ifaces)
}

// hitsToSchemaList converts wallarm.Hit objects to the schema list format.
func hitsToSchemaList(hits []*wallarm.Hit) []interface{} {
	result := make([]interface{}, 0, len(hits))
	for _, h := range hits {
		pointStrings := make([]interface{}, 0, len(h.Point))
		for _, p := range h.Point {
			pointStrings = append(pointStrings, fmt.Sprintf("%v", p))
		}

		pointWrapped := resourcerule.WrapPointElements(h.Point)
		wrappedForSchema := make([]interface{}, 0, len(pointWrapped))
		for _, pw := range pointWrapped {
			inner := make([]interface{}, 0, len(pw))
			for _, s := range pw {
				inner = append(inner, s)
			}
			wrappedForSchema = append(wrappedForSchema, inner)
		}

		result = append(result, map[string]interface{}{
			"id":            h.ID,
			"type":          h.Type,
			"ip":            h.IP,
			"statuscode":    h.StatusCode,
			"time":          h.Time,
			"value":         h.Value,
			"stamps":        h.Stamps,
			"stamps_hash":   h.StampsHash,
			"point":         pointStrings,
			"point_wrapped": wrappedForSchema,
			"point_hash":    resourcerule.PointHash(h.Point),
			"poolid":        h.PoolID,
			"attack_id":     h.AttackID,
			"block_status":  h.BlockStatus,
			"request_id":    h.RequestID,
			"domain":        h.Domain,
			"path":          h.Path,
			"protocol":      h.Protocol,
			"known_attack":  h.KnownAttack,
			"node_uuid":     h.NodeUUID,
		})
	}
	return result
}

// buildActionFromHit converts hit domain, path and poolid into Wallarm rule
// action conditions in the exact format used by wallarm_rule_* resources.
//
// Conventions match HashActionDetails in resourcerule:
//
//	point type    | type   | value  | point map
//	--------------+--------+--------+-------------------------
//	instance      | equal  | ""     | {"instance": "<id>"}
//	header        | iequal | domain | {"header": "HOST"}
//	path (equal)  | equal  | seg    | {"path": "<N>"}
//	path (absent) | absent | ""     | {"path": "<N>"}
//	action_name   | equal  | ""     | {"action_name": name}
//	action_ext    | equal  | ""     | {"action_ext": ext}
func buildActionFromHit(domain, urlPath string, poolID int, includeInstance bool) []map[string]interface{} {
	var conditions []map[string]interface{}

	// Instance — included when includeInstance is true (default).
	// Note: the API's ActionReadByHitID does NOT include instance in conditions,
	// so validation strips it before comparing hashes.
	if includeInstance && poolID > 0 {
		conditions = append(conditions, map[string]interface{}{
			"type":  "equal",
			"value": "",
			"point": map[string]interface{}{hitsPointKeyInstance: strconv.Itoa(poolID)},
		})
	}

	// HOST header — always iequal.
	if domain != "" {
		conditions = append(conditions, map[string]interface{}{
			"type":  "iequal",
			"value": domain,
			"point": map[string]interface{}{hitsPointKeyHeader: "HOST"},
		})
	}

	// hitsPathMultiple means the attack spans multiple paths — no path/action_name/action_ext
	// conditions, producing a /**/*.* wildcard scope (HOST header only).
	if urlPath == hitsPathMultiple {
		log.Printf("[INFO] wallarm_hits: path=[multiple] — skipping path/action_name/action_ext conditions (wildcard scope)")
	} else {
		conditions = append(conditions, locationToConditions(urlPath)...)
	}

	return conditions
}

// locationToConditions converts a URL path into action conditions.
// Port of the Ruby LocationToConditions class.
func locationToConditions(location string) []map[string]interface{} {
	// URI point is mutually exclusive with path/action_name/action_ext (validated
	// in ActionScopeCustomizeDiff). Path is always decomposed into segments.

	parts := strings.Split(location, "/")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	// Root path "/" → action_name is empty string, path[0] is absent.
	if len(parts) == 0 {
		return []map[string]interface{}{
			{
				"type":  "equal",
				"value": "",
				"point": map[string]interface{}{hitsPointKeyActionName: ""},
			},
			{
				"type":  hitsCondTypeAbsent,
				"value": "",
				"point": map[string]interface{}{"path": "0"},
			},
		}
	}

	last := parts[len(parts)-1]
	pathParts := parts[:len(parts)-1]

	// Pre-allocate: len(pathParts) path conditions + up to 2 from actionNameExtConditions + 1 terminating absent.
	conditions := make([]map[string]interface{}, 0, len(parts)+2)

	conditions = append(conditions, actionNameExtConditions(last)...)

	for i, part := range pathParts {
		conditions = append(conditions, map[string]interface{}{
			"type":  "equal",
			"value": part,
			"point": map[string]interface{}{"path": strconv.Itoa(i)},
		})
	}

	// Terminating absent — fixes the length of the path chain.
	conditions = append(conditions, map[string]interface{}{
		"type":  hitsCondTypeAbsent,
		"value": "",
		"point": map[string]interface{}{"path": strconv.Itoa(len(pathParts))},
	})

	return conditions
}

// actionNameExtConditions splits a path segment into action_name / action_ext.
// The matched string goes in the point map value; value field is always "".
func actionNameExtConditions(segment string) []map[string]interface{} {
	if dotIdx := strings.LastIndex(segment, "."); dotIdx >= 0 {
		name := segment[:dotIdx]
		ext := segment[dotIdx+1:]
		return []map[string]interface{}{
			{
				"type":  "equal",
				"value": "",
				"point": map[string]interface{}{hitsPointKeyActionName: name},
			},
			{
				"type":  "equal",
				"value": "",
				"point": map[string]interface{}{"action_ext": ext},
			},
		}
	}

	return []map[string]interface{}{
		{
			"type":  "equal",
			"value": "",
			"point": map[string]interface{}{hitsPointKeyActionName: segment},
		},
		{
			"type":  hitsCondTypeAbsent,
			"value": "",
			"point": map[string]interface{}{"action_ext": ""},
		},
	}
}

// groupHitsForRules groups hits by point_hash, filtering by allowed attack types,
// and converts action details to schema format. Returns the groups and schema actions
// for use by both the expanded rules output and the aggregated output.
func groupHitsForRules(hits []*wallarm.Hit, actionDetails []wallarm.ActionDetails, attackTypes []string) (map[string]*pointGroup, []map[string]interface{}) {
	// Build attack type filter set.
	attackTypeSet := make(map[string]bool, len(attackTypes))
	for _, at := range attackTypes {
		attackTypeSet[at] = true
	}

	// Group hits by point_hash + attack_type. Each stamp belongs to a specific
	// attack type, so grouping per type gives accurate stamp-to-type traceability.
	groups := make(map[string]*pointGroup)
	for _, h := range hits {
		if !attackTypeSet[h.Type] {
			continue
		}

		ph := resourcerule.PointHash(h.Point)
		if ph == "" {
			continue
		}

		// Key includes attack type — each type gets its own stamp list.
		groupKey := ph
		if h.Type != "" {
			groupKey = ph + "_" + h.Type
		}

		g, exists := groups[groupKey]
		if !exists {
			wrapped := resourcerule.WrapPointElements(h.Point)
			pointStrs := make([][]string, 0, len(wrapped))
			pointStrs = append(pointStrs, wrapped...)
			g = &pointGroup{PointWrapped: pointStrs}
			groups[groupKey] = g
		}

		// Merge stamps for this attack type.
		for _, s := range h.Stamps {
			if s > 0 && !containsInt(g.Stamps, s) {
				g.Stamps = append(g.Stamps, s)
			}
		}

		// Track attack type (one per group).
		if h.Type != "" && !containsStr(g.AttackTypes, h.Type) {
			g.AttackTypes = append(g.AttackTypes, h.Type)
		}
	}

	// Sort for deterministic output.
	for _, g := range groups {
		sort.Ints(g.Stamps)
		sort.Strings(g.AttackTypes)
	}

	// Convert action details to schema format.
	schemaActions := make([]map[string]interface{}, 0, len(actionDetails))
	for _, ad := range actionDetails {
		item := resourcerule.ActionDetailToSchemaItem(ad)
		pointMap := make(map[string]interface{})
		if pm, ok := item["point"].(map[string]interface{}); ok {
			for k, v := range pm {
				pointMap[k] = fmt.Sprintf("%v", v)
			}
		}
		schemaActions = append(schemaActions, map[string]interface{}{
			"type":  item["type"],
			"value": item["value"],
			"point": pointMap,
		})
	}

	return groups, schemaActions
}

// aggregatedGroup is one entry in the aggregated JSON output.
// Each group is keyed by point_hash + attack_type. Contains stamps for that
// attack type at that point, plus the attack_type itself. Stampless types
// (xxe, invalid_xml) have empty stamps but still produce disable_attack_type rules.
type aggregatedGroup struct {
	Key               string     `json:"key"`
	Point             [][]string `json:"point"`
	Stamps            []int      `json:"stamps"`
	AttackType        string     `json:"attack_type"`
	DisableAttackType bool       `json:"disable_attack_type"`
}

// aggregatedOutput is the compact representation stored in the aggregated field.
type aggregatedOutput struct {
	ActionHash string                   `json:"action_hash"`
	Action     []map[string]interface{} `json:"action"`
	Groups     []aggregatedGroup        `json:"groups"`
}

// buildAggregatedJSON builds the compact JSON for the aggregated output.
// ruleTypes filters which data is included: stamps for disable_stamp, attack_types for disable_attack_type.
func buildAggregatedJSON(actionHash string, schemaActions []map[string]interface{}, groups map[string]*pointGroup, ruleTypes []string) (string, error) {
	rtSet := make(map[string]bool, len(ruleTypes))
	for _, rt := range ruleTypes {
		rtSet[rt] = true
	}
	includeStamps := rtSet[ruleTypeDisableStamp]
	includeAttackTypes := rtSet[ruleTypeDisableAttackType]

	// Sort group keys for deterministic output.
	phKeys := make([]string, 0, len(groups))
	for ph := range groups {
		phKeys = append(phKeys, ph)
	}
	sort.Strings(phKeys)

	aggGroups := make([]aggregatedGroup, 0, len(groups))
	for _, gk := range phKeys {
		g := groups[gk]

		// Group key is "point_hash_attack_type". Truncate only the point_hash part.
		attackType := ""
		if len(g.AttackTypes) > 0 {
			attackType = g.AttackTypes[0]
		}
		// Extract point_hash by stripping the "_attack_type" suffix.
		phPart := gk
		if attackType != "" && len(gk) > len(attackType)+1 {
			phPart = gk[:len(gk)-len(attackType)-1]
		}
		prefix := phPart[:min(16, len(phPart))]
		if attackType != "" {
			prefix = prefix + "_" + attackType
		}

		// Determine what to include based on rule_types filter.
		// Ensure non-nil slice (nil marshals to JSON null, causing HCL errors).
		stamps := g.Stamps
		if stamps == nil || !includeStamps {
			stamps = []int{}
		}

		// Skip group if nothing to include after filtering.
		hasStamps := len(stamps) > 0
		hasAttackType := includeAttackTypes && attackType != ""
		if !hasStamps && !hasAttackType {
			continue
		}

		aggGroups = append(aggGroups, aggregatedGroup{
			Key:               prefix,
			Point:             g.PointWrapped,
			Stamps:            stamps,
			AttackType:        attackType,
			DisableAttackType: hasAttackType,
		})
	}

	out := aggregatedOutput{
		ActionHash: actionHash[:min(16, len(actionHash))],
		Action:     schemaActions,
		Groups:     aggGroups,
	}

	data, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("failed to marshal aggregated output: %w", err)
	}
	return string(data), nil
}
