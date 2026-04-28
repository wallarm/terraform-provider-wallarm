package resourcerule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// CounterDelete must return without dereferencing the meta argument.
// Passing nil meta proves the contract: any attempt to reach apiClient(m)
// or HintDelete would panic.
func TestCounterDelete_NoAPICallNoPanic(t *testing.T) {
	del := CounterDelete("wallarm_rule_test_counter")

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, map[string]interface{}{})
	d.SetId("1/2/3")

	diags := del(context.Background(), d, nil)
	if diags.HasError() {
		t.Fatalf("Delete returned diagnostics: %v", diags)
	}
}
