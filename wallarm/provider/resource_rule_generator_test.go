package wallarm

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestGroupHitsByPoint(t *testing.T) {
	hits := []map[string]interface{}{
		{
			"type":          "sqli",
			"stamps":        []int{111, 222},
			"point_hash":    "aabbccdd11223344",
			"point_wrapped": [][]string{{"header", "X-API-Key"}},
		},
		{
			"type":          "xss",
			"stamps":        []int{333},
			"point_hash":    "aabbccdd11223344", // same point
			"point_wrapped": [][]string{{"header", "X-API-Key"}},
		},
		{
			"type":          "sqli",
			"stamps":        []int{444},
			"point_hash":    "eeff00112233aabb", // different point
			"point_wrapped": [][]string{{"post"}, {"form_urlencoded", "username"}},
		},
	}

	hitsJSON, _ := json.Marshal(hits)
	groups, err := groupHitsByPoint(string(hitsJSON))
	if err != nil {
		t.Fatalf("groupHitsByPoint failed: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	// First group: merged stamps and attack types.
	g1 := groups["aabbccdd11223344"]
	if g1 == nil {
		t.Fatal("missing group for aabbccdd11223344")
	}
	if len(g1.Stamps) != 3 {
		t.Errorf("expected 3 stamps, got %d: %v", len(g1.Stamps), g1.Stamps)
	}
	if len(g1.AttackTypes) != 2 {
		t.Errorf("expected 2 attack types, got %d: %v", len(g1.AttackTypes), g1.AttackTypes)
	}

	// Second group.
	g2 := groups["eeff00112233aabb"]
	if g2 == nil {
		t.Fatal("missing group for eeff00112233aabb")
	}
	if len(g2.Stamps) != 1 || g2.Stamps[0] != 444 {
		t.Errorf("expected stamps [444], got %v", g2.Stamps)
	}
}

func TestGroupHitsByPoint_DuplicateStamps(t *testing.T) {
	hits := []map[string]interface{}{
		{"type": "sqli", "stamps": []int{111, 222}, "point_hash": "aabb", "point_wrapped": [][]string{{"post"}}},
		{"type": "sqli", "stamps": []int{111, 333}, "point_hash": "aabb", "point_wrapped": [][]string{{"post"}}},
	}

	hitsJSON, _ := json.Marshal(hits)
	groups, err := groupHitsByPoint(string(hitsJSON))
	if err != nil {
		t.Fatalf("groupHitsByPoint failed: %v", err)
	}

	g := groups["aabb"]
	if len(g.Stamps) != 3 {
		t.Errorf("expected 3 deduplicated stamps, got %d: %v", len(g.Stamps), g.Stamps)
	}
	// Should be sorted.
	if g.Stamps[0] != 111 || g.Stamps[1] != 222 || g.Stamps[2] != 333 {
		t.Errorf("stamps not sorted: %v", g.Stamps)
	}
}

func TestGroupHitsByPoint_EmptyInput(t *testing.T) {
	groups, err := groupHitsByPoint("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestGroupHitsByPoint_SkipsEmptyPointHash(t *testing.T) {
	hits := []map[string]interface{}{
		{"type": "sqli", "stamps": []int{111}, "point_hash": "", "point_wrapped": [][]string{{"post"}}},
		{"type": "sqli", "stamps": []int{222}, "point_hash": "aabb", "point_wrapped": [][]string{{"post"}}},
	}

	hitsJSON, _ := json.Marshal(hits)
	groups, err := groupHitsByPoint(string(hitsJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group (empty point_hash skipped), got %d", len(groups))
	}
}

func TestExpandRules(t *testing.T) {
	groups := map[string]*pointGroup{
		"aabbccdd11223344": {
			PointWrapped: [][]string{{"header", "X-API-Key"}},
			Stamps:       []int{111, 222},
			AttackTypes:  []string{"sqli", "xss"},
		},
	}

	// Both types.
	rules := expandRules(groups, []string{"disable_stamp", "disable_attack_type"})
	if len(rules) != 4 {
		t.Fatalf("expected 4 rules (2 stamps + 2 attack_types), got %d", len(rules))
	}

	// Stamps only.
	rules = expandRules(groups, []string{"disable_stamp"})
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules (stamps only), got %d", len(rules))
	}

	// Attack types only.
	rules = expandRules(groups, []string{"disable_attack_type"})
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules (attack_types only), got %d", len(rules))
	}
}

func TestExpandRules_EmptyGroups(t *testing.T) {
	groups := map[string]*pointGroup{}
	rules := expandRules(groups, []string{"disable_stamp", "disable_attack_type"})
	if len(rules) != 0 {
		t.Errorf("expected 0 rules from empty groups, got %d", len(rules))
	}
}

func TestExpandRules_NoStamps(t *testing.T) {
	groups := map[string]*pointGroup{
		"aabb": {
			PointWrapped: [][]string{{"post"}},
			Stamps:       nil,
			AttackTypes:  []string{"sqli"},
		},
	}
	rules := expandRules(groups, []string{"disable_stamp", "disable_attack_type"})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule (attack_type only, no stamps), got %d", len(rules))
	}
	if rules[0].RuleType != "disable_attack_type" {
		t.Errorf("expected disable_attack_type, got %s", rules[0].RuleType)
	}
}

func TestGenerateStaticHCL(t *testing.T) {
	dir := t.TempDir()

	actions := []ActionCondition{
		{Type: "iequal", Point: []string{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []string{"path", "0"}, Value: "api"},
	}

	rules := []expandedRule{
		{Key: "aabbccdd_111", RuleType: "disable_stamp", Point: [][]string{{"header", "X-API-Key"}}, Stamp: 111},
		{Key: "aabbccdd_sqli", RuleType: "disable_attack_type", Point: [][]string{{"header", "X-API-Key"}}, AttackType: "sqli"},
	}

	files, err := generateStaticFiles(dir, "fp", 8649, "Managed by Terraform", actions, rules, false, "")
	if err != nil {
		t.Fatalf("generateStaticFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	hcl := string(content)

	// Check resource blocks exist.
	if !strings.Contains(hcl, `resource "wallarm_rule_disable_stamp" "fp_aabbccdd_111"`) {
		t.Error("missing disable_stamp resource block")
	}
	if !strings.Contains(hcl, `resource "wallarm_rule_disable_attack_type" "fp_aabbccdd_sqli"`) {
		t.Error("missing disable_attack_type resource block")
	}
	if !strings.Contains(hcl, `client_id = 8649`) {
		t.Error("missing client_id")
	}
	if !strings.Contains(hcl, "stamp") || !strings.Contains(hcl, "111") {
		t.Error("missing stamp value")
	}
	if !strings.Contains(hcl, `attack_type`) || !strings.Contains(hcl, `"sqli"`) {
		t.Error("missing attack_type value")
	}
	if !strings.Contains(hcl, `header = "HOST"`) {
		t.Error("missing action block with header HOST")
	}
	// No moved blocks when movedFrom is empty.
	if strings.Contains(hcl, "moved") {
		t.Error("unexpected moved block when movedFrom is empty")
	}
}

func TestSplitStaticFiles(t *testing.T) {
	dir := t.TempDir()

	actions := []ActionCondition{
		{Type: "iequal", Point: []string{"header", "HOST"}, Value: "example.com"},
	}

	rules := []expandedRule{
		{Key: "aabb_111", RuleType: "disable_stamp", Point: [][]string{{"header", "X-Key"}}, Stamp: 111},
		{Key: "aabb_sqli", RuleType: "disable_attack_type", Point: [][]string{{"header", "X-Key"}}, AttackType: "sqli"},
	}

	files, err := generateStaticFiles(dir, "fp", 8649, "Test", actions, rules, true, "")
	if err != nil {
		t.Fatalf("generateStaticFiles split failed: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files (split), got %d", len(files))
	}

	// Verify each file contains only one resource.
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		count := strings.Count(string(content), "resource \"wallarm_rule_")
		if count != 1 {
			t.Errorf("file %s has %d resource blocks, expected 1", f, count)
		}
	}
}

func TestGenerateStaticWithMovedBlocks(t *testing.T) {
	dir := t.TempDir()

	actions := []ActionCondition{
		{Type: "iequal", Point: []string{"header", "HOST"}, Value: "example.com"},
	}

	rules := []expandedRule{
		{Key: "aabb_111", RuleType: "disable_stamp", Point: [][]string{{"header", "X-Key"}}, Stamp: 111},
		{Key: "aabb_sqli", RuleType: "disable_attack_type", Point: [][]string{{"header", "X-Key"}}, AttackType: "sqli"},
	}

	files, err := generateStaticFiles(dir, "fp", 8649, "Test", actions, rules, false, "fp")
	if err != nil {
		t.Fatalf("generateStaticFiles with moved failed: %v", err)
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	hcl := string(content)

	// Check moved blocks exist.
	if !strings.Contains(hcl, "moved {") {
		t.Fatal("missing moved block")
	}

	// Check from address uses for_each key format.
	if !strings.Contains(hcl, `wallarm_rule_disable_stamp.fp["aabb_111"]`) {
		t.Errorf("missing moved from address for disable_stamp, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `wallarm_rule_disable_attack_type.fp["aabb_sqli"]`) {
		t.Errorf("missing moved from address for disable_attack_type, got:\n%s", hcl)
	}

	// Check to address uses static name format.
	if !strings.Contains(hcl, "wallarm_rule_disable_stamp.fp_aabb_111") {
		t.Error("missing moved to address for disable_stamp")
	}
	if !strings.Contains(hcl, "wallarm_rule_disable_attack_type.fp_aabb_sqli") {
		t.Error("missing moved to address for disable_attack_type")
	}

	// Count: 2 resources + 2 moved blocks.
	if count := strings.Count(hcl, "moved {"); count != 2 {
		t.Errorf("expected 2 moved blocks, got %d", count)
	}
}

func TestGenerateStaticWithMovedBlocks_Split(t *testing.T) {
	dir := t.TempDir()

	actions := []ActionCondition{
		{Type: "iequal", Point: []string{"header", "HOST"}, Value: "example.com"},
	}

	rules := []expandedRule{
		{Key: "aabb_111", RuleType: "disable_stamp", Point: [][]string{{"header", "X-Key"}}, Stamp: 111},
	}

	files, err := generateStaticFiles(dir, "fp", 8649, "Test", actions, rules, true, "old_name")
	if err != nil {
		t.Fatalf("generateStaticFiles split+moved failed: %v", err)
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	hcl := string(content)

	// Each split file should have 1 resource + 1 moved block.
	if count := strings.Count(hcl, `resource "wallarm_rule_`); count != 1 {
		t.Errorf("expected 1 resource block, got %d", count)
	}
	if count := strings.Count(hcl, "moved {"); count != 1 {
		t.Errorf("expected 1 moved block, got %d", count)
	}
	if !strings.Contains(hcl, `wallarm_rule_disable_stamp.old_name["aabb_111"]`) {
		t.Errorf("moved from should use old_name prefix, got:\n%s", hcl)
	}
}

func TestWriteActionBlocks_PointValueTypes(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	block := f.Body().AppendNewBlock("resource", []string{"test", "foo"})

	conditions := []ActionCondition{
		// Instance: type="" (omitted), value="" (omitted), point={instance="13"}
		{Type: "equal", Point: []string{"instance"}, Value: "13"},
		// Header: type="iequal", value="example.com", point={header="HOST"}
		{Type: "iequal", Point: []string{"header", "HOST"}, Value: "example.com"},
		// Action name: type="equal", value="" (omitted), point={action_name="login"}
		{Type: "equal", Point: []string{"action_name"}, Value: "login"},
		// Path: type="equal", value="api", point={path="0"}
		{Type: "equal", Point: []string{"path", "0"}, Value: "api"},
		// Absent path: type="absent", value="" (omitted), point={path="1"}
		{Type: "absent", Point: []string{"path", "1"}, Value: ""},
	}

	writeActionBlocks(block.Body(), conditions)
	hcl := string(hclwrite.Format(f.Bytes()))

	// Instance block: should have no type, no value — just point.
	// Verify the instance action block has no type attribute.
	instanceIdx := strings.Index(hcl, `instance = "13"`)
	if instanceIdx < 0 {
		t.Fatal("missing instance point value")
	}
	// Extract the action block containing instance.
	instanceBlock := hcl[strings.LastIndex(hcl[:instanceIdx], "action {"):instanceIdx]
	if strings.Contains(instanceBlock, "type") {
		t.Error("instance action block should not have type attribute")
	}
	if strings.Contains(instanceBlock, "value") {
		t.Error("instance action block should not have value attribute")
	}

	// Action name: value should be in point map, not in value field.
	if !strings.Contains(hcl, `action_name = "login"`) {
		t.Error("missing action_name in point map")
	}

	// Header: type and value present.
	if !strings.Contains(hcl, `"iequal"`) {
		t.Error("missing iequal type for header")
	}
	if !strings.Contains(hcl, `"example.com"`) {
		t.Error("missing value for header")
	}

	// Path with value.
	if !strings.Contains(hcl, `"api"`) {
		t.Error("missing value for path")
	}

	// Absent: type present, no value.
	if !strings.Contains(hcl, `"absent"`) {
		t.Error("missing absent type")
	}
}

func TestWriteMovedBlock(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	writeMovedBlock(f, "wallarm_rule_disable_stamp", "fp", "aabb_111", "fp_req1_aabb_111")

	hcl := string(hclwrite.Format(f.Bytes()))

	if !strings.Contains(hcl, "moved {") {
		t.Fatal("missing moved block")
	}
	if !strings.Contains(hcl, `from = wallarm_rule_disable_stamp.fp["aabb_111"]`) {
		t.Errorf("wrong from address, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `to   = wallarm_rule_disable_stamp.fp_req1_aabb_111`) {
		// hclwrite may format with different spacing
		if !strings.Contains(hcl, `to = wallarm_rule_disable_stamp.fp_req1_aabb_111`) {
			t.Errorf("wrong to address, got:\n%s", hcl)
		}
	}
}

func TestParseActionConditionsJSON(t *testing.T) {
	raw := []byte(`[
		{"type": "iequal", "point": ["header", "HOST"], "value": "example.com"},
		{"type": "equal", "point": ["instance"], "value": "13"},
		{"type": "absent", "point": ["path", "1"], "value": ""}
	]`)

	conditions, err := parseActionConditionsJSON(raw)
	if err != nil {
		t.Fatalf("parseActionConditionsJSON failed: %v", err)
	}

	if len(conditions) != 3 {
		t.Fatalf("expected 3 conditions, got %d", len(conditions))
	}

	if conditions[0].Type != "iequal" || conditions[0].Value != "example.com" {
		t.Errorf("wrong first condition: %+v", conditions[0])
	}
	if len(conditions[0].Point) != 2 || conditions[0].Point[0] != "header" {
		t.Errorf("wrong first condition point: %v", conditions[0].Point)
	}

	if conditions[1].Type != "equal" || conditions[1].Value != "13" {
		t.Errorf("wrong instance condition: %+v", conditions[1])
	}

	if conditions[2].Type != "absent" || conditions[2].Value != "" {
		t.Errorf("wrong absent condition: %+v", conditions[2])
	}
}
