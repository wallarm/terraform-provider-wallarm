package wallarm

import (
	"log"

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
