package wallarm

import (
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/wallarm/wallarm-go"

	"github.com/pkg/errors"
)

const (
	apiURL = "https://api.wallarm.com"
)

var (
	// ClientID of the Provider User communicating with
	// the Wallarm Cloud
	ClientID int

	// APIURL is the base URL of the Wallarm API (set during provider configuration).
	APIURL string

	// APIHeaders contains the authentication headers for direct HTTP calls.
	APIHeaders http.Header

	// HTTPClient is the HTTP client used for direct API calls.
	HTTPClient *http.Client
)

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
