package wallarm

import (
	"encoding/json"
	"fmt"
)

type (
	WallarmMode interface {
		WallarmModeUpdate(wallarmModeBody *WallarmModeParams, clientID int) (*WallarmModeResponse, error)
		WallarmModeRead(clientID int) (*WallarmModeResponse, error)
	}

	WallarmModeParams struct {
		Mode string `json:"mode"`
	}

	WallarmModeResponse struct {
		Status int               `json:"status"`
		Body   WallarmModeParams `json:"body"`
	}
)

// WallarmModeUpdate changes wallarm_mode state.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) WallarmModeUpdate(wallarmModeBody *WallarmModeParams, clientID int) (*WallarmModeResponse, error) {
	url := fmt.Sprintf("/v2/client/%d/rules/wallarm_mode", clientID)
	rawResp, err := api.makeRequest("PUT", url, "wallarm_mode", wallarmModeBody)
	if err != nil {
		return nil, err
	}

	var resp WallarmModeResponse
	if err = json.Unmarshal(rawResp, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// WallarmModeRead requests wallarm_mode info.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) WallarmModeRead(clientID int) (*WallarmModeResponse, error) {
	url := fmt.Sprintf("/v2/client/%d/rules/wallarm_mode", clientID)
	rawResp, err := api.makeRequest("GET", url, "wallarm_mode", nil)
	if err != nil {
		return nil, err
	}

	var resp WallarmModeResponse
	if err = json.Unmarshal(rawResp, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
