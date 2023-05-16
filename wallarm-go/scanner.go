package wallarm

import (
	"encoding/json"
	"fmt"
)

type (
	// Scanner contains operations available on Scanner resource
	Scanner interface {
		ScannerCreate(scannerBody *ScannerCreate) (*ScannerCreateBody, error)
		ScannerDelete(scannerBody *ScannerDelete, resType string) error
		ScannerUpdate(scannerBody *ScannerUpdate, resType string, resID int) error
	}

	// ScannerCreate is a request query to put in a new resource.
	ScannerCreate struct {
		Query    string `json:"query"`
		Clientid int    `json:"clientid"`
	}

	// ScannerCreateBody is a response on the Create action.
	ScannerCreateBody struct {
		Body struct {
			Result  string `json:"result"`
			Objects []struct {
				Rps                     interface{} `json:"rps"`
				ID                      int         `json:"id"`
				IP                      string      `json:"ip"`
				Domain                  string      `json:"domain"`
				New                     bool        `json:"new"`
				Datacenter              interface{} `json:"datacenter"`
				Disabled                bool        `json:"disabled"`
				CreatedAt               string      `json:"created_at"`
				UpdatedAt               string      `json:"updated_at"`
				LastDisabled            interface{} `json:"last_disabled"`
				Group                   bool        `json:"group"`
				ParentID                int         `json:"parent_id"`
				Clientid                int         `json:"clientid"`
				EnabledDomainBinds      int         `json:"enabled_domain_binds"`
				DisabledDomainBinds     int         `json:"disabled_domain_binds"`
				EnabledServiceBinds     int         `json:"enabled_service_binds"`
				DisabledServiceBinds    int         `json:"disabled_service_binds"`
				DiscoveredAutomatically bool        `json:"discovered_automatically"`
				DisabledEdge            interface{} `json:"disabled_edge"`
			} `json:"objects"`
		} `json:"body"`
	}

	// ScannerDelete is used to delete scanner resources in bulk.
	ScannerDelete struct {
		Bulk *[]ScannerDeleteBulk `json:"bulk"`
	}

	// ScannerFilter is used as a filter for delete query.
	ScannerFilter struct {
		*ScannerCreate
		ID []int `json:"id"`
	}

	// ScannerDeleteBulk is used to update scope resource.
	ScannerDeleteBulk struct {
		Filter *ScannerFilter `json:"filter"`
	}

	// ScannerUpdate is used to update scope resource.
	// The only field is disabling of resource so far.
	ScannerUpdate struct {
		Disabled bool `json:"disabled"`
		Clientid int  `json:"clientid"`
	}
)

// ScannerCreate creates a new resource to scan.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) ScannerCreate(scannerBody *ScannerCreate) (*ScannerCreateBody, error) {

	uri := "/v2/scope/new"
	res, err := api.makeRequest("PUT", uri, "scanner", scannerBody)
	if err != nil {
		return nil, err
	}

	var data ScannerCreateBody
	if err = json.Unmarshal(res, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// ScannerDelete deletes resources which have been created previously.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) ScannerDelete(scannerBody *ScannerDelete, resType string) error {

	uri := fmt.Sprintf("/v2/scope/%s/bulk", resType)
	_, err := api.makeRequest("POST", uri, "", scannerBody)
	if err != nil {
		return err
	}
	return nil
}

// ScannerUpdate updates resources which have been created in prior.
// For example, you may disable the resource.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) ScannerUpdate(scannerBody *ScannerUpdate, resType string, resID int) error {

	uri := fmt.Sprintf("/v2/scope/%s/%d", resType, resID)
	_, err := api.makeRequest("POST", uri, "scanner", scannerBody)
	if err != nil {
		return err
	}
	return nil
}
