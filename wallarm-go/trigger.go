package wallarm

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type (
	// Trigger contains operations available on Triggers resource
	Trigger interface {
		TriggerRead(clientID int) (*TriggerRead, error)
		TriggerCreate(triggerBody *TriggerCreate, clientID int) (*TriggerCreateResp, error)
		TriggerDelete(clientID, triggerID int) error
		TriggerUpdate(triggerBody *TriggerCreate, clientID, triggerID int) (*TriggerResp, error)
	}

	// TriggerFilters is used to specify params for Trigger["Filters"] which is used as a slice
	TriggerFilters struct {
		ID       string        `json:"id"`
		Operator string        `json:"operator"`
		Values   []interface{} `json:"values"`
	}

	// TriggerActions is used to specify params for Trigger["Actions"] which is used as a slice
	TriggerActions struct {
		ID     string `json:"id"`
		Params struct {
			IntegrationIds []int `json:"integration_ids,omitempty"`
			LockTime       int   `json:"lock_time,omitempty"`
		} `json:"params"`
	}

	// TriggerThreshold is used to specify params for Trigger["Threshold"]
	TriggerThreshold struct {
		Period           int      `json:"period"`
		Operator         string   `json:"operator"`
		AllowedOperators []string `json:"allowed_operators"`
		Count            int      `json:"count"`
	}

	// TriggerParam is used to specify params for TriggerCreate
	TriggerParam struct {
		Filters    *[]TriggerFilters `json:"filters"`
		Actions    *[]TriggerActions `json:"actions"`
		TemplateID string            `json:"template_id"`
		Threshold  *TriggerThreshold `json:"threshold"`
		Enabled    bool              `json:"enabled"`
		Name       string            `json:"name,omitempty"`
		Comment    string            `json:"comment,omitempty"`
	}

	// TriggerCreate is used to define JSON body for create action
	TriggerCreate struct {
		Trigger *TriggerParam `json:"trigger"`
	}

	// TriggerResp is returned on successful trigger updating and creating
	TriggerResp struct {
		ID       int           `json:"id"`
		Name     string        `json:"name"`
		Comment  interface{}   `json:"comment"`
		Enabled  bool          `json:"enabled"`
		ClientID int           `json:"client_id"`
		Filters  []interface{} `json:"filters"`
		Actions  []struct {
			ID string `json:"id"`
		} `json:"actions"`
		Thresholds []struct {
			Operator string `json:"operator"`
			Period   int    `json:"period"`
			Count    int    `json:"count"`
		} `json:"thresholds"`
		Template struct {
			ID      string `json:"id"`
			Filters []struct {
				ID               string        `json:"id"`
				Required         bool          `json:"required"`
				Values           []interface{} `json:"values"`
				AllowedOperators []string      `json:"allowed_operators"`
				Operator         string        `json:"operator"`
			} `json:"filters"`
			Threshold struct {
				AllowedOperators []string `json:"allowed_operators"`
				Operator         string   `json:"operator"`
				Period           int      `json:"period"`
				Count            int      `json:"count"`
			} `json:"threshold"`
			Actions []struct {
				ID     string `json:"id"`
				Params struct {
					LockTime int `json:"lock_time"`
				} `json:"params,omitempty"`
			} `json:"actions"`
		} `json:"template"`
		Threshold struct {
			Operator string `json:"operator"`
			Period   int    `json:"period"`
			Count    int    `json:"count"`
		} `json:"threshold"`
	}

	// TriggerCreateResp is the response on the creating a trigger.
	TriggerCreateResp struct {
		*TriggerResp `json:"trigger"`
	}

	// TriggerRead is the response which contains information about all the created Triggers within an account
	TriggerRead struct {
		Triggers []TriggerResp `json:"triggers"`
	}
)

// TriggerRead is used to return Trigger response that is used to distinguish distinct Trigger ID of the trigger.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) TriggerRead(clientID int) (*TriggerRead, error) {

	uri := fmt.Sprintf("/v2/clients/%d/triggers", clientID)
	q := url.Values{}
	q.Add("denormalize", "true")
	query := q.Encode()
	respBody, err := api.makeRequest("GET", uri, "trigger", query)
	if err != nil {
		return nil, err
	}
	var t TriggerRead
	if err = json.Unmarshal(respBody, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// TriggerCreate creates Trigger with the parameters in JSON body.
// For example, define filters and thresholds which trigger actions.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) TriggerCreate(triggerBody *TriggerCreate, clientID int) (*TriggerCreateResp, error) {

	uri := fmt.Sprintf("/v2/clients/%d/triggers", clientID)
	respBody, err := api.makeRequest("POST", uri, "trigger", triggerBody)
	if err != nil {
		return nil, err
	}
	var t TriggerCreateResp
	if err = json.Unmarshal(respBody, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// TriggerDelete deletes Trigger defined by distinct ID.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) TriggerDelete(clientID, triggerID int) error {

	uri := fmt.Sprintf("/v2/clients/%d/triggers/%d", clientID, triggerID)
	_, err := api.makeRequest("DELETE", uri, "trigger", nil)
	if err != nil {
		return err
	}
	return nil
}

// TriggerUpdate updates existing trigger using unique ID.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) TriggerUpdate(triggerBody *TriggerCreate, clientID, triggerID int) (*TriggerResp, error) {

	uri := fmt.Sprintf("/v2/clients/%d/triggers/%d", clientID, triggerID)
	respBody, err := api.makeRequest("PUT", uri, "trigger", triggerBody)
	if err != nil {
		return nil, err
	}
	var t TriggerResp
	if err = json.Unmarshal(respBody, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
