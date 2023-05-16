package wallarm

import (
	"encoding/json"
)

type (
	// Vulnerability contains operations available on Vulnerability resource
	Vulnerability interface {
		GetVulnRead(getVulnBody *GetVulnRead) (*GetVulnReadResp, error)
	}

	// GetVulnRead is a root object for requesting vulnerabilities.
	// Limit is a number between 0 - 1000
	// Offset is a number between 0 - 1000
	GetVulnRead struct {
		Filter    *GetVulnFilter `json:"filter"`
		Limit     int            `json:"limit"`
		Offset    int            `json:"offset"`
		OrderBy   string         `json:"order_by"`
		OrderDesc bool           `json:"order_desc"`
	}

	// GetVulnFilter is used to filter the vulnerability status
	// Possible values: "active", "closed", "falsepositive"
	GetVulnFilter struct {
		Status string `json:"status"`
	}

	// GetVulnReadResp is the response on the inquiry of vulnerabilities by filter
	GetVulnReadResp struct {
		Status int `json:"status"`
		Body   []struct {
			ValidateTime   int         `json:"validate_time"`
			InvalidateTime interface{} `json:"invalidate_time"`
			LastCheck      interface{} `json:"last_check"`
			Incidents      interface{} `json:"incidents"`
			ID             int         `json:"id"`
			Wid            string      `json:"wid"`
			Template       string      `json:"template"`
			Status         string      `json:"status"`
			Target         string      `json:"target"`
			Type           string      `json:"type"`
			Threat         int         `json:"threat"`
			Clientid       int         `json:"clientid"`
			TestrunID      interface{} `json:"testrun_id"`
			TicketStatus   interface{} `json:"ticket_status"`
			TicketHistory  []struct {
				Time    int         `json:"time"`
				Type    string      `json:"type"`
				Message string      `json:"message"`
				Link    interface{} `json:"link"`
			} `json:"ticket_history"`
			Method         string `json:"method"`
			Domain         string `json:"domain"`
			Path           string `json:"path"`
			Parameter      string `json:"parameter"`
			Title          string `json:"title"`
			Description    string `json:"description"`
			Additional     string `json:"additional"`
			ExploitExample string `json:"exploit_example"`
			Filter         []struct {
				Method    string `json:"method"`
				Domain    string `json:"domain"`
				Path      string `json:"path"`
				Parameter string `json:"parameter"`
			} `json:"filter"`
			Validated       bool        `json:"validated"`
			Hidden          bool        `json:"hidden"`
			DetectionMethod string      `json:"detection_method"`
			VulnRecheckType interface{} `json:"vuln_recheck_type"`
			TemplateParams  struct {
				Method         string `json:"method"`
				Domain         string `json:"domain"`
				Path           string `json:"path"`
				Parameter      string `json:"parameter"`
				ExploitExample string `json:"exploit_example"`
			} `json:"template_params,omitempty"`
		} `json:"body"`
	}
)

// GetVulnRead is used to get `Vulnerabilities` by filter in a body.
// It returns all requested vulnerabilities but not more than 1000
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) GetVulnRead(getVulnBody *GetVulnRead) (*GetVulnReadResp, error) {

	uri := "/v1/objects/vuln"
	respBody, err := api.makeRequest("POST", uri, "vuln", getVulnBody)
	if err != nil {
		return nil, err
	}
	var v GetVulnReadResp
	if err = json.Unmarshal(respBody, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
