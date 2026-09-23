package wallarm

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

type vectorsAPI struct {
	wallarm.API
	resp   *wallarm.AttackVectorsByRequestResp
	params *wallarm.AttackVectorsByRequestParams
	client int
}

func (m *vectorsAPI) AttackVectorsByRequest(clientID int, p *wallarm.AttackVectorsByRequestParams) (*wallarm.AttackVectorsByRequestResp, error) {
	m.client, m.params = clientID, p
	return m.resp, nil
}

var misleadingConditions = []wallarm.AttackVectorCondition{
	{Type: "equal", Point: []any{"scheme"}, Value: "http"},
	{Type: "absent", Point: []any{"action_name"}},
}

func readVectors(t *testing.T, api *vectorsAPI, raw map[string]any) *schema.ResourceData {
	t.Helper()
	d := schema.TestResourceDataRaw(t, dataSourceWallarmHits().Schema, raw)
	diags := dataSourceWallarmHitsRead(context.Background(), d, &ProviderMeta{Client: api, DefaultClientID: 8649})
	if diags.HasError() {
		t.Fatal(diags)
	}
	return d
}

func TestDataSourceHitsRead_ActionHashMatchesBuilder(t *testing.T) {
	cases := []struct {
		name, host, path string
		app              int
		inc              bool
		want             string
	}{
		{"abc", "127.0.0.1:8080", "/a/b/c", 13, true, "379f1a537cd4392116b0c492842daf3b73afb6e5ed146db5ed2b02cd4a5ab08d"},
		{"abc no instance", "127.0.0.1:8080", "/a/b/c", 13, false, "f780176a5e29f5f62dcaefd84377d03855f54ace12bcbbcd89de83fd502d984f"},
		{"php", "34.66.161.204:80", "/app/vendor/phpunit/phpunit/src/Util/PHP/eval-stdin.php", -1, true, "2c2c1c3c351035477a5a8bd894bb20fb36a96253eb9fc10a7eb64bcac4b41fa5"},
		{"root", "h:80", "/", -1, true, ""},
		{"dotfile", "h:80", "/.env", -1, true, ""},
		{"multi-dot dotfile", "h:80", "/.env.dev", -1, false, ""},
		{"trailing slash", "h:80", "/tmp/", -1, true, ""},
		{"trailing slash dotted", "h:80", "/etc/apt/sources.list.d/", 13, true, ""},
		{"empty host", "", "/setup.cgi", -1, true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api := &vectorsAPI{resp: &wallarm.AttackVectorsByRequestResp{Data: []wallarm.AttackVector{
				{VectorID: "v1", RequestID: "r1", ApplicationID: tc.app, Type: "xss", Point: `["get","q"]`, Stamps: []int{7}, Host: tc.host, Path: tc.path, ActionConditions: misleadingConditions},
			}}}
			d := readVectors(t, api, map[string]any{"request_id": "r1", "include_instance": tc.inc})
			action := buildActionFromHit(tc.host, tc.path, tc.app, tc.inc)
			builder := resourcerule.ConditionsHash(schemaActionToDetails(action))
			if got := d.Get("action_hash"); got != builder {
				t.Fatalf("action_hash %v, builder %v", got, builder)
			}
			if tc.want != "" && builder != tc.want {
				t.Fatalf("builder hash %v, old flow %v", builder, tc.want)
			}
			if !d.Get("action").(*schema.Set).Equal(actionToSchemaSet(action)) {
				t.Fatalf("action set differs from builder: %v", d.Get("action"))
			}
			if n := len(d.Get("action_conditions").([]any)); n != len(action) {
				t.Fatalf("action_conditions %d, builder %d", n, len(action))
			}
		})
	}
}

func TestDataSourceHitsRead_Vectors(t *testing.T) {
	api := &vectorsAPI{resp: &wallarm.AttackVectorsByRequestResp{Data: []wallarm.AttackVector{
		{VectorID: "v1", RequestID: "r1", ApplicationID: 13, Type: "xss", Point: `["header","XSS"]`, Stamps: []int{1, 2}, Host: "127.0.0.1:8080", Path: "/a/b/c", RemoteAddr4: "1.2.3.4", RequestTime: 1790142071217, KnownAttack: "a,b"},
		{VectorID: "v2", RequestID: "r1", ApplicationID: 13, Type: "ptrav", Point: `["get","q"]`, Stamps: []int{3}, Host: "127.0.0.1:8080", Path: "/a/b/c"},
		{VectorID: "v3", RequestID: "r1", ApplicationID: 13, Type: "sqli", Point: `["post"]`, Stamps: []int{9}, Host: "127.0.0.1:8080", Path: "/a/b/c"},
	}}}
	d := readVectors(t, api, map[string]any{"request_id": "r1", "attack_types": []any{"xss", "ptrav"}})
	if api.client != 8649 || api.params.RequestID != "r1" || api.params.Limit != 0 {
		t.Fatalf("call: %d %+v", api.client, api.params)
	}
	if d.Id() != "hits_8649_r1" || d.Get("hits_count") != 3 {
		t.Fatalf("id=%v count=%v", d.Id(), d.Get("hits_count"))
	}
	if d.Get("hits.0.time") != 1790142071 || !reflect.DeepEqual(d.Get("hits.0.known_attack"), []any{"a", "b"}) || !reflect.DeepEqual(d.Get("hits.0.id"), []any{"v1"}) || d.Get("hits.0.poolid") != 13 || d.Get("hits.0.domain") != "127.0.0.1:8080" {
		t.Fatalf("hit0: %v", d.Get("hits.0"))
	}
	xssPoint := []any{"header", "XSS"}
	xssHash := resourcerule.PointHash(xssPoint)
	if xssHash == "" || d.Get("hits.0.point_hash") != xssHash {
		t.Fatalf("hits.0.point_hash %v, want %v", d.Get("hits.0.point_hash"), xssHash)
	}
	if !reflect.DeepEqual(d.Get("hits.0.stamps"), []any{1, 2}) {
		t.Fatalf("hits.0.stamps %v", d.Get("hits.0.stamps"))
	}

	var agg aggregatedOutput
	if err := json.Unmarshal([]byte(d.Get("aggregated").(string)), &agg); err != nil {
		t.Fatal(err)
	}
	want := map[string]aggregatedGroup{
		"xss":   {Key: xssHash[:16] + "_xss", Point: resourcerule.WrapPointElements(xssPoint), Stamps: []int{1, 2}, AttackType: "xss", DisableAttackType: true},
		"ptrav": {Key: resourcerule.PointHash([]any{"get", "q"})[:16] + "_ptrav", Point: resourcerule.WrapPointElements([]any{"get", "q"}), Stamps: []int{3}, AttackType: "ptrav", DisableAttackType: true},
	}
	got := make(map[string]aggregatedGroup, len(agg.Groups))
	for _, g := range agg.Groups {
		got[g.AttackType] = g
	}
	if len(agg.Groups) != len(want) || !reflect.DeepEqual(got, want) {
		t.Fatalf("aggregated groups %+v, want %+v", agg.Groups, want)
	}
	if agg.ActionHash != d.Get("action_hash").(string)[:16] {
		t.Fatalf("aggregated action_hash %v, action_hash %v", agg.ActionHash, d.Get("action_hash"))
	}
}

func TestDataSourceHitsRead_InconsistentVectors(t *testing.T) {
	api := &vectorsAPI{resp: &wallarm.AttackVectorsByRequestResp{Data: []wallarm.AttackVector{
		{VectorID: "v1", Point: `["get","q"]`, Host: "h", Path: "/a", ApplicationID: 1},
		{VectorID: "v2", Point: `["get","q"]`, Host: "h", Path: "/b", ApplicationID: 1},
	}}}
	d := schema.TestResourceDataRaw(t, dataSourceWallarmHits().Schema, map[string]any{"request_id": "r1"})
	if diags := dataSourceWallarmHitsRead(context.Background(), d, &ProviderMeta{Client: api, DefaultClientID: 8649}); !diags.HasError() {
		t.Fatal("expected an inconsistency error")
	}
}

func TestDataSourceHitsRead_EmptyAndBadPoint(t *testing.T) {
	api := &vectorsAPI{resp: &wallarm.AttackVectorsByRequestResp{}}
	d := readVectors(t, api, map[string]any{"request_id": "r1"})
	if d.Id() != "hits_8649_r1" || d.Get("hits_count") != 0 || d.Get("action_hash") != "" {
		t.Fatalf("empty: id=%v count=%v", d.Id(), d.Get("hits_count"))
	}
	api.resp = &wallarm.AttackVectorsByRequestResp{Data: []wallarm.AttackVector{{VectorID: "v1", Point: "not json"}}}
	d = schema.TestResourceDataRaw(t, dataSourceWallarmHits().Schema, map[string]any{"request_id": "r1"})
	if diags := dataSourceWallarmHitsRead(context.Background(), d, &ProviderMeta{Client: api, DefaultClientID: 8649}); !diags.HasError() {
		t.Fatal("expected a point decode error")
	}
}

func TestDataSourceHitsSchema_Removals(t *testing.T) {
	s := dataSourceWallarmHits().Schema
	for _, k := range []string{"mode", "time"} {
		if _, ok := s[k]; ok {
			t.Errorf("input %s still in schema", k)
		}
	}
	if _, ok := s["hits"].Elem.(*schema.Resource).Schema["attack_id"]; ok {
		t.Error("hits.attack_id still in schema")
	}
}
