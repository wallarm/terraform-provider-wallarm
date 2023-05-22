package wallarm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type (
	// Integration contains operations available on Integration resource
	Integration interface {
		IntegrationCreate(integrationBody *IntegrationCreate) (*IntegrationCreateResp, error)
		IntegrationUpdate(integrationBody *IntegrationCreate, integrationID int) (*IntegrationCreateResp, error)
		IntegrationRead(clientID int, id int) (*IntegrationObject, error)
		IntegrationDelete(integrationID int) error
		IntegrationWithAPICreate(integrationBody *IntegrationWithAPICreate) (*IntegrationCreateResp, error)
		IntegrationWithAPIUpdate(integrationBody *IntegrationWithAPICreate, integrationID int) (*IntegrationCreateResp, error)
		EmailIntegrationCreate(emailBody *EmailIntegrationCreate) (*IntegrationCreateResp, error)
		EmailIntegrationUpdate(integrationBody *EmailIntegrationCreate, integrationID int) (*IntegrationCreateResp, error)
	}

	// IntegrationEvents represents `Events` object while creating a new integration.
	// Event possible values: "hit", "vuln_high", "vuln_medium", "vuln_low", "system", "scope".
	// If `IntegrationObject.Type` is "opsgenie" possible values: "hit", "vuln".
	// `Active` identifies whether the current Event should be reported.
	IntegrationEvents struct {
		Event  string `json:"event"`
		Active bool   `json:"active"`
	}

	// IntegrationObject is an inner object for the Read function containing.
	// ID is a unique identifier of the Integration.
	IntegrationObject struct {
		ID        int         `json:"id"`
		Active    bool        `json:"active"`
		Name      string      `json:"name"`
		Type      string      `json:"type"`
		CreatedAt int         `json:"created_at"`
		CreatedBy string      `json:"created_by"`
		Target    interface{} `json:"target"`
		Events    []struct {
			IntegrationEvents
		} `json:"events"`
	}

	// IntegrationRead is the response on the Read action.
	// This is used for correct Unmarshalling of the response as a container.
	IntegrationRead struct {
		Body struct {
			Result string               `json:"result"`
			Object *[]IntegrationObject `json:"object"`
		} `json:"body"`
	}

	// IntegrationCreate defines how to configure Integration.
	// `Type` possible values: "insight_connect", "opsgenie", "slack",
	//  "pager_duty", "splunk", "sumo_logic"
	IntegrationCreate struct {
		Name     string               `json:"name"`
		Active   bool                 `json:"active"`
		Target   string               `json:"target"`
		Events   *[]IntegrationEvents `json:"events"`
		Type     string               `json:"type"`
		Clientid int                  `json:"clientid,omitempty"`
	}

	// IntegrationCreateResp represents successful creating of
	// an integration entity with the associative parameters.
	IntegrationCreateResp struct {
		Body struct {
			Result            string `json:"result"`
			IntegrationObject `json:"object"`
		} `json:"body"`
	}

	// IntegrationWithAPITarget is used to create an Integration with the following parameters.
	// On purpose to fulfil a custom Webhooks integration.
	IntegrationWithAPITarget struct {
		Token       string                 `json:"token,omitempty"`
		API         string                 `json:"api,omitempty"`
		URL         string                 `json:"url,omitempty"`
		HTTPMethod  string                 `json:"http_method,omitempty"`
		Headers     map[string]interface{} `json:"headers"`
		CaFile      string                 `json:"ca_file"`
		CaVerify    bool                   `json:"ca_verify"`
		Timeout     int                    `json:"timeout,omitempty"`
		OpenTimeout int                    `json:"open_timeout,omitempty"`
	}

	// IntegrationWithAPICreate is a root object of Create action for Integrations.
	// It aids to set `Events` to trigger this integration.
	// `Type` possible values: "web_hooks"
	// `Target` is a struct for a Webhooks endpoint containing params such as URL, Token, etc.
	IntegrationWithAPICreate struct {
		Name     string                    `json:"name"`
		Active   bool                      `json:"active"`
		Target   *IntegrationWithAPITarget `json:"target"`
		Events   *[]IntegrationEvents      `json:"events"`
		Type     string                    `json:"type"`
		Clientid int                       `json:"clientid,omitempty"`
	}

	// EmailIntegrationCreate is a root object of `Create` action for the `email` integration.
	// Temporary workaround (`Target` is a slice instead of string) to not check type many times.
	// Then it will be changed to interface{} with type checking
	EmailIntegrationCreate struct {
		Name     string               `json:"name,omitempty"`
		Active   bool                 `json:"active"`
		Target   []string             `json:"target,omitempty"`
		Events   *[]IntegrationEvents `json:"events,omitempty"`
		Type     string               `json:"type,omitempty"`
		Clientid int                  `json:"clientid,omitempty"`
	}
)

// IntegrationCreate returns create object if Integration
// has been created successfully, otherwise - error.
// It accepts a body with defined settings namely Event types, Name, Target.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) IntegrationCreate(integrationBody *IntegrationCreate) (*IntegrationCreateResp, error) {

	uri := "/v2/integration"
	respBody, err := api.makeRequest("POST", uri, "integration", integrationBody)
	if err != nil {
		return nil, err
	}

	var i IntegrationCreateResp
	if err = json.Unmarshal(respBody, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// IntegrationUpdate is used to Update existing resources.
// It utilises the same format of body as the Create function.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) IntegrationUpdate(integrationBody *IntegrationCreate, integrationID int) (*IntegrationCreateResp, error) {

	uri := fmt.Sprintf("/v2/integration/%d", integrationID)
	respBody, err := api.makeRequest("PUT", uri, "integration", integrationBody)
	if err != nil {
		return nil, err
	}
	var i IntegrationCreateResp
	if err = json.Unmarshal(respBody, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// IntegrationRead is used to read existing integrations.
// It returns the list of Integrations
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) IntegrationRead(clientID int, id int) (*IntegrationObject, error) {

	uri := "/v2/integration"
	q := url.Values{}
	q.Add("clientid", strconv.Itoa(clientID))
	query := q.Encode()
	respBody, err := api.makeRequest("GET", uri, "integration", query)
	if err != nil {
		return nil, err
	}
	var i IntegrationRead
	if err = json.Unmarshal(respBody, &i); err != nil {
		return nil, err
	}
	for _, obj := range *i.Body.Object {
		if obj.ID == id {
			return &obj, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Body: %s", string(respBody)))
}

// IntegrationDelete is used to delete an existing integration.
// If successful, returns nothing.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) IntegrationDelete(integrationID int) error {

	uri := fmt.Sprintf("/v2/integration/%d", integrationID)
	_, err := api.makeRequest("DELETE", uri, "integration", nil)
	if err != nil {
		return err
	}
	return nil
}

// IntegrationWithAPICreate returns created object if an integration
//
//	has been created successfully, otherwise - error.
//
// It accepts defined settings namely Event types, Name, Target.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) IntegrationWithAPICreate(integrationBody *IntegrationWithAPICreate) (*IntegrationCreateResp, error) {

	uri := "/v2/integration"
	respBody, err := api.makeRequest("POST", uri, "integration", integrationBody)
	if err != nil {
		return nil, err
	}

	var i IntegrationCreateResp
	if err = json.Unmarshal(respBody, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// IntegrationWithAPIUpdate is used to Update existing API integration resources.
// It utilises the same format of body as the Create function.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) IntegrationWithAPIUpdate(integrationBody *IntegrationWithAPICreate, integrationID int) (*IntegrationCreateResp, error) {

	uri := fmt.Sprintf("/v2/integration/%d", integrationID)
	respBody, err := api.makeRequest("PUT", uri, "integration", integrationBody)
	if err != nil {
		return nil, err
	}

	var i IntegrationCreateResp
	if err = json.Unmarshal(respBody, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// EmailIntegrationCreate returns created object if the `email` Integration
// has been created successfully, otherwise - error.
// It accepts defined settings namely Event types, Name, Target.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) EmailIntegrationCreate(emailBody *EmailIntegrationCreate) (*IntegrationCreateResp, error) {

	uri := "/v2/integration"
	respBody, err := api.makeRequest("POST", uri, "email", emailBody)
	if err != nil {
		return nil, err
	}

	var i IntegrationCreateResp
	if err = json.Unmarshal(respBody, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// EmailIntegrationUpdate is used to Update existing resources.
// It utilises the same format of body as the Create function.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) EmailIntegrationUpdate(integrationBody *EmailIntegrationCreate, integrationID int) (*IntegrationCreateResp, error) {

	uri := fmt.Sprintf("/v2/integration/%d", integrationID)
	respBody, err := api.makeRequest("PUT", uri, "email", integrationBody)
	if err != nil {
		return nil, err
	}

	var i IntegrationCreateResp
	if err = json.Unmarshal(respBody, &i); err != nil {
		return nil, err
	}
	return &i, nil
}
