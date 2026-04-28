package resourcerule

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"
)

// Delete returns a DeleteContextFunc that issues HintDelete via the
// provider's wallarm.API client. It drops the resource from state on success
// and on the no-op path (HTTP 200 with empty body — rule already absent or
// blocked server-side); the no-op case is logged at WARN to surface drift
// that would otherwise go unnoticed.
func Delete(cp func(m interface{}) wallarm.API) schema.DeleteContextFunc {
	return func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		clientID := d.Get("client_id").(int)
		ruleID := d.Get("rule_id").(int)

		resp, err := cp(m).HintDelete(&wallarm.HintDelete{
			Filter: &wallarm.HintDeleteFilter{
				Clientid: []int{clientID},
				ID:       []int{ruleID},
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
		LogIfHintDeleteNoOp(resp, ruleID)
		return nil
	}
}

// LogIfHintDeleteNoOp emits a WARN log when HintDelete returned an empty body,
// indicating the API rejected the delete (counter hint, already deleted, or
// out-of-band removal). Shared by Delete factory and per-resource Delete
// functions that need additional cleanup (e.g. credential_stuffing rules
// invalidating CredentialStuffingCache).
func LogIfHintDeleteNoOp(resp *wallarm.HintDeleteResp, ruleID int) {
	if resp == nil || len(resp.Body) == 0 {
		log.Printf("[WARN] HintDelete on rule %d returned empty body — "+
			"rule may have been deleted out-of-band or blocked server-side; "+
			"removing from state", ruleID)
	}
}
