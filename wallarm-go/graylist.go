package wallarm

type Graylist interface {
	GraylistRead(clientID int) ([]IPRule, error)
	GraylistCreate(clientID int, params IPRuleCreationParams) error
	GraylistDelete(clientID int, ids []int) error
}

// GraylistRead requests the current graylist for the future purposes.
// It is going to respond with the list of IP addresses.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) GraylistRead(clientID int) ([]IPRule, error) {
	return api.IPListRead(GraylistType, clientID)
}

// GraylistCreate creates a graylist in the Wallarm Cloud.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) GraylistCreate(clientID int, params IPRuleCreationParams) error {
	params.List = GraylistType
	return api.IPListCreate(clientID, params)
}

// GraylistDelete deletes a graylist for the client.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) GraylistDelete(clientID int, ids []int) error {
	return api.IPListDelete(GraylistType, clientID, ids)
}
