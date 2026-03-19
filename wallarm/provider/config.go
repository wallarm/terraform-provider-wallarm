package wallarm

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"

	"github.com/pkg/errors"
)

// ProviderMeta holds the configured API client and provider-level settings.
// It is returned from ProviderConfigure and passed as meta to all CRUD functions.
type ProviderMeta struct {
	Client                  wallarm.API
	DefaultClientID         int
	RequireExplicitClientID bool
	IPListCache             *IPListCache
}

// RetrieveClientID returns the client_id from the resource if set,
// otherwise falls back to the provider's default. Returns an error
// if require_explicit_client_id is enabled and no resource-level client_id is set.
func (pm *ProviderMeta) RetrieveClientID(d *schema.ResourceData) (int, error) {
	if v, ok := d.GetOk("client_id"); ok {
		return v.(int), nil
	}
	if pm.RequireExplicitClientID {
		return 0, fmt.Errorf(
			"client_id is required on this resource because " +
				"require_explicit_client_id is enabled on the provider")
	}
	return pm.DefaultClientID, nil
}

// Config specifies client related parameters used within calls.
type Config struct {
	Options []wallarm.Option
}

// Client returns a new client to access the Wallarm Cloud.
func (c *Config) Client() (wallarm.API, error) {
	client, err := wallarm.New(c.Options...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a new Wallarm client")
	}

	log.Printf("[INFO] Wallarm Client configured")
	return client, nil
}

func GetValueWithTypeCastingOrOverridedDefault[T any](d *schema.ResourceData, name string, overridedDefaultValue T) T {
	resourceValue := d.Get(name)
	if resourceValue == nil {
		return overridedDefaultValue
	}
	v, ok := resourceValue.(T)
	if !ok {
		return overridedDefaultValue
	}
	return v
}
