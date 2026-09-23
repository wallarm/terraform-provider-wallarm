package wallarm

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

type legacyHitReadAPI struct {
	wallarm.API
	pages [][]*wallarm.Hit
	calls int
}

func (m *legacyHitReadAPI) HitRead(*wallarm.HitReadRequest) ([]*wallarm.Hit, error) {
	m.calls++
	if m.calls > len(m.pages) {
		return nil, nil
	}
	return m.pages[m.calls-1], nil
}

func TestLegacyFetchDirectHits(t *testing.T) {
	api := &legacyHitReadAPI{pages: [][]*wallarm.Hit{{{ID: []string{"a", "1"}}}}}
	got, err := fetchDirectHits(api, 1, "r", nil)
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestLegacyFetchRelatedHitsByAttackIDs(t *testing.T) {
	direct := []*wallarm.Hit{{ID: []string{"a", "1"}, AttackID: []string{"idx", "att"}, Domain: "d", Path: "/p", PoolID: 1}}
	api := &legacyHitReadAPI{pages: [][]*wallarm.Hit{{
		{ID: []string{"a", "2"}, Domain: "d", Path: "/p", PoolID: 1},
		{ID: []string{"a", "3"}, Domain: "other", Path: "/p", PoolID: 1},
	}}}
	got, err := fetchRelatedHitsByAttackIDs(api, 1, direct, []string{"sqli"}, nil, "d", "/p", 1)
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestLegacyBuildTimeRange(t *testing.T) {
	s := map[string]*schema.Schema{"time": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeInt}}}
	d := schema.TestResourceDataRaw(t, s, map[string]any{"time": []any{1, 2}})
	got := buildTimeRange(d)
	if len(got) != 1 || got[0][0] != 1 || got[0][1] != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestLegacyHitsToSchemaList(t *testing.T) {
	got := hitsToSchemaList([]*wallarm.Hit{{ID: []string{"a", "1"}, Point: []any{"get", "q"}}})
	if len(got) != 1 {
		t.Fatalf("got %v", got)
	}
}
