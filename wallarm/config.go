package wallarm

import (
	"log"

	wallarm "github.com/416e64726579/wallarm-go"
	"github.com/pkg/errors"
)

const (
	apiURL = "https://api.wallarm.com"
)

var (
	// ClientID of the Provider User communicating with
	// the Wallarm Cloud
	ClientID int
)

// Config specifies client related parameters used within calls.
type Config struct {
	apiURL    string
	apiUUID   string
	apiSecret string
	Options   []wallarm.Option
}

// Client returns a new client to access the Wallarm Cloud.
func (c *Config) Client() (wallarm.API, error) {
	client, err := wallarm.New(c.Options...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a new Wallarm client")
	}

	log.Printf("Wallarm Client configured for the user with UUID: %s", c.apiUUID)
	return client, nil
}
