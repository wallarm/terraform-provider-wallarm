package wallarm

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

const maxPathDepth = 10

func dataSourceWallarmHits() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWallarmHitsRead,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"request_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique request identifier to fetch all related hits",
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

func dataSourceWallarmHitsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	requestID := d.Get("request_id").(string)

	var timeRange [][]interface{}
	if v, ok := d.GetOk("time"); ok {
		tl := v.([]interface{})
		if len(tl) == 2 {
			timeRange = [][]interface{}{{tl[0], tl[1]}}
		}
	}
	if len(timeRange) == 0 {
		sixMonthsAgo := time.Now().AddDate(0, -6, 0).Unix()
		now := time.Now().Unix()
		timeRange = [][]interface{}{{sixMonthsAgo, now}}
	}

	req := &wallarm.HitReadRequest{
		Filter: &wallarm.HitFilter{
			ClientID:          clientID,
			RequestID:         requestID,
			State:             nil,
			NotType:           []string{"warn", "infoleak"},
			Time:              timeRange,
			NotState:          "falsepositive",
			SecurityIssueID:   nil,
			NotExperimental:   true,
			NotAasmEvent:      true,
			NotWallarmScanner: true,
		},
		Limit:     200,
		Offset:    0,
		OrderBy:   "time",
		OrderDesc: true,
	}

	resp, err := client.HitRead(req)
	if err != nil {
		return fmt.Errorf("error reading hits for request_id %s: %s", requestID, err)
	}

	// Stable, deterministic ID — does not change between plans.
	d.SetId(fmt.Sprintf("hits_%d_%s", clientID, requestID))

	if len(resp) == 0 {
		_ = d.Set("action", schema.NewSet(schema.HashResource(
			defaultResourceRuleActionSchema.Elem.(*schema.Resource)), []interface{}{}))
		_ = d.Set("action_hash", "")
		_ = d.Set("hits", []interface{}{})
		return nil
	}

	// Validate all hits share the same domain/path/poolid.
	refDomain := resp[0].Domain
	refPath := resp[0].Path
	refPoolID := resp[0].PoolID
	for _, h := range resp[1:] {
		if h.Domain != refDomain || h.Path != refPath || h.PoolID != refPoolID {
			return fmt.Errorf(
				"inconsistent hit data for request_id %s: expected domain=%s path=%s poolid=%d, got domain=%s path=%s poolid=%d",
				requestID, refDomain, refPath, refPoolID, h.Domain, h.Path, h.PoolID,
			)
		}
	}

	action := buildActionFromHit(refDomain, refPath, refPoolID)

	sortedAction := make([]map[string]interface{}, len(action))
	copy(sortedAction, action)
	sort.Slice(sortedAction, func(i, j int) bool {
		return fmt.Sprintf("%v", sortedAction[i]["point"]) < fmt.Sprintf("%v", sortedAction[j]["point"])
	})
	actionHash := hashAction(sortedAction)

	actionIfaces := make([]interface{}, len(action))
	for i, a := range action {
		actionIfaces[i] = a
	}
	actionSet := schema.NewSet(
		schema.HashResource(defaultResourceRuleActionSchema.Elem.(*schema.Resource)),
		actionIfaces,
	)

	hits := make([]interface{}, 0, len(resp))
	for _, h := range resp {
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

		hits = append(hits, map[string]interface{}{
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

	if err := d.Set("action", actionSet); err != nil {
		return fmt.Errorf("error setting action: %s", err)
	}
	if err := d.Set("action_hash", actionHash); err != nil {
		return fmt.Errorf("error setting action_hash: %s", err)
	}
	if err := d.Set("hits", hits); err != nil {
		return fmt.Errorf("error setting hits: %s", err)
	}

	return nil
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

	var conditions []map[string]interface{}

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
