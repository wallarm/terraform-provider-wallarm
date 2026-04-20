package resourcerule

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Import is used by rule resources whose import is purely a
// string parse with no API lookup.
func Import(ruleType string) schema.StateContextFunc {
	return func(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
		parts := strings.SplitN(d.Id(), "/", 4)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{clientID}/{actionID}/{ruleID}\"", d.Id())
		}
		clientID, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid client_id: %w", err)
		}
		actionID, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid action_id: %w", err)
		}
		ruleID, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid rule_id: %w", err)
		}
		d.Set("action_id", actionID)
		d.Set("rule_id", ruleID)
		d.Set("rule_type", ruleType)
		d.SetId(fmt.Sprintf("%d/%d/%d", clientID, actionID, ruleID))
		return []*schema.ResourceData{d}, nil
	}
}
