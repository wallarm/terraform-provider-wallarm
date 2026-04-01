package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ─── Unit tests ─────────────────────────────────────────────────────────────

func TestSyncCachedRequestIDs_Basic(t *testing.T) {
	r := resourceWallarmHitsIndex()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"request_ids": []interface{}{"abc123", "def456", "ghi789"},
	})

	diags := syncCachedRequestIDs(d)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	got := d.Get("cached_request_ids").(string)
	want := "abc123,def456,ghi789"
	if got != want {
		t.Errorf("cached_request_ids = %q, want %q", got, want)
	}
}

func TestSyncCachedRequestIDs_Sorted(t *testing.T) {
	r := resourceWallarmHitsIndex()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"request_ids": []interface{}{"zzz", "aaa", "mmm"},
	})

	diags := syncCachedRequestIDs(d)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	got := d.Get("cached_request_ids").(string)
	want := "aaa,mmm,zzz"
	if got != want {
		t.Errorf("cached_request_ids = %q, want %q (should be sorted)", got, want)
	}
}

func TestSyncCachedRequestIDs_Empty(t *testing.T) {
	r := resourceWallarmHitsIndex()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"request_ids": []interface{}{},
	})

	diags := syncCachedRequestIDs(d)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	got := d.Get("cached_request_ids").(string)
	if got != "" {
		t.Errorf("cached_request_ids = %q, want empty string", got)
	}
}

func TestSyncCachedRequestIDs_Single(t *testing.T) {
	r := resourceWallarmHitsIndex()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"request_ids": []interface{}{"only_one"},
	})

	diags := syncCachedRequestIDs(d)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	got := d.Get("cached_request_ids").(string)
	want := "only_one"
	if got != want {
		t.Errorf("cached_request_ids = %q, want %q", got, want)
	}
}

// ─── Acceptance tests ───────────────────────────────────────────────────────

func TestAccWallarmHitsIndex_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_hits_index." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmHitsIndexConfig(rnd, []string{"abc123", "def456"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "request_ids.#", "2"),
					resource.TestCheckResourceAttr(name, "cached_request_ids", "abc123,def456"),
				),
			},
		},
	})
}

func TestAccWallarmHitsIndex_Update(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_hits_index." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmHitsIndexConfig(rnd, []string{"abc123"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "request_ids.#", "1"),
					resource.TestCheckResourceAttr(name, "cached_request_ids", "abc123"),
				),
			},
			{
				Config: testWallarmHitsIndexConfig(rnd, []string{"abc123", "def456", "ghi789"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "request_ids.#", "3"),
					resource.TestCheckResourceAttr(name, "cached_request_ids", "abc123,def456,ghi789"),
				),
			},
			{
				Config: testWallarmHitsIndexConfig(rnd, []string{"def456"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "request_ids.#", "1"),
					resource.TestCheckResourceAttr(name, "cached_request_ids", "def456"),
				),
			},
		},
	})
}

func testWallarmHitsIndexConfig(resourceID string, requestIDs []string) string {
	ids := ""
	for _, id := range requestIDs {
		ids += fmt.Sprintf("    %q,\n", id)
	}
	return fmt.Sprintf(`
resource "wallarm_hits_index" "%s" {
  request_ids = [
%s  ]
}`, resourceID, ids)
}
