package wallarm

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"
)

type ruleNotFoundError struct {
	clientID int
	ruleID   int
}

func (e *ruleNotFoundError) Error() string {
	return fmt.Sprintf("rule %d for client %d not found", e.ruleID, e.clientID)
}

// retrieveClientID extracts client_id from a resource or falls back to the provider default.
func retrieveClientID(d *schema.ResourceData, m interface{}) (int, error) {
	meta := m.(*ProviderMeta)
	return meta.RetrieveClientID(d)
}

// apiClient extracts the wallarm.API client from the provider meta.
func apiClient(m interface{}) wallarm.API {
	return m.(*ProviderMeta).Client
}
