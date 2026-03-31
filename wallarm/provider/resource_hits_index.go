package wallarm

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmHitsIndex() *schema.Resource {
	return &schema.Resource{
		Description: "Persistent index of request_ids whose hits have been fetched. " +
			"Tracks which request_ids are cached. Use cached_request_ids to gate " +
			"data.wallarm_hits — only fetch for IDs not in the index.",

		CreateContext: resourceHitsIndexCreate,
		ReadContext:   resourceHitsIndexRead,
		UpdateContext: resourceHitsIndexUpdate,
		DeleteContext: resourceHitsIndexDelete,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"request_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of request_ids to track.",
			},

			// Computed — the persistent index
			"cached_request_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Comma-separated request IDs currently in the index.",
			},
		},
	}
}

// ─── CRUD ───────────────────────────────────────────────────────────────────

func resourceHitsIndexCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("hits_index_%d", clientID))
	if err := d.Set("client_id", clientID); err != nil {
		return diag.FromErr(err)
	}

	return syncCachedRequestIDs(d)
}

func resourceHitsIndexRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Sync cached_request_ids on every Read to ensure state is correct.
	return syncCachedRequestIDs(d)
}

func resourceHitsIndexUpdate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return syncCachedRequestIDs(d)
}

func resourceHitsIndexDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

// ─── Core logic ─────────────────────────────────────────────────────────────

// syncCachedRequestIDs sets cached_request_ids to match request_ids from config.
// On Create: all request_ids become cached (they'll be fetched by data.wallarm_hits in the same apply).
// On Update: new request_ids are added, removed ones are dropped.
func syncCachedRequestIDs(d *schema.ResourceData) diag.Diagnostics {
	requestIDsSet := d.Get("request_ids").(*schema.Set)
	cachedIDs := expandInterfaceToStringList(requestIDsSet.List())
	sort.Strings(cachedIDs)

	joined := strings.Join(cachedIDs, ",")
	log.Printf("[INFO] wallarm_hits_index: syncing cached_request_ids = %q (%d entries)", joined, len(cachedIDs))
	if err := d.Set("cached_request_ids", joined); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
