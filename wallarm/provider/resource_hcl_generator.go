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
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

// Rule type and source constants for the HCL generator.
const (
	ruleTypeDisableStamp      = "disable_stamp"
	ruleTypeDisableAttackType = "disable_attack_type"
	generatorSourceAPI        = "api"
	generatorSourceRules      = "rules"
)

var validRuleTypes = []string{ruleTypeDisableStamp, ruleTypeDisableAttackType}

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
			"rules_json": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Description: "JSON-encoded list of pre-built rules (same structure as data.wallarm_hits rules output). " +
					"Required when source = 'rules'. Each rule must have key, resource_type, stamp/attack_type, point, and action.",
			},
			"source": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      generatorSourceRules,
				ValidateFunc: validation.StringInSlice([]string{generatorSourceRules, generatorSourceAPI}, false),
				Description:  "Source of rules: 'rules' (from pre-built rules via rules_json) or 'api' (fetch existing rules from Wallarm API).",
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

// rulesJSONAction is one action condition entry in rules_json input.
type rulesJSONAction struct {
	Type  string            `json:"type"`
	Value string            `json:"value"`
	Point map[string]string `json:"point"`
}

// rulesJSONEntry is one rule entry in rules_json input.
type rulesJSONEntry struct {
	Key          string            `json:"key"`
	ResourceType string            `json:"resource_type"`
	Stamp        int               `json:"stamp"`
	AttackType   string            `json:"attack_type"`
	Point        [][]string        `json:"point"`
	Action       []rulesJSONAction `json:"action"`
}

func generateRuleFiles(d *schema.ResourceData, clientID int, m interface{}) ([]string, int, error) {
	outputDir := d.Get("output_dir").(string)
	ruleTypes := resolveRuleTypes(d)
	movedFrom, _ := d.Get("moved_from").(string)

	// Apply defaults for Optional+Computed fields.
	source, _ := d.Get("source").(string)
	prefix := "fp"
	if source == generatorSourceAPI {
		prefix = "rule"
	} else if source == "" {
		source = generatorSourceRules
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

	switch source {
	case generatorSourceAPI:
		return generateFromAPI(m, clientID, outputDir, prefix, filename, comment, ruleTypes, split, movedFrom)
	default:
		return generateFromRulesJSON(d, clientID, outputDir, prefix, filename, comment, ruleTypes, split, movedFrom)
	}
}

// generateFromRulesJSON generates HCL from pre-built rules (rules_json input).
// Accepts the same structure as data.wallarm_hits rules output: [{key, resource_type, stamp, attack_type, point, action}].
func generateFromRulesJSON(d *schema.ResourceData, clientID int, outputDir, prefix, filename, comment string, ruleTypes []string, split bool, movedFrom string) ([]string, int, error) {
	rulesJSON := d.Get("rules_json").(string)
	if rulesJSON == "" {
		return nil, 0, fmt.Errorf("rules_json is required when source = 'rules'")
	}

	var rawRules []rulesJSONEntry
	if err := json.Unmarshal([]byte(rulesJSON), &rawRules); err != nil {
		return nil, 0, fmt.Errorf("failed to parse rules_json: %w", err)
	}

	// Build rule type filter.
	rtSet := make(map[string]bool, len(ruleTypes))
	for _, rt := range ruleTypes {
		rtSet[rt] = true
	}

	// Convert to expandedRule with per-rule action conditions.
	expanded := make([]expandedRule, 0, len(rawRules))
	for _, r := range rawRules {
		ruleType := strings.TrimPrefix(r.ResourceType, "wallarm_rule_")
		if !rtSet[ruleType] {
			continue
		}

		// Convert action format: point map → ActionCondition with correct Point/Value split.
		var ruleActions []ActionCondition
		for _, a := range r.Action {
			var point []string
			value := a.Value

			for k, v := range a.Point {
				point = append(point, k)
				if resourcerule.PointValuePoints[k] {
					value = v
				} else if v != "" {
					point = append(point, v)
				}
			}

			ruleActions = append(ruleActions, ActionCondition{
				Type:  a.Type,
				Value: value,
				Point: point,
			})
		}

		expanded = append(expanded, expandedRule{
			Key:        r.Key,
			RuleType:   ruleType,
			Point:      r.Point,
			Stamp:      r.Stamp,
			AttackType: r.AttackType,
			Actions:    ruleActions,
		})
	}

	if len(expanded) == 0 {
		return nil, 0, nil
	}

	files, err := generateStaticFiles(outputDir, prefix, filename, clientID, comment, nil, expanded, split, movedFrom)
	if err != nil {
		return nil, 0, err
	}

	return files, len(expanded), nil
}

// generateFromAPI fetches existing rules from the Wallarm API and generates HCL configs.
// Each rule becomes a standalone resource block with its point, action conditions, and rule-specific fields.
func generateFromAPI(m interface{}, clientID int, outputDir, prefix, filename, comment string, ruleTypes []string, split bool, movedFrom string) ([]string, int, error) {
	client := apiClient(m)

	// Map API rule types to internal types used by the generator.
	apiTypeMap := map[string]string{
		ruleTypeDisableStamp:      ruleTypeDisableStamp,
		ruleTypeDisableAttackType: ruleTypeDisableAttackType,
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
		case ruleTypeDisableStamp:
			er.Stamp = rule.Stamp
		case ruleTypeDisableAttackType:
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
				case ruleTypeDisableStamp:
					generateStaticDisableStamp(f, name, cfg)
				case ruleTypeDisableAttackType:
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

// pointGroup aggregates hits sharing the same point_hash.
type pointGroup struct {
	PointWrapped [][]string
	Stamps       []int
	AttackTypes  []string
}

// expandedRule is a single expanded rule ready for HCL generation.
type expandedRule struct {
	Key        string // for_each key or resource name suffix
	RuleType   string // ruleTypeDisableStamp or ruleTypeDisableAttackType
	Point      [][]string
	Stamp      int
	AttackType string
	Actions    []ActionCondition // per-rule action conditions (may differ across rules)
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
		prefix := ph[:min(16, len(ph))]

		if rtSet[ruleTypeDisableStamp] {
			for _, s := range g.Stamps {
				rules = append(rules, expandedRule{
					Key:      fmt.Sprintf("%s_%d", prefix, s),
					RuleType: ruleTypeDisableStamp,
					Point:    g.PointWrapped,
					Stamp:    s,
				})
			}
		}

		if rtSet[ruleTypeDisableAttackType] {
			for _, at := range g.AttackTypes {
				rules = append(rules, expandedRule{
					Key:        fmt.Sprintf("%s_%s", prefix, at),
					RuleType:   ruleTypeDisableAttackType,
					Point:      g.PointWrapped,
					AttackType: at,
				})
			}
		}
	}

	return rules
}

// ─── File generation ─────────────────────────────────────────────────────────────

// generateStaticFiles writes HCL resource blocks and optional moved blocks.
func generateStaticFiles(outputDir, prefix, filename string, clientID int, comment string, actions []ActionCondition, rules []expandedRule, split bool, movedFrom string) ([]string, error) {
	if !split {
		// All in one file.
		f := hclwrite.NewEmptyFile()
		for _, r := range rules {
			name := fmt.Sprintf("%s_%s", prefix, r.Key)
			ruleActions := actions
			if len(r.Actions) > 0 {
				ruleActions = r.Actions
			}
			cfg := StaticRuleConfig{
				ClientID:   clientID,
				Comment:    comment,
				Point:      r.Point,
				Actions:    ruleActions,
				Stamp:      r.Stamp,
				AttackType: r.AttackType,
			}
			resourceType := "wallarm_rule_" + r.RuleType
			switch r.RuleType {
			case ruleTypeDisableStamp:
				generateStaticDisableStamp(f, name, cfg)
			case ruleTypeDisableAttackType:
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
	files := make([]string, 0, len(rules))
	for _, r := range rules {
		f := hclwrite.NewEmptyFile()
		name := fmt.Sprintf("%s_%s", prefix, r.Key)
		ruleActions := actions
		if len(r.Actions) > 0 {
			ruleActions = r.Actions
		}
		cfg := StaticRuleConfig{
			ClientID:   clientID,
			Comment:    comment,
			Point:      r.Point,
			Actions:    ruleActions,
			Stamp:      r.Stamp,
			AttackType: r.AttackType,
		}
		resourceType := "wallarm_rule_" + r.RuleType
		switch r.RuleType {
		case ruleTypeDisableStamp:
			generateStaticDisableStamp(f, name, cfg)
		case ruleTypeDisableAttackType:
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
	return []string{ruleTypeDisableStamp, ruleTypeDisableAttackType}
}

func writeHCLFile(path string, f *hclwrite.File) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	content := hclwrite.Format(f.Bytes())
	if err := os.WriteFile(path, content, 0600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	log.Printf("[DEBUG] wallarm_rule_generator: wrote %s (%d bytes)", path, len(content))
	return nil
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}
