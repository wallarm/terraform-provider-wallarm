package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	wallarm "github.com/wallarm/wallarm-go"
)

// --- Unit tests for pure transformation functions (no API access) ---

func TestBuildActionFromHit_BasicWithInstance(t *testing.T) {
	result := buildActionFromHit("example.com", "/api/v1/users", 42, true)

	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// Check instance condition is present.
	found := false
	for _, c := range result {
		pm, _ := c["point"].(map[string]interface{})
		if _, ok := pm["instance"]; ok {
			found = true
			if pm["instance"] != "42" {
				t.Errorf("expected instance=42, got %v", pm["instance"])
			}
		}
	}
	if !found {
		t.Error("expected instance condition when includeInstance=true")
	}

	// Check HOST header is present with iequal.
	found = false
	for _, c := range result {
		pm, _ := c["point"].(map[string]interface{})
		if v, ok := pm["header"]; ok && v == "HOST" {
			found = true
			if c["type"] != "iequal" {
				t.Errorf("expected type=iequal for HOST header, got %v", c["type"])
			}
			if c["value"] != "example.com" {
				t.Errorf("expected value=example.com, got %v", c["value"])
			}
		}
	}
	if !found {
		t.Error("expected HOST header condition")
	}
}

func TestBuildActionFromHit_WithoutInstance(t *testing.T) {
	result := buildActionFromHit("example.com", "/api", 42, false)

	for _, c := range result {
		pm, _ := c["point"].(map[string]interface{})
		if _, ok := pm["instance"]; ok {
			t.Error("expected no instance condition when includeInstance=false")
		}
	}
}

func TestBuildActionFromHit_RootPath(t *testing.T) {
	result := buildActionFromHit("example.com", "/", 1, false)

	hasActionName := false
	hasAbsentPath := false
	for _, c := range result {
		pm, _ := c["point"].(map[string]interface{})
		if _, ok := pm["action_name"]; ok {
			hasActionName = true
			if c["type"] != "equal" {
				t.Errorf("expected type=equal for action_name, got %v", c["type"])
			}
		}
		if v, ok := pm["path"]; ok && v == "0" && c["type"] == "absent" {
			hasAbsentPath = true
		}
	}
	if !hasActionName {
		t.Error("expected action_name condition for root path")
	}
	if !hasAbsentPath {
		t.Error("expected absent path[0] condition for root path")
	}
}

func TestBuildActionFromHit_PathWithExtension(t *testing.T) {
	result := buildActionFromHit("example.com", "/api/v1/users.json", 1, false)

	hasActionName := false
	hasActionExt := false
	for _, c := range result {
		pm, _ := c["point"].(map[string]interface{})
		if v, ok := pm["action_name"]; ok && v == "users" {
			hasActionName = true
		}
		if v, ok := pm["action_ext"]; ok && v == "json" {
			hasActionExt = true
			if c["type"] != "equal" {
				t.Errorf("expected type=equal for action_ext, got %v", c["type"])
			}
		}
	}
	if !hasActionName {
		t.Error("expected action_name=users")
	}
	if !hasActionExt {
		t.Error("expected action_ext=json")
	}
}

func TestBuildActionFromHit_EmptyDomain(t *testing.T) {
	result := buildActionFromHit("", "/api", 0, true)

	for _, c := range result {
		pm, _ := c["point"].(map[string]interface{})
		if _, ok := pm["header"]; ok {
			t.Error("expected no HOST header condition when domain is empty")
		}
		if _, ok := pm["instance"]; ok {
			t.Error("expected no instance condition when poolID=0")
		}
	}
}

func TestLocationToConditions_RootPath(t *testing.T) {
	result := locationToConditions("/")

	// "/" splits to [""], so last="" with no path segments.
	// Produces: action_name="", absent(action_ext), absent(path[0]).
	if len(result) != 3 {
		t.Fatalf("expected 3 conditions for root path, got %d", len(result))
	}

	pm0, _ := result[0]["point"].(map[string]interface{})
	if _, ok := pm0["action_name"]; !ok {
		t.Error("expected action_name in first condition")
	}
	if result[0]["type"] != "equal" {
		t.Errorf("expected type=equal for action_name, got %v", result[0]["type"])
	}

	pm1, _ := result[1]["point"].(map[string]interface{})
	if _, ok := pm1["action_ext"]; !ok {
		t.Error("expected action_ext in second condition")
	}
	if result[1]["type"] != "absent" {
		t.Errorf("expected type=absent for action_ext, got %v", result[1]["type"])
	}

	pm2, _ := result[2]["point"].(map[string]interface{})
	if v, ok := pm2["path"]; !ok || v != "0" {
		t.Error("expected absent path[0] in third condition")
	}
	if result[2]["type"] != "absent" {
		t.Errorf("expected type=absent for path, got %v", result[2]["type"])
	}
}

func TestLocationToConditions_SimplePath(t *testing.T) {
	result := locationToConditions("/api")

	// /api -> action_name=api, absent(action_ext), absent(path 0)
	if len(result) != 3 {
		t.Fatalf("expected 3 conditions for /api, got %d", len(result))
	}

	pm0, _ := result[0]["point"].(map[string]interface{})
	if v := pm0["action_name"]; v != "api" {
		t.Errorf("expected action_name=api, got %v", v)
	}

	pm1, _ := result[1]["point"].(map[string]interface{})
	if _, ok := pm1["action_ext"]; !ok {
		t.Error("expected absent action_ext condition")
	}
	if result[1]["type"] != "absent" {
		t.Errorf("expected type=absent for action_ext, got %v", result[1]["type"])
	}

	if result[2]["type"] != "absent" {
		t.Errorf("expected type=absent for terminating path, got %v", result[2]["type"])
	}
}

func TestLocationToConditions_MultiSegment(t *testing.T) {
	result := locationToConditions("/api/v1/users")

	// Expect: action_name=users, absent(action_ext), path[0]=api, path[1]=v1, absent(path[2])
	if len(result) != 5 {
		t.Fatalf("expected 5 conditions for /api/v1/users, got %d", len(result))
	}

	pm0, _ := result[0]["point"].(map[string]interface{})
	if v := pm0["action_name"]; v != "users" {
		t.Errorf("expected action_name=users, got %v", v)
	}

	pm2, _ := result[2]["point"].(map[string]interface{})
	if v := pm2["path"]; v != "0" {
		t.Errorf("expected path index 0, got %v", v)
	}
	if result[2]["value"] != "api" {
		t.Errorf("expected path value=api, got %v", result[2]["value"])
	}

	pm3, _ := result[3]["point"].(map[string]interface{})
	if v := pm3["path"]; v != "1" {
		t.Errorf("expected path index 1, got %v", v)
	}
	if result[3]["value"] != "v1" {
		t.Errorf("expected path value=v1, got %v", result[3]["value"])
	}

	if result[4]["type"] != "absent" {
		t.Errorf("expected absent for terminating path, got %v", result[4]["type"])
	}
}

func TestLocationToConditions_PathWithExtension(t *testing.T) {
	result := locationToConditions("/api/data.json")

	// Expect: action_name=data, action_ext=json, path[0]=api, absent(path[1])
	if len(result) != 4 {
		t.Fatalf("expected 4 conditions for /api/data.json, got %d", len(result))
	}

	pm0, _ := result[0]["point"].(map[string]interface{})
	if v := pm0["action_name"]; v != "data" {
		t.Errorf("expected action_name=data, got %v", v)
	}
	pm1, _ := result[1]["point"].(map[string]interface{})
	if v := pm1["action_ext"]; v != "json" {
		t.Errorf("expected action_ext=json, got %v", v)
	}
}

func TestLocationToConditions_PathNoExtension(t *testing.T) {
	result := locationToConditions("/api/login")

	// Expect: action_name=login, absent(action_ext), path[0]=api, absent(path[1])
	if len(result) != 4 {
		t.Fatalf("expected 4 conditions for /api/login, got %d", len(result))
	}

	pm0, _ := result[0]["point"].(map[string]interface{})
	if v := pm0["action_name"]; v != "login" {
		t.Errorf("expected action_name=login, got %v", v)
	}
	if result[1]["type"] != "absent" {
		t.Errorf("expected type=absent for action_ext, got %v", result[1]["type"])
	}
}

func TestActionNameExtConditions_WithExtension(t *testing.T) {
	result := actionNameExtConditions("file.json")

	if len(result) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(result))
	}

	pm0, _ := result[0]["point"].(map[string]interface{})
	if v := pm0["action_name"]; v != "file" {
		t.Errorf("expected action_name=file, got %v", v)
	}
	if result[0]["type"] != "equal" {
		t.Errorf("expected type=equal for action_name, got %v", result[0]["type"])
	}

	pm1, _ := result[1]["point"].(map[string]interface{})
	if v := pm1["action_ext"]; v != "json" {
		t.Errorf("expected action_ext=json, got %v", v)
	}
	if result[1]["type"] != "equal" {
		t.Errorf("expected type=equal for action_ext, got %v", result[1]["type"])
	}
}

func TestActionNameExtConditions_WithoutExtension(t *testing.T) {
	result := actionNameExtConditions("login")

	if len(result) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(result))
	}

	pm0, _ := result[0]["point"].(map[string]interface{})
	if v := pm0["action_name"]; v != "login" {
		t.Errorf("expected action_name=login, got %v", v)
	}

	pm1, _ := result[1]["point"].(map[string]interface{})
	if _, ok := pm1["action_ext"]; !ok {
		t.Error("expected action_ext key in second condition")
	}
	if result[1]["type"] != "absent" {
		t.Errorf("expected type=absent for action_ext, got %v", result[1]["type"])
	}
}

func TestMergeHits_NoOverlap(t *testing.T) {
	direct := []*wallarm.Hit{
		{ID: []string{"a", "1"}, Type: "sqli"},
		{ID: []string{"a", "2"}, Type: "xss"},
	}
	related := []*wallarm.Hit{
		{ID: []string{"b", "3"}, Type: "rce"},
	}

	merged := mergeHits(direct, related)
	if len(merged) != 3 {
		t.Fatalf("expected 3 merged hits, got %d", len(merged))
	}
}

func TestMergeHits_WithDuplicates(t *testing.T) {
	direct := []*wallarm.Hit{
		{ID: []string{"a", "1"}, Type: "sqli"},
		{ID: []string{"a", "2"}, Type: "xss"},
	}
	related := []*wallarm.Hit{
		{ID: []string{"a", "1"}, Type: "sqli"}, // duplicate
		{ID: []string{"b", "3"}, Type: "rce"},
	}

	merged := mergeHits(direct, related)
	if len(merged) != 3 {
		t.Fatalf("expected 3 merged hits (1 dedup), got %d", len(merged))
	}
}

func TestMergeHits_EmptyRelated(t *testing.T) {
	direct := []*wallarm.Hit{
		{ID: []string{"a", "1"}, Type: "sqli"},
	}

	merged := mergeHits(direct, nil)
	if len(merged) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(merged))
	}
	if merged[0].Type != "sqli" {
		t.Errorf("expected type=sqli, got %s", merged[0].Type)
	}
}

func TestSchemaActionToDetails_HeaderCondition(t *testing.T) {
	action := []map[string]interface{}{
		{
			"type":  "iequal",
			"value": "example.com",
			"point": map[string]interface{}{"header": "HOST"},
		},
	}

	details := schemaActionToDetails(action)
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	d := details[0]
	if d.Type != "iequal" {
		t.Errorf("expected type=iequal, got %s", d.Type)
	}
	if d.Value != "example.com" {
		t.Errorf("expected value=example.com, got %v", d.Value)
	}
	point := d.Point
	if len(point) != 2 || point[0] != "header" || point[1] != "HOST" {
		t.Errorf("expected point=[header, HOST], got %v", point)
	}
}

func TestSchemaActionToDetails_PathCondition(t *testing.T) {
	action := []map[string]interface{}{
		{
			"type":  "equal",
			"value": "api",
			"point": map[string]interface{}{"path": "0"},
		},
	}

	details := schemaActionToDetails(action)
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	d := details[0]
	point := d.Point
	if len(point) != 2 || point[0] != "path" {
		t.Errorf("expected point=[path, 0.0], got %v", point)
	}
	// path index is converted to float64.
	if idx, ok := point[1].(float64); !ok || idx != 0 {
		t.Errorf("expected path index 0 as float64, got %v (%T)", point[1], point[1])
	}
	if d.Value != "api" {
		t.Errorf("expected value=api, got %v", d.Value)
	}
}

func TestSchemaActionToDetails_InstanceCondition(t *testing.T) {
	action := []map[string]interface{}{
		{
			"type":  "equal",
			"value": "",
			"point": map[string]interface{}{"instance": "42"},
		},
	}

	details := schemaActionToDetails(action)
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	d := details[0]
	if len(d.Point) != 1 || d.Point[0] != "instance" {
		t.Errorf("expected point=[instance], got %v", d.Point)
	}
	if d.Value != "42" {
		t.Errorf("expected value=42, got %v", d.Value)
	}
}

func TestSchemaActionToDetails_ActionNameCondition(t *testing.T) {
	action := []map[string]interface{}{
		{
			"type":  "equal",
			"value": "",
			"point": map[string]interface{}{"action_name": "users"},
		},
	}

	details := schemaActionToDetails(action)
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	d := details[0]
	if len(d.Point) != 1 || d.Point[0] != "action_name" {
		t.Errorf("expected point=[action_name], got %v", d.Point)
	}
	if d.Value != "users" {
		t.Errorf("expected value=users, got %v", d.Value)
	}
}

func TestSchemaActionToDetails_AbsentCondition(t *testing.T) {
	action := []map[string]interface{}{
		{
			"type":  "absent",
			"value": "",
			"point": map[string]interface{}{"path": "0"},
		},
	}

	details := schemaActionToDetails(action)
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	d := details[0]
	if d.Type != "absent" {
		t.Errorf("expected type=absent, got %s", d.Type)
	}
	if d.Value != nil {
		t.Errorf("expected nil value for absent condition, got %v", d.Value)
	}
}

func TestBuildRulesFromHits_BasicGrouping(t *testing.T) {
	hits := []*wallarm.Hit{
		{
			ID:     []string{"a", "1"},
			Type:   "sqli",
			Stamps: []int{100, 200},
			Point:  []interface{}{"get", "q"},
			Domain: "example.com",
			Path:   "/api",
			PoolID: 1,
		},
		{
			ID:     []string{"a", "2"},
			Type:   "xss",
			Stamps: []int{300},
			Point:  []interface{}{"get", "q"},
			Domain: "example.com",
			Path:   "/api",
			PoolID: 1,
		},
	}

	actionDetails := []wallarm.ActionDetails{
		{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
	}

	groups, schemaActions := groupHitsForRules(hits, actionDetails, defaultAllowedAttackTypes)

	// Should produce 1 point group with 3 stamps and 2 attack types.
	if len(groups) != 1 {
		t.Fatalf("expected 1 point group, got %d", len(groups))
	}
	for _, g := range groups {
		if len(g.Stamps) != 3 {
			t.Errorf("expected 3 stamps, got %d", len(g.Stamps))
		}
		if len(g.AttackTypes) != 2 {
			t.Errorf("expected 2 attack types, got %d", len(g.AttackTypes))
		}
	}
	if len(schemaActions) != 1 {
		t.Errorf("expected 1 schema action, got %d", len(schemaActions))
	}
}

func TestGroupHitsForRules_EmptyHits(t *testing.T) {
	groups, _ := groupHitsForRules(nil, nil, defaultAllowedAttackTypes)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for nil hits, got %d", len(groups))
	}

	groups, _ = groupHitsForRules([]*wallarm.Hit{}, []wallarm.ActionDetails{}, defaultAllowedAttackTypes)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for empty hits, got %d", len(groups))
	}
}

func TestAccDataSourceHits(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceHitsConfig("nonexistent-request-id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceHitsExists("data.wallarm_hits.test"),
					resource.TestCheckResourceAttr("data.wallarm_hits.test", "hits.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceHitsAttackMode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceHitsAttackModeConfig("nonexistent-request-id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceHitsExists("data.wallarm_hits.test_attack"),
					resource.TestCheckResourceAttr("data.wallarm_hits.test_attack", "hits.#", "0"),
					resource.TestCheckResourceAttr("data.wallarm_hits.test_attack", "mode", "attack"),
				),
			},
		},
	})
}

func testAccCheckDataSourceHitsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		return nil
	}
}

func testAccDataSourceHitsConfig(requestID string) string {
	return fmt.Sprintf(`data "wallarm_hits" "test" {
  request_id = "%s"
}`, requestID)
}

func testAccDataSourceHitsAttackModeConfig(requestID string) string {
	return fmt.Sprintf(`data "wallarm_hits" "test_attack" {
  request_id = "%s"
  mode       = "attack"
}`, requestID)
}
