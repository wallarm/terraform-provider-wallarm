package wallarm

import (
	"fmt"
	"log"

	wallarm "github.com/416e64726579/wallarm-go"
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
		return nil, fmt.Errorf("Error creating a new Wallarm client: %s", err)
	}
	log.Printf("Wallarm Client configured for the user with UUID: %s", c.apiUUID)
	return client, nil
}
