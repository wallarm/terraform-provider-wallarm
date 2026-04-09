package wallarm

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

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

	files, err := generateStaticFiles(dir, "fp", "fp_rules.tf", 8649, "Managed by Terraform", actions, rules, false, "")
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
	if !strings.Contains(hcl, `client_id`) || !strings.Contains(hcl, `8649`) {
		t.Errorf("missing client_id, got:\n%s", hcl)
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

	files, err := generateStaticFiles(dir, "fp", "fp_rules.tf", 8649, "Test", actions, rules, true, "")
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

	files, err := generateStaticFiles(dir, "fp", "fp_rules.tf", 8649, "Test", actions, rules, false, "fp")
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

	files, err := generateStaticFiles(dir, "fp", "fp_rules.tf", 8649, "Test", actions, rules, true, "old_name")
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
	if !strings.Contains(instanceBlock, `type`) {
		t.Error("instance action block should have type attribute (preserved)")
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
