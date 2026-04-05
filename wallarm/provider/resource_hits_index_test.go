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

	cached := d.Get("cached_request_ids").(*schema.Set)
	if cached.Len() != 3 {
		t.Errorf("expected 3 cached IDs, got %d", cached.Len())
	}
	for _, id := range []string{"abc123", "def456", "ghi789"} {
		if !cached.Contains(id) {
			t.Errorf("expected cached_request_ids to contain %q", id)
		}
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

	cached := d.Get("cached_request_ids").(*schema.Set)
	if cached.Len() != 0 {
		t.Errorf("expected 0 cached IDs, got %d", cached.Len())
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
					resource.TestCheckResourceAttr(name, "ready", "true"),
					resource.TestCheckResourceAttr(name, "cached_request_ids.#", "2"),
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
					// After Create+Read, cached_request_ids syncs to match.
					resource.TestCheckResourceAttr(name, "cached_request_ids.#", "1"),
				),
			},
			{
				Config: testWallarmHitsIndexConfig(rnd, []string{"abc123", "def456", "ghi789"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "request_ids.#", "3"),
					// cached_request_ids lags by design: CustomizeDiff preserves
					// old value so new IDs appear as "uncached" for gating.
					// After this apply+refresh cycle, it catches up.
				),
			},
			{
				// Third step: by now Read has synced cached_request_ids.
				// Shrink to 1 ID to test removal.
				Config: testWallarmHitsIndexConfig(rnd, []string{"def456"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "request_ids.#", "1"),
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
