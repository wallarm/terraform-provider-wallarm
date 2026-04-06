package wallarm

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmHitsIndex() *schema.Resource {
	return &schema.Resource{
		Description: "Persistent index of request_ids whose hits have been fetched. " +
			"Tracks which request_ids are cached. Use ready and cached_request_ids " +
			"to gate data.wallarm_hits.",

		CreateContext: resourceHitsIndexCreate,
		ReadContext:   resourceHitsIndexRead,
		UpdateContext: resourceHitsIndexUpdate,
		DeleteContext: resourceHitsIndexDelete,

		CustomizeDiff: customdiff.All(hitsIndexCustomizeDiff),

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"request_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of request_ids to track.",
			},

			"ready": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "False on first create, true after. Use to gate data source: ready ? _new_ids : all_ids.",
			},

			"cached_request_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of request IDs currently in the index.",
			},
		},
	}
}

// hitsIndexCustomizeDiff makes ready and cached_request_ids known at plan time.
// On Create: ready=false, cached_request_ids=empty.
// On Update: ready=true (from state), cached_request_ids=request_ids (from state).
func hitsIndexCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if d.Id() == "" {
		// Create: not ready yet, no cached IDs.
		if err := d.SetNew("ready", false); err != nil {
			return err
		}
		return d.SetNew("cached_request_ids", schema.NewSet(schema.HashString, []interface{}{}))
	}
	// Update: preserve old cached_request_ids from state so new IDs appear as
	// "uncached" during plan, enabling the data source gating to fetch them.
	// Update+Read will sync to match request_ids — the SDK allows Computed
	// fields to differ from the planned value during apply.
	return nil
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

	// Set ready=true so next plan sees it as ready.
	if err := d.Set("ready", true); err != nil {
		return diag.FromErr(err)
	}

	// Sync cached_request_ids = request_ids.
	return syncCachedRequestIDs(d)
}

func resourceHitsIndexRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Ensure ready is true (resource exists).
	if err := d.Set("ready", true); err != nil {
		return diag.FromErr(err)
	}
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
func syncCachedRequestIDs(d *schema.ResourceData) diag.Diagnostics {
	requestIDsSet := d.Get("request_ids").(*schema.Set)
	log.Printf("[INFO] wallarm_hits_index: syncing cached_request_ids (%d entries)", requestIDsSet.Len())
	if err := d.Set("cached_request_ids", requestIDsSet); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
