package wallarm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

var validRuleTypes = []string{"disable_stamp", "disable_attack_type"}

func resourceWallarmRuleGenerator() *schema.Resource {
	return &schema.Resource{
		Description:   "Generates HCL config files for Wallarm rules from hit data. State-only delete — files persist on disk.",
		CreateContext: resourceWallarmRuleGeneratorCreate,
		ReadContext:   resourceWallarmRuleGeneratorRead,
		UpdateContext: resourceWallarmRuleGeneratorUpdate,
		DeleteContext: resourceWallarmRuleGeneratorDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Client ID for generated HCL resource blocks. Defaults to provider's client_id.",
			},
			"output_dir": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Directory to write generated .tf files.",
			},
			"output_filename": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filename for the generated .tf file (used when split = false). Defaults to '{prefix}_rules.tf'.",
			},
			"requests_json": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Description: "JSON-encoded map of request_id → {hits, action_conditions} from data.wallarm_hits. " +
					"Required when source = 'hits'. Not used when source = 'api'.",
			},
			"source": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "hits",
				ValidateFunc: validation.StringInSlice([]string{"hits", "api"}, false),
				Description:  "Source of rules: 'hits' (from data.wallarm_hits via requests_json) or 'api' (fetch existing rules from Wallarm API).",
			},
			"rule_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(validRuleTypes, false),
				},
				Description: "Rule types to generate. Defaults to all supported types.",
			},
			"resource_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Prefix for resource and local names in generated HCL. Defaults to 'fp'.",
			},
			"split": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, generate one file per rule. When false, all rules in one file. Defaults to false.",
			},
			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Comment for generated rule resources. Defaults to 'Managed by Terraform'.",
			},
			"moved_from": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Resource name to generate moved blocks from. E.g. 'fp' generates: moved { from = wallarm_rule_disable_stamp.fp[\"key\"] to = wallarm_rule_disable_stamp.fp_req_key }",
			},

			// Computed outputs.
			"generated_files": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Paths of generated .tf files.",
			},
			"rules_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of generated rules.",
			},
		},
	}
}

// ─── CRUD ───────────────────────────────────────────────────────────────────────

func resourceWallarmRuleGeneratorCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := resolveGeneratorClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	files, rulesCount, err := generateRuleFiles(d, clientID, m)
	if err != nil {
		return diag.FromErr(err)
	}

	outputDir := d.Get("output_dir").(string)
	id := fmt.Sprintf("%d/%s", clientID, hashString(outputDir))
	d.SetId(id)
	d.Set("client_id", clientID)
	if err := d.Set("generated_files", files); err != nil {
		return diag.FromErr(err)
	}
	d.Set("rules_count", rulesCount)

	log.Printf("[INFO] wallarm_rule_generator: created %d files (%d rules) in %s", len(files), rulesCount, outputDir)
	return nil
}

func resourceWallarmRuleGeneratorRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	files, ok := d.Get("generated_files").([]interface{})
	if !ok || len(files) == 0 {
		return nil
	}

	// Check if all files still exist.
	allMissing := true
	for _, f := range files {
		if path, ok := f.(string); ok {
			if _, err := os.Stat(path); err == nil {
				allMissing = false
				break
			}
		}
	}

	if allMissing {
		log.Printf("[DEBUG] wallarm_rule_generator: all generated files deleted, removing from state")
		d.SetId("")
	}
	return nil
}

func resourceWallarmRuleGeneratorUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := resolveGeneratorClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	files, rulesCount, err := generateRuleFiles(d, clientID, m)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("generated_files", files); err != nil {
		return diag.FromErr(err)
	}
	d.Set("rules_count", rulesCount)

	log.Printf("[INFO] wallarm_rule_generator: updated %d files (%d rules)", len(files), rulesCount)
	return nil
}

func resourceWallarmRuleGeneratorDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] wallarm_rule_generator: removing from state (files persist on disk)")
	d.SetId("")
	return nil
}

// ─── Core pipeline ──────────────────────────────────────────────────────────────

// resolveGeneratorClientID returns client_id from schema or falls back to provider default.
func resolveGeneratorClientID(d *schema.ResourceData, m interface{}) (int, error) {
	if v, ok := d.GetOk("client_id"); ok && v.(int) > 0 {
		return v.(int), nil
	}
	meta, ok := m.(*ProviderMeta)
	if !ok || meta == nil {
		return 0, fmt.Errorf("client_id is required (not set in resource or provider)")
	}
	if meta.DefaultClientID == 0 {
		return 0, fmt.Errorf("client_id is required (provider has no default client_id)")
	}
	return meta.DefaultClientID, nil
}

// requestEntry is one entry in the requests_json map.
type requestEntry struct {
	Hits             json.RawMessage `json:"hits"`
	ActionConditions json.RawMessage `json:"action_conditions"`
	RuleTypes        []string        `json:"rule_types,omitempty"`
}

func generateRuleFiles(d *schema.ResourceData, clientID int, m interface{}) ([]string, int, error) {
	outputDir := d.Get("output_dir").(string)
	ruleTypes := resolveRuleTypes(d)
	movedFrom, _ := d.Get("moved_from").(string)

	// Apply defaults for Optional+Computed fields.
	source, _ := d.Get("source").(string)
	prefix := "fp"
	if source == "api" {
		prefix = "rule"
	}
	if v, ok := d.GetOk("resource_prefix"); ok {
		prefix = v.(string)
	}
	d.Set("resource_prefix", prefix)

	comment := "Managed by Terraform"
	if v, ok := d.GetOk("comment"); ok && v.(string) != "" {
		comment = v.(string)
	}
	d.Set("comment", comment)

	split := d.Get("split").(bool)
	d.Set("split", split)

	filename := fmt.Sprintf("%s_rules.tf", prefix)
	if v, ok := d.GetOk("output_filename"); ok && v.(string) != "" {
		filename = v.(string)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, 0, fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	if source == "" {
		source = "hits"
	}

	if source == "api" {
		return generateFromAPI(m, clientID, outputDir, prefix, filename, comment, ruleTypes, split, movedFrom)
	}

	return generateFromHits(d, clientID, outputDir, prefix, filename, comment, ruleTypes, split, movedFrom)
}

// generateFromHits generates HCL from data.wallarm_hits (requests_json input).
func generateFromHits(d *schema.ResourceData, clientID int, outputDir, prefix, filename, comment string, ruleTypes []string, split bool, movedFrom string) ([]string, int, error) {
	reqJSON := d.Get("requests_json").(string)
	if reqJSON == "" {
		return nil, 0, fmt.Errorf("requests_json is required when source = 'hits'")
	}

	var requests map[string]requestEntry
	if err := json.Unmarshal([]byte(reqJSON), &requests); err != nil {
		return nil, 0, fmt.Errorf("failed to parse requests_json: %w", err)
	}

	var allFiles []string
	totalRules := 0

	reqIDs := make([]string, 0, len(requests))
	for id := range requests {
		reqIDs = append(reqIDs, id)
	}
	sort.Strings(reqIDs)

	for _, reqID := range reqIDs {
		entry := requests[reqID]
		short := reqID[:min(8, len(reqID))]
		filePrefix := fmt.Sprintf("%s_%s", prefix, short)

		actions, err := parseActionConditionsJSON(entry.ActionConditions)
		if err != nil {
			return nil, 0, fmt.Errorf("request %s: %w", reqID, err)
		}

		groups, err := groupHitsByPoint(string(entry.Hits))
		if err != nil {
			return nil, 0, fmt.Errorf("request %s: %w", reqID, err)
		}

		effectiveRuleTypes := ruleTypes
		if len(entry.RuleTypes) > 0 {
			effectiveRuleTypes = entry.RuleTypes
		}

		expanded := expandRules(groups, effectiveRuleTypes)
		if len(expanded) == 0 {
			log.Printf("[WARN] wallarm_rule_generator: no rules for request %s", reqID)
			continue
		}

		files, err := generateStaticFiles(outputDir, filePrefix, filename, clientID, comment, actions, expanded, split, movedFrom)
		if err != nil {
			return nil, 0, fmt.Errorf("request %s: %w", reqID, err)
		}

		allFiles = append(allFiles, files...)
		totalRules += len(expanded)
	}

	return allFiles, totalRules, nil
}

// generateFromAPI fetches existing rules from the Wallarm API and generates HCL configs.
// Each rule becomes a standalone resource block with its point, action conditions, and rule-specific fields.
func generateFromAPI(m interface{}, clientID int, outputDir, prefix, filename, comment string, ruleTypes []string, split bool, movedFrom string) ([]string, int, error) {
	client := apiClient(m)

	// Map API rule types to internal types used by the generator.
	apiTypeMap := map[string]string{
		"disable_stamp":       "disable_stamp",
		"disable_attack_type": "disable_attack_type",
	}

	// Build the filter for HintRead.
	systemFalse := false
	var allRules []wallarm.ActionBody

	for page, offset := 0, 0; page < 200; page++ {
		resp, err := client.HintRead(&wallarm.HintRead{
			Limit:     APIListLimit,
			Offset:    offset,
			OrderBy:   "id",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				System:   &systemFalse,
			},
		})
		if err != nil {
			return nil, 0, fmt.Errorf("failed to fetch rules from API: %w", err)
		}
		if resp.Body == nil || len(*resp.Body) == 0 {
			break
		}
		allRules = append(allRules, *resp.Body...)
		if len(*resp.Body) < APIListLimit {
			break
		}
		offset += APIListLimit
	}

	log.Printf("[INFO] wallarm_rule_generator (api): fetched %d rules for client %d", len(allRules), clientID)

	// Filter by requested rule types and convert to expandedRule.
	rtSet := make(map[string]bool, len(ruleTypes))
	for _, rt := range ruleTypes {
		rtSet[rt] = true
	}

	// Group rules by action_id for shared action conditions.
	type actionGroup struct {
		conditions []ActionCondition
		rules      []expandedRule
	}
	groups := make(map[int]*actionGroup)

	for _, rule := range allRules {
		internalType, ok := apiTypeMap[rule.Type]
		if !ok {
			continue
		}
		if len(rtSet) > 0 && !rtSet[internalType] {
			continue
		}

		// Convert point.
		point := convertAPIPoint(rule.Point)

		// Build expanded rule.
		er := expandedRule{
			Key:      fmt.Sprintf("%d", rule.ID),
			RuleType: internalType,
			Point:    point,
		}
		switch internalType {
		case "disable_stamp":
			er.Stamp = rule.Stamp
		case "disable_attack_type":
			er.AttackType = rule.AttackType
		}

		// Get or create action group.
		ag, exists := groups[rule.ActionID]
		if !exists {
			// Parse action conditions from the rule's action field.
			var conditions []ActionCondition
			if len(rule.Action) > 0 {
				for _, ac := range rule.Action {
					point := make([]string, len(ac.Point))
					for i, p := range ac.Point {
						point[i] = fmt.Sprintf("%v", p)
					}
					c := ActionCondition{
						Type:  ac.Type,
						Point: point,
					}
					if ac.Value != nil {
						c.Value = fmt.Sprintf("%v", ac.Value)
					}
					conditions = append(conditions, c)
				}
			}
			ag = &actionGroup{conditions: conditions}
			groups[rule.ActionID] = ag
		}
		ag.rules = append(ag.rules, er)
	}

	// Sort action IDs for deterministic output.
	actionIDs := make([]int, 0, len(groups))
	for id := range groups {
		actionIDs = append(actionIDs, id)
	}
	sort.Ints(actionIDs)

	var allFiles []string
	totalRules := 0

	if !split {
		// Single file: merge all action groups into one HCL file.
		f := hclwrite.NewEmptyFile()
		for _, actionID := range actionIDs {
			ag := groups[actionID]
			for _, r := range ag.rules {
				name := fmt.Sprintf("%s_%s", prefix, r.Key)
				cfg := StaticRuleConfig{
					ClientID:   clientID,
					Comment:    comment,
					Point:      r.Point,
					Actions:    ag.conditions,
					Stamp:      r.Stamp,
					AttackType: r.AttackType,
				}
				switch r.RuleType {
				case "disable_stamp":
					generateStaticDisableStamp(f, name, cfg)
				case "disable_attack_type":
					generateStaticDisableAttackType(f, name, cfg)
				}
				totalRules++
			}
		}
		filePath := filepath.Join(outputDir, filename)
		if err := writeHCLFile(filePath, f); err != nil {
			return nil, 0, err
		}
		allFiles = append(allFiles, filePath)
	} else {
		// Split: one file per rule.
		for _, actionID := range actionIDs {
			ag := groups[actionID]
			files, err := generateStaticFiles(outputDir, prefix, filename, clientID, comment, ag.conditions, ag.rules, true, movedFrom)
			if err != nil {
				return nil, 0, fmt.Errorf("action %d: %w", actionID, err)
			}
			allFiles = append(allFiles, files...)
			totalRules += len(ag.rules)
		}
	}

	log.Printf("[INFO] wallarm_rule_generator (api): generated %d files with %d rules", len(allFiles), totalRules)
	return allFiles, totalRules, nil
}

// convertAPIPoint converts the API's flat point []interface{} to wrapped [][]string.
func convertAPIPoint(point []interface{}) [][]string {
	wrapped := resourcerule.WrapPointElements(point)
	result := make([][]string, 0, len(wrapped))
	for _, inner := range wrapped {
		row := make([]string, 0, len(inner))
		for _, v := range inner {
			row = append(row, fmt.Sprintf("%v", v))
		}
		result = append(result, row)
	}
	return result
}

// ─── Hit parsing & grouping ─────────────────────────────────────────────────────

// hitJSON is the minimal struct we parse from hits_json.
type hitJSON struct {
	Type         string          `json:"type"`
	Stamps       []int           `json:"stamps"`
	PointHash    string          `json:"point_hash"`
	PointWrapped [][]interface{} `json:"point_wrapped"`
}

// pointGroup aggregates hits sharing the same point_hash.
type pointGroup struct {
	PointWrapped [][]string
	Stamps       []int
	AttackTypes  []string
}

func groupHitsByPoint(hitsJSONStr string) (map[string]*pointGroup, error) {
	var rawHits []json.RawMessage
	if err := json.Unmarshal([]byte(hitsJSONStr), &rawHits); err != nil {
		return nil, fmt.Errorf("failed to parse hits_json: %w", err)
	}

	groups := make(map[string]*pointGroup)

	for _, raw := range rawHits {
		var h hitJSON
		if err := json.Unmarshal(raw, &h); err != nil {
			log.Printf("[WARN] wallarm_rule_generator: skipping unparseable hit: %v", err)
			continue
		}

		if h.PointHash == "" {
			continue
		}

		g, exists := groups[h.PointHash]
		if !exists {
			g = &pointGroup{
				PointWrapped: convertPointWrapped(h.PointWrapped),
			}
			groups[h.PointHash] = g
		}

		// Merge stamps.
		for _, s := range h.Stamps {
			if s > 0 && !containsInt(g.Stamps, s) {
				g.Stamps = append(g.Stamps, s)
			}
		}

		// Merge attack types.
		if h.Type != "" && !containsStr(g.AttackTypes, h.Type) {
			g.AttackTypes = append(g.AttackTypes, h.Type)
		}
	}

	// Sort for deterministic output.
	for _, g := range groups {
		sort.Ints(g.Stamps)
		sort.Strings(g.AttackTypes)
	}

	return groups, nil
}

// expandedRule is a single expanded rule ready for HCL generation.
type expandedRule struct {
	Key        string // for_each key or resource name suffix
	RuleType   string // "disable_stamp" or "disable_attack_type"
	Point      [][]string
	Stamp      int
	AttackType string
}

func expandRules(groups map[string]*pointGroup, ruleTypes []string) []expandedRule {
	var rules []expandedRule

	// Sort group keys for deterministic output.
	phKeys := make([]string, 0, len(groups))
	for ph := range groups {
		phKeys = append(phKeys, ph)
	}
	sort.Strings(phKeys)

	rtSet := make(map[string]bool, len(ruleTypes))
	for _, rt := range ruleTypes {
		rtSet[rt] = true
	}

	for _, ph := range phKeys {
		g := groups[ph]
		prefix := ph[:min(8, len(ph))]

		if rtSet["disable_stamp"] {
			for _, s := range g.Stamps {
				rules = append(rules, expandedRule{
					Key:      fmt.Sprintf("%s_%d", prefix, s),
					RuleType: "disable_stamp",
					Point:    g.PointWrapped,
					Stamp:    s,
				})
			}
		}

		if rtSet["disable_attack_type"] {
			for _, at := range g.AttackTypes {
				rules = append(rules, expandedRule{
					Key:        fmt.Sprintf("%s_%s", prefix, at),
					RuleType:   "disable_attack_type",
					Point:      g.PointWrapped,
					AttackType: at,
				})
			}
		}
	}

	return rules
}

// ─── File generation ─────────────────────────────────────────────────────────────

func generateStaticFiles(outputDir, prefix, filename string, clientID int, comment string, actions []ActionCondition, rules []expandedRule, split bool, movedFrom string) ([]string, error) {
	if !split {
		// All in one file.
		f := hclwrite.NewEmptyFile()
		for _, r := range rules {
			name := fmt.Sprintf("%s_%s", prefix, r.Key)
			cfg := StaticRuleConfig{
				ClientID:   clientID,
				Comment:    comment,
				Point:      r.Point,
				Actions:    actions,
				Stamp:      r.Stamp,
				AttackType: r.AttackType,
			}
			resourceType := "wallarm_rule_" + r.RuleType
			switch r.RuleType {
			case "disable_stamp":
				generateStaticDisableStamp(f, name, cfg)
			case "disable_attack_type":
				generateStaticDisableAttackType(f, name, cfg)
			}
			if movedFrom != "" {
				writeMovedBlock(f, resourceType, movedFrom, r.Key, name)
			}
		}
		filePath := filepath.Join(outputDir, filename)
		if err := writeHCLFile(filePath, f); err != nil {
			return nil, err
		}
		return []string{filePath}, nil
	}

	// Split: one file per rule.
	var files []string
	for _, r := range rules {
		f := hclwrite.NewEmptyFile()
		name := fmt.Sprintf("%s_%s", prefix, r.Key)
		cfg := StaticRuleConfig{
			ClientID:   clientID,
			Comment:    comment,
			Point:      r.Point,
			Actions:    actions,
			Stamp:      r.Stamp,
			AttackType: r.AttackType,
		}
		resourceType := "wallarm_rule_" + r.RuleType
		switch r.RuleType {
		case "disable_stamp":
			generateStaticDisableStamp(f, name, cfg)
		case "disable_attack_type":
			generateStaticDisableAttackType(f, name, cfg)
		}
		if movedFrom != "" {
			writeMovedBlock(f, resourceType, movedFrom, r.Key, name)
		}
		filePath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.tf", prefix, r.Key))
		if err := writeHCLFile(filePath, f); err != nil {
			return nil, err
		}
		files = append(files, filePath)
	}
	return files, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────────

func resolveRuleTypes(d *schema.ResourceData) []string {
	if v, ok := d.GetOk("rule_types"); ok {
		items := v.([]interface{})
		types := make([]string, 0, len(items))
		for _, item := range items {
			types = append(types, item.(string))
		}
		if len(types) > 0 {
			return types
		}
	}
	return []string{"disable_stamp", "disable_attack_type"}
}

func parseActionConditionsJSON(raw json.RawMessage) ([]ActionCondition, error) {
	var rawConditions []struct {
		Type  string   `json:"type"`
		Point []string `json:"point"`
		Value string   `json:"value"`
	}

	if err := json.Unmarshal(raw, &rawConditions); err != nil {
		return nil, fmt.Errorf("failed to parse action_conditions: %w", err)
	}

	conditions := make([]ActionCondition, 0, len(rawConditions))
	for _, rc := range rawConditions {
		conditions = append(conditions, ActionCondition{
			Type:  rc.Type,
			Point: rc.Point,
			Value: rc.Value,
		})
	}

	return conditions, nil
}

func convertPointWrapped(pw [][]interface{}) [][]string {
	result := make([][]string, 0, len(pw))
	for _, inner := range pw {
		row := make([]string, 0, len(inner))
		for _, v := range inner {
			row = append(row, fmt.Sprintf("%v", v))
		}
		result = append(result, row)
	}
	return result
}

func writeHCLFile(path string, f *hclwrite.File) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	content := hclwrite.Format(f.Bytes())
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	log.Printf("[DEBUG] wallarm_rule_generator: wrote %s (%d bytes)", path, len(content))
	return nil
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func containsStr(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
