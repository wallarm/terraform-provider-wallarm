package resourcerule

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// CounterDelete returns a state-only DeleteContextFunc for counter rule
// resources (bruteforce_counter, dirbust_counter, bola_counter).
//
// Counter deletes are silently rejected by the Wallarm API (200 + empty body).
// Counters auto-clean ~30s after their last trigger reference is removed.
// Issuing the API call would falsely report destroy success while the counter
// persists server-side; this helper drops the resource from state and emits
// an INFO log instead.
func CounterDelete(resourceName string) schema.DeleteContextFunc {
	return func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
		log.Printf("[INFO] %s %s removed from state; "+
			"the Wallarm API auto-cleans counters after their last trigger reference is gone",
			resourceName, d.Id())
		return nil
	}
}
