package wallarm

import (
	"context"
	"crypto/sha256"
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
	wallarm "github.com/wallarm/wallarm-go"
)

const (
	maxPathDepth      = 10
	hitFetchBatchSize = 500
)

var defaultAllowedAttackTypes = []string{
	"xss", "sqli", "rce", "xxe", "ptrav", "crlf", "redir",
	"nosqli", "ldapi", "scanner", "mass_assignment", "ssrf",
	"ssi", "mail_injection", "ssti",
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
				Description: "Allowed attack types for filtering in attack mode. Defaults to the standard FP-relevant types.",
				Elem:        &schema.Schema{Type: schema.TypeString},
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
			"action": defaultResourceRuleActionSchema,

			"action_hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA256 of the sorted action conditions for grouping rules with the same scope",
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
		attackTypes := resolveAttackTypes(d)
		relatedHits, err := fetchRelatedHitsByAttackIDs(client, clientID, directHits, attackTypes, timeRange, refDomain, refPath, refPoolID)
		if err != nil {
			return diag.FromErr(err)
		}
		allHits = mergeHits(directHits, relatedHits)
	}

	// Phase 4: Build action, hash, and set state.
	action := buildActionFromHit(refDomain, refPath, refPoolID)
	actionHash := computeActionHash(action)
	actionSet := actionToSchemaSet(action)

	hitsForSchema := hitsToSchemaList(allHits)

	if err := d.Set("action", actionSet); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action: %s", err))
	}
	if err := d.Set("action_hash", actionHash); err != nil {
		return diag.FromErr(fmt.Errorf("error setting action_hash: %s", err))
	}
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
		Limit:     hitFetchBatchSize,
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
	// Collect unique attack IDs.
	attackIDSet := make(map[string]bool)
	for _, h := range directHits {
		for _, aid := range h.AttackID {
			if aid != "" {
				attackIDSet[aid] = true
			}
		}
	}

	if len(attackIDSet) == 0 {
		log.Printf("[DEBUG] No attack_ids found in direct hits, skipping related hits fetch")
		return nil, nil
	}

	attackIDs := make([]string, 0, len(attackIDSet))
	for aid := range attackIDSet {
		attackIDs = append(attackIDs, aid)
	}

	log.Printf("[INFO] Fetching related hits for %d attack_ids: %v", len(attackIDs), attackIDs)

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
			Limit:     hitFetchBatchSize,
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
		for _, h := range resp {
			if h.Domain == refDomain && h.Path == refPath && h.PoolID == refPoolID {
				allRelated = append(allRelated, h)
			} else {
				discarded++
				log.Printf("[DEBUG] Discarding related hit %v: action mismatch (domain=%s path=%s poolid=%d)",
					h.ID, h.Domain, h.Path, h.PoolID)
			}
		}

		if len(resp) < hitFetchBatchSize {
			break
		}
		offset += hitFetchBatchSize
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
	_ = d.Set("action", schema.NewSet(schema.HashResource(
		defaultResourceRuleActionSchema.Elem.(*schema.Resource)), []interface{}{}))
	_ = d.Set("action_hash", "")
	_ = d.Set("hits", []interface{}{})
	return nil
}

// computeActionHash produces a deterministic SHA256 hex string from action conditions.
func computeActionHash(action []map[string]interface{}) string {
	sorted := make([]map[string]interface{}, len(action))
	copy(sorted, action)
	sort.Slice(sorted, func(i, j int) bool {
		return fmt.Sprintf("%v", sorted[i]["point"]) < fmt.Sprintf("%v", sorted[j]["point"])
	})
	return hashAction(sorted)
}

// actionToSchemaSet converts action conditions to a schema.Set.
func actionToSchemaSet(action []map[string]interface{}) *schema.Set {
	ifaces := make([]interface{}, len(action))
	for i, a := range action {
		ifaces[i] = a
	}
	return schema.NewSet(
		schema.HashResource(defaultResourceRuleActionSchema.Elem.(*schema.Resource)),
		ifaces,
	)
}

// hitsToSchemaList converts wallarm.Hit objects to the schema list format.
func hitsToSchemaList(hits []*wallarm.Hit) []interface{} {
	result := make([]interface{}, 0, len(hits))
	for _, h := range hits {
		pointStrings := make([]interface{}, 0, len(h.Point))
		for _, p := range h.Point {
			pointStrings = append(pointStrings, fmt.Sprintf("%v", p))
		}

		pointWrapped := wrapPointElements(h.Point)
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
// Conventions match hashResponseActionDetails in utils.go:
//
//	point type    | type   | value  | point map
//	--------------+--------+--------+-------------------------
//	instance      | ""     | ""     | {"instance": "<id>"}
//	header        | iequal | domain | {"header": "HOST"}
//	path (equal)  | equal  | seg    | {"path": "<N>"}
//	path (absent) | absent | ""     | {"path": "<N>"}
//	action_name   | equal  | ""     | {"action_name": name}
//	action_ext    | equal  | ""     | {"action_ext": ext}
func buildActionFromHit(domain, urlPath string, poolID int) []map[string]interface{} {
	var conditions []map[string]interface{}

	// Instance — type and value must be empty per hashResponseActionDetails.
	if poolID > 0 {
		conditions = append(conditions, map[string]interface{}{
			"type":  "",
			"value": "",
			"point": map[string]interface{}{"instance": strconv.Itoa(poolID)},
		})
	}

	// HOST header — always iequal.
	if domain != "" {
		conditions = append(conditions, map[string]interface{}{
			"type":  iequal,
			"value": domain,
			"point": map[string]interface{}{"header": "HOST"},
		})
	}

	conditions = append(conditions, locationToConditions(urlPath)...)

	return conditions
}

// locationToConditions converts a URL path into action conditions.
// Port of the Ruby LocationToConditions class.
func locationToConditions(location string) []map[string]interface{} {
	if strings.Count(location, "/") > maxPathDepth {
		return []map[string]interface{}{
			{
				"type":  "equal",
				"value": location,
				"point": map[string]interface{}{"uri": location},
			},
		}
	}

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
				"point": map[string]interface{}{"action_name": ""},
			},
			{
				"type":  "absent",
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
		"type":  "absent",
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
				"point": map[string]interface{}{"action_name": name},
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
			"point": map[string]interface{}{"action_name": segment},
		},
		{
			"type":  "absent",
			"value": "",
			"point": map[string]interface{}{"action_ext": ""},
		},
	}
}

// hashAction produces a deterministic SHA256 hex string from a sorted slice
// of action conditions.
func hashAction(action []map[string]interface{}) string {
	data, _ := json.Marshal(action)
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}
