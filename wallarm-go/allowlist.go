package wallarm

type Allowlist interface {
	AllowlistRead(clientID int) ([]IPRule, error)
	AllowlistCreate(clientID int, params IPRuleCreationParams) error
	AllowlistDelete(clientID int, ids []int) error
}

// AllowlistRead requests the current allowlist for the future purposes.
// It is going to respond with the list of IP addresses.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) AllowlistRead(clientID int) ([]IPRule, error) {
	return api.IPListRead(AllowlistType, clientID)
}

// AllowlistCreate creates a allowlist in the Wallarm Cloud.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) AllowlistCreate(clientID int, params IPRuleCreationParams) error {
	params.List = AllowlistType
	return api.IPListCreate(clientID, params)
}

// AllowlistDelete deletes a allowlist for the client.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) AllowlistDelete(clientID int, ids []int) error {
	return api.IPListDelete(AllowlistType, clientID, ids)
}
