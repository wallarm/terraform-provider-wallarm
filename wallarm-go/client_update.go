package wallarm

import "encoding/json"

type (
	// Client contains operations available on Client resource
	Client interface {
		ClientUpdate(clientBody *ClientUpdate) (*ClientInfo, error)
		ClientRead(clientBody *ClientRead) (*ClientInfo, error)
	}

	// ClientFields defines fields which are subject to update.
	ClientFields struct {
		ScannerMode         string `json:"scanner_mode,omitempty"`
		AttackRecheckerMode string `json:"attack_rechecker_mode,omitempty"`
	}

	// ClientFilter is used for filtration.
	// ID is a Client ID entity.
	ClientFilter struct {
		ID int `json:"id"`
	}

	// ClientUpdate is a root object for updating.
	ClientUpdate struct {
		Filter *ClientFilter `json:"filter"`
		Fields *ClientFields `json:"fields"`
	}

	// ClientRead is used for filtration of Client Info.
	ClientRead struct {
		Filter *ClientReadFilter `json:"filter"`
		Limit  int               `json:"limit"`
		Offset int               `json:"offset"`
	}

	// ClientReadFilter is the inner object for Filter.
	ClientReadFilter struct {
		ClientFilter
		Enabled bool   `json:"enabled,omitempty"`
		Name    string `json:"name,omitempty"`
	}

	// ClientInfo is the response on the Client Read.
	// It shows the common information about the client.
	ClientInfo struct {
		Status int `json:"status"`
		Body   []struct {
			ClientFilter
			Name             string   `json:"name"`
			Components       []string `json:"components"`
			VulnPrefix       string   `json:"vuln_prefix"`
			SupportPlan      string   `json:"support_plan"`
			DateFormat       string   `json:"date_format"`
			BlockingType     string   `json:"blocking_type"`
			ScannerMode      string   `json:"scanner_mode"`
			QratorBlacklists bool     `json:"qrator_blacklists"`
			Notifications    struct {
				ReportDaily struct {
					Email     []interface{} `json:"email"`
					Telegram  []interface{} `json:"telegram"`
					Slack     []interface{} `json:"slack"`
					Splunk    []interface{} `json:"splunk"`
					PagerDuty []interface{} `json:"pager_duty"`
				} `json:"report_daily"`
				ReportWeekly struct {
					Email     []interface{} `json:"email"`
					Telegram  []interface{} `json:"telegram"`
					Slack     []interface{} `json:"slack"`
					Splunk    []interface{} `json:"splunk"`
					PagerDuty []interface{} `json:"pager_duty"`
				} `json:"report_weekly"`
				ReportMonthly struct {
					Email     []interface{} `json:"email"`
					Telegram  []interface{} `json:"telegram"`
					Slack     []interface{} `json:"slack"`
					Splunk    []interface{} `json:"splunk"`
					PagerDuty []interface{} `json:"pager_duty"`
				} `json:"report_monthly"`
				System struct {
					Email     []interface{} `json:"email"`
					Telegram  []interface{} `json:"telegram"`
					Slack     []interface{} `json:"slack"`
					Splunk    []interface{} `json:"splunk"`
					PagerDuty []interface{} `json:"pager_duty"`
				} `json:"system"`
				Vuln struct {
					Email     []interface{} `json:"email"`
					Telegram  []interface{} `json:"telegram"`
					Slack     []interface{} `json:"slack"`
					Splunk    []interface{} `json:"splunk"`
					PagerDuty []interface{} `json:"pager_duty"`
				} `json:"vuln"`
				Scope struct {
					Email     []interface{} `json:"email"`
					Telegram  []interface{} `json:"telegram"`
					Slack     []interface{} `json:"slack"`
					Splunk    []interface{} `json:"splunk"`
					PagerDuty []interface{} `json:"pager_duty"`
				} `json:"scope"`
			} `json:"notifications"`
			LastScan            interface{} `json:"last_scan"`
			ScannerCluster      string      `json:"scanner_cluster"`
			ScannerScopeCluster string      `json:"scanner_scope_cluster"`
			ScannerState        struct {
				LastScan      int         `json:"last_scan"`
				LastVuln      int         `json:"last_vuln"`
				LastVulnCheck interface{} `json:"last_vuln_check"`
				LastWapi      interface{} `json:"last_wapi"`
			} `json:"scanner_state"`
			Language            string `json:"language"`
			AttackRecheckerMode string `json:"attack_rechecker_mode"`
			VulnRecheckerMode   string `json:"vuln_rechecker_mode"`
			Validated           bool   `json:"validated"`
			Enabled             bool   `json:"enabled"`
			CreateAt            int    `json:"create_at"`
			Partnerid           int    `json:"partnerid"`
			CanEnableBlacklist  bool   `json:"can_enable_blacklist"`
			BlacklistDisabledAt int    `json:"blacklist_disabled_at"`
			HiddenVulns         bool   `json:"hidden_vulns"`
			ScannerPriority     string `json:"scanner_priority"`
		} `json:"body"`
	}
)

// ClientUpdate changes client state.
// It can be used with global Scanner, Attack Rechecker Statuses.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) ClientUpdate(clientBody *ClientUpdate) (*ClientInfo, error) {

	uri := "/v1/objects/client/update"
	respBody, err := api.makeRequest("POST", uri, "client", clientBody)
	if err != nil {
		return nil, err
	}
	var c ClientInfo
	if err = json.Unmarshal(respBody, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// ClientRead requests common info about the account.
// There is info about Scanner, Attack Rechecker, and others.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) ClientRead(clientBody *ClientRead) (*ClientInfo, error) {

	uri := "/v1/objects/client"
	respBody, err := api.makeRequest("POST", uri, "client", clientBody)
	if err != nil {
		return nil, err
	}
	var c ClientInfo
	if err = json.Unmarshal(respBody, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
