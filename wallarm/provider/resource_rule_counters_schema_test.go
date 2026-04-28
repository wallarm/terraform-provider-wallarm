package wallarm

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Counter resources have no UpdateContext (state-only Delete, no Update).
// Every common field made mutable in v2.3.7 (title, active, set, comment,
// variativity_disabled) must be Computed-only on counters — otherwise the
// SDK plans an update-in-place and panics calling a nil UpdateContext.
func TestCounter_ReadOnlyMutableCommonFields(t *testing.T) {
	resources := []struct {
		name string
		res  *schema.Resource
	}{
		{"bruteforce_counter", resourceWallarmBruteForceCounter()},
		{"dirbust_counter", resourceWallarmDirbustCounter()},
		{"bola_counter", resourceWallarmBolaCounter()},
	}
	readOnlyFields := []string{"title", "active", "set", "comment", "variativity_disabled"}

	for _, tc := range resources {
		t.Run(tc.name, func(t *testing.T) {
			if tc.res.UpdateContext != nil {
				t.Errorf("UpdateContext must be nil — counters are immutable")
			}
			for _, f := range readOnlyFields {
				s, ok := tc.res.Schema[f]
				if !ok {
					t.Errorf("missing field %q", f)
					continue
				}
				if !s.Computed {
					t.Errorf("field %q must be Computed (counter has no UpdateContext)", f)
				}
				if s.Optional {
					t.Errorf("field %q must NOT be Optional on counter (would let HCL write a plan that hits nil UpdateContext)", f)
				}
			}
		})
	}
}
