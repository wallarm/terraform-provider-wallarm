package wallarm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestGenerateFromRulesJSON_MultipleActionScopes verifies that rules with different
// action scopes get their own correct action conditions in generated HCL.
func TestGenerateFromRulesJSON_MultipleActionScopes(t *testing.T) {
	// Build _all_rules similar to how the HCL module does it.
	allRulesJSON := `[
		{"key": "aaaa1111bbbb2222_cccc3333dddd4444_7994", "resource_type": "wallarm_rule_disable_stamp", "stamp": 7994, "attack_type": "",
		 "point": [["get", "search"]],
		 "action": [
			{"type": "", "value": "", "point": {"instance": "5"}},
			{"type": "iequal", "value": "example.com", "point": {"header": "HOST"}},
			{"type": "equal", "value": "", "point": {"action_name": "users"}},
			{"type": "absent", "value": "", "point": {"action_ext": ""}},
			{"type": "equal", "value": "api", "point": {"path": "0"}},
			{"type": "equal", "value": "v1", "point": {"path": "1"}},
			{"type": "absent", "value": "", "point": {"path": "2"}}
		 ]},
		{"key": "eeee5555ffff6666_dddd7777eeee8888_1234", "resource_type": "wallarm_rule_disable_stamp", "stamp": 1234, "attack_type": "",
		 "point": [["post"], ["form_urlencoded", "password"]],
		 "action": [
			{"type": "", "value": "", "point": {"instance": "5"}},
			{"type": "iequal", "value": "admin.example.com", "point": {"header": "HOST"}},
			{"type": "equal", "value": "", "point": {"action_name": "login"}},
			{"type": "absent", "value": "", "point": {"action_ext": ""}},
			{"type": "absent", "value": "", "point": {"path": "0"}}
		 ]},
		{"key": "eeee5555ffff6666_dddd7777eeee8888_sqli", "resource_type": "wallarm_rule_disable_attack_type", "stamp": 0, "attack_type": "sqli",
		 "point": [["post"], ["form_urlencoded", "password"]],
		 "action": [
			{"type": "", "value": "", "point": {"instance": "5"}},
			{"type": "iequal", "value": "admin.example.com", "point": {"header": "HOST"}},
			{"type": "equal", "value": "", "point": {"action_name": "login"}},
			{"type": "absent", "value": "", "point": {"action_ext": ""}},
			{"type": "absent", "value": "", "point": {"path": "0"}}
		 ]}
	]`

	dir := t.TempDir()

	r := resourceWallarmRuleGenerator()
	d := newTestResourceData(t, r, map[string]interface{}{
		"output_dir": dir,
		"rules_json": allRulesJSON,
		"source":     "rules",
		"split":      true,
	})

	files, rulesCount, err := generateRuleFiles(d, 8649, nil)
	if err != nil {
		t.Fatalf("generateRuleFiles failed: %v", err)
	}

	if rulesCount != 3 {
		t.Errorf("expected 3 rules, got %d", rulesCount)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}

	// Verify scope 1 rule has scope 1 action (example.com, action_name=users).
	scope1File := filepath.Join(dir, "fp_aaaa1111bbbb2222_cccc3333dddd4444_7994.tf")
	scope1Content := readFileStr(t, scope1File)
	if !strings.Contains(scope1Content, `value = "example.com"`) {
		t.Errorf("scope 1 rule missing example.com in action:\n%s", scope1Content)
	}
	if !strings.Contains(scope1Content, `action_name = "users"`) {
		t.Errorf("scope 1 rule missing action_name=users:\n%s", scope1Content)
	}
	if strings.Contains(scope1Content, `admin.example.com`) {
		t.Errorf("scope 1 rule should NOT have admin.example.com:\n%s", scope1Content)
	}

	// Verify scope 2 stamp rule has scope 2 action (admin.example.com, action_name=login).
	scope2StampFile := filepath.Join(dir, "fp_eeee5555ffff6666_dddd7777eeee8888_1234.tf")
	scope2StampContent := readFileStr(t, scope2StampFile)
	if !strings.Contains(scope2StampContent, `value = "admin.example.com"`) {
		t.Errorf("scope 2 stamp rule missing admin.example.com:\n%s", scope2StampContent)
	}
	if !strings.Contains(scope2StampContent, `action_name = "login"`) {
		t.Errorf("scope 2 stamp rule missing action_name=login:\n%s", scope2StampContent)
	}

	// Verify scope 2 attack_type rule also has scope 2 action.
	scope2ATFile := filepath.Join(dir, "fp_eeee5555ffff6666_dddd7777eeee8888_sqli.tf")
	scope2ATContent := readFileStr(t, scope2ATFile)
	if !strings.Contains(scope2ATContent, `value = "admin.example.com"`) {
		t.Errorf("scope 2 attack_type rule missing admin.example.com:\n%s", scope2ATContent)
	}
	if !strings.Contains(scope2ATContent, `"sqli"`) {
		t.Errorf("scope 2 attack_type rule missing sqli:\n%s", scope2ATContent)
	}
}

// TestGenerateFromRulesJSON_ActionConditionFormat verifies point-value types,
// absent conditions, and iequal are generated correctly.
func TestGenerateFromRulesJSON_ActionConditionFormat(t *testing.T) {
	rulesJSON := `[
		{"key": "aaaa1111bbbb2222_cccc3333dddd4444_7994", "resource_type": "wallarm_rule_disable_stamp", "stamp": 7994, "attack_type": "",
		 "point": [["get", "search"]],
		 "action": [
			{"type": "", "value": "", "point": {"instance": "5"}},
			{"type": "iequal", "value": "example.com", "point": {"header": "HOST"}},
			{"type": "equal", "value": "", "point": {"action_name": "users"}},
			{"type": "absent", "value": "", "point": {"action_ext": ""}},
			{"type": "equal", "value": "api", "point": {"path": "0"}},
			{"type": "equal", "value": "v1", "point": {"path": "1"}},
			{"type": "absent", "value": "", "point": {"path": "2"}}
		 ]}
	]`

	dir := t.TempDir()

	r := resourceWallarmRuleGenerator()
	d := newTestResourceData(t, r, map[string]interface{}{
		"output_dir": dir,
		"rules_json": rulesJSON,
		"source":     "rules",
		"split":      true,
	})

	_, _, err := generateRuleFiles(d, 8649, nil)
	if err != nil {
		t.Fatalf("generateRuleFiles failed: %v", err)
	}

	content := readFileStr(t, filepath.Join(dir, "fp_aaaa1111bbbb2222_cccc3333dddd4444_7994.tf"))

	// Instance: point-value type, no type/value attributes in HCL.
	if !strings.Contains(content, `instance = "5"`) {
		t.Errorf("missing instance=5:\n%s", content)
	}

	// HOST header: iequal with value.
	if !strings.Contains(content, `type  = "iequal"`) {
		t.Errorf("missing iequal type:\n%s", content)
	}
	if !strings.Contains(content, `value = "example.com"`) {
		t.Errorf("missing example.com value:\n%s", content)
	}

	// action_name: point-value type, value in point map.
	if !strings.Contains(content, `action_name = "users"`) {
		t.Errorf("missing action_name=users:\n%s", content)
	}

	// action_ext: absent condition.
	if !strings.Contains(content, `type = "absent"`) {
		t.Errorf("missing absent type:\n%s", content)
	}

	// Path segments.
	if !strings.Contains(content, `value = "api"`) {
		t.Errorf("missing path value api:\n%s", content)
	}
	if !strings.Contains(content, `value = "v1"`) {
		t.Errorf("missing path value v1:\n%s", content)
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func readFileStr(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

// newTestResourceData creates a ResourceData from a schema and raw values,
// suitable for calling generateRuleFiles directly.
func newTestResourceData(t *testing.T, r *schema.Resource, values map[string]interface{}) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, r.Schema, values)
}
