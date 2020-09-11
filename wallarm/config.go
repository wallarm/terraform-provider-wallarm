package wallarm

import (
	"log"

	wallarm "github.com/416e64726579/wallarm-go"
	"github.com/pkg/errors"
)

// Config specifies client related parameters used within calls.
type Config struct {
	apiURL    string
	apiUUID   string
	apiSecret string
	ClientID  []int
	baseURL   string
	Options   []wallarm.Option
}

// Client returns a new client to access Wallarm Cloud.
func (c *Config) Client() (*wallarm.API, error) {
	var err error
	var client *wallarm.API

	client, err = wallarm.New(c.apiURL, c.Options...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a new Wallarm client")
	}
	log.Printf("Wallarm Client configured for the user with UUID: %s", c.apiUUID)
	return client, nil
}
