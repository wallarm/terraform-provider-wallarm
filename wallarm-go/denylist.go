package wallarm

type Denylist interface {
	DenylistRead(clientID int) ([]IPRule, error)
	DenylistCreate(clientID int, params IPRuleCreationParams) error
	DenylistDelete(clientID int, ids []int) error
}

// DenylistRead requests the current denylist for the future purposes.
// It is going to respond with the list of IP addresses.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) DenylistRead(clientID int) ([]IPRule, error) {
	return api.IPListRead(DenylistType, clientID)
}

// DenylistCreate creates a denylist in the Wallarm Cloud.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) DenylistCreate(clientID int, params IPRuleCreationParams) error {
	params.List = DenylistType
	return api.IPListCreate(clientID, params)
}

// DenylistDelete deletes a denylist for the client.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) DenylistDelete(clientID int, ids []int) error {
	return api.IPListDelete(DenylistType, clientID, ids)
}
