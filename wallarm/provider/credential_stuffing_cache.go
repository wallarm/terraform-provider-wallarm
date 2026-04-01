package wallarm

import (
	"log"
	"sync"
	"time"

	wallarm "github.com/wallarm/wallarm-go"
)

// CredentialStuffingCache caches the full list of credential stuffing configs.
// The API returns all configs in a single call — no pagination needed.
// First Read fetches from API, subsequent Reads serve from cache.
type CredentialStuffingCache struct {
	mu       sync.Mutex
	configs  map[int]*wallarm.ActionBody // keyed by rule ID
	loaded   bool
	clientID int
}

// NewCredentialStuffingCache creates an empty cache.
func NewCredentialStuffingCache() *CredentialStuffingCache {
	return &CredentialStuffingCache{
		configs: make(map[int]*wallarm.ActionBody),
	}
}

// GetOrFetch returns the config matching ruleID, fetching from the API on first call.
func (c *CredentialStuffingCache) GetOrFetch(client wallarm.API, clientID, ruleID int) (*wallarm.ActionBody, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded && c.clientID == clientID {
		if config, ok := c.configs[ruleID]; ok {
			log.Printf("[DEBUG] CredentialStuffingCache: cache hit for rule_id=%d", ruleID)
			return config, nil
		}
		log.Printf("[DEBUG] CredentialStuffingCache: rule_id=%d not in cache (%d entries)", ruleID, len(c.configs))
		return nil, &ruleNotFoundError{clientID: clientID, ruleID: ruleID}
	}

	// Fetch all configs from API.
	start := time.Now()
	configs, err := client.CredentialStuffingConfigsRead(clientID)
	if err != nil {
		return nil, err
	}

	c.configs = make(map[int]*wallarm.ActionBody, len(configs))
	for i := range configs {
		c.configs[configs[i].ID] = &configs[i]
	}
	c.loaded = true
	c.clientID = clientID

	log.Printf("[INFO] CredentialStuffingCache: loaded %d configs for client_id=%d in %s",
		len(configs), clientID, time.Since(start).Round(time.Millisecond))

	if config, ok := c.configs[ruleID]; ok {
		return config, nil
	}
	return nil, &ruleNotFoundError{clientID: clientID, ruleID: ruleID}
}

// LoadAll returns all configs, fetching from the API only if not already cached.
// Used by data_source_rules.
func (c *CredentialStuffingCache) LoadAll(client wallarm.API, clientID int) ([]wallarm.ActionBody, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded && c.clientID == clientID {
		configs := make([]wallarm.ActionBody, 0, len(c.configs))
		for _, config := range c.configs {
			configs = append(configs, *config)
		}
		log.Printf("[DEBUG] CredentialStuffingCache: LoadAll cache hit (%d configs)", len(configs))
		return configs, nil
	}

	start := time.Now()
	configs, err := client.CredentialStuffingConfigsRead(clientID)
	if err != nil {
		return nil, err
	}

	c.configs = make(map[int]*wallarm.ActionBody, len(configs))
	for i := range configs {
		c.configs[configs[i].ID] = &configs[i]
	}
	c.loaded = true
	c.clientID = clientID

	log.Printf("[INFO] CredentialStuffingCache: loaded %d configs for client_id=%d in %s",
		len(configs), clientID, time.Since(start).Round(time.Millisecond))

	return configs, nil
}

// Invalidate clears the cache so the next access triggers a fresh fetch.
func (c *CredentialStuffingCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.configs = make(map[int]*wallarm.ActionBody)
	c.loaded = false
	log.Printf("[DEBUG] CredentialStuffingCache: invalidated")
}
