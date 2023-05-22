package wallarm

import (
	"encoding/json"
	"fmt"
)

type (

	// Action contains operations available on Action resource
	Action interface {
		HintRead(hintBody *HintRead) (*HintReadResp, error)
		RuleRead(ruleBody *ActionRead) (*ActionFetch, error)
		HintCreate(ruleBody *ActionCreate) (*ActionCreateResp, error)
		RuleDelete(actionID int) error
		HintDelete(hintbody *HintDelete) error
	}

	// ActionDetails defines the Action of how to parse the request.
	// Point represents a part of the request where the condition should be satisfied.
	// ActionDetails is used to define the particular assets of the Action field.
	ActionDetails struct {
		Type  string        `json:"type,omitempty"`
		Point []interface{} `json:"point,omitempty"`
		Value interface{}   `json:"value,omitempty"`
	}

	// ActionCreate is a creation skeleton for the Rule.
	ActionCreate struct {
		Type       string              `json:"type"`
		Action     *[]ActionDetails    `json:"action,omitempty"`
		Clientid   int                 `json:"clientid,omitempty"`
		Validated  bool                `json:"validated"`
		Point      TwoDimensionalSlice `json:"point,omitempty"`
		Rules      []string            `json:"rules,omitempty"`
		AttackType string              `json:"attack_type,omitempty"`
		Mode       string              `json:"mode,omitempty"`
		Counter    string              `json:"counter,omitempty"`
		Regex      string              `json:"regex,omitempty"`
		RegexID    int                 `json:"regex_id,omitempty"`
		Enabled    *bool               `json:"enabled,omitempty"`
		Name       string              `json:"name,omitempty"`
		Values     []string            `json:"values,omitempty"`
		Comment    string              `json:"comment,omitempty"`
		FileType   string              `json:"file_type,omitempty"`
		Parser     string              `json:"parser,omitempty"`
		State      string              `json:"state,omitempty"`
		VarType    string              `json:"var_type,omitempty"`
	}

	// ActionFilter is the specific filter for getting the rules.
	// This is an inner structure.
	ActionFilter struct {
		ID       []int    `json:"id,omitempty"`
		NotID    []int    `json:"!id,omitempty"`
		Clientid []int    `json:"clientid,omitempty"`
		HintType []string `json:"hint_type,omitempty"`
	}

	// TwoDimensionalSlice is used for Point and HintsCount structures.
	TwoDimensionalSlice [][]interface{}

	// ActionRead is used as a filter to fetch the rules.
	ActionRead struct {
		Filter *ActionFilter `json:"filter"`
		Limit  int           `json:"limit"`
		Offset int           `json:"offset"`
	}

	// ActionFetch is a response struct which portrays
	// all conditions set for requests of filtered type.
	ActionFetch struct {
		Status int `json:"status"`
		Body   []struct {
			ID                int           `json:"id"`
			Clientid          int           `json:"clientid"`
			Name              interface{}   `json:"name"`
			Conditions        []interface{} `json:"conditions"`
			Hints             int           `json:"hints"`
			GroupedHintsCount int           `json:"grouped_hints_count"`
			UpdatedAt         int           `json:"updated_at"`
		} `json:"body"`
	}

	// ActionBody is an inner body for the Action and Hint responses.
	ActionBody struct {
		ID           int             `json:"id"`
		ActionID     int             `json:"actionid"`
		Clientid     int             `json:"clientid"`
		Action       []ActionDetails `json:"action"`
		CreateTime   int             `json:"create_time"`
		CreateUserid int             `json:"create_userid"`
		Validated    bool            `json:"validated"`
		System       bool            `json:"system"`
		RegexID      interface{}     `json:"regex_id"`
		UpdatedAt    int             `json:"updated_at"`
		Type         string          `json:"type"`
		Enabled      bool            `json:"enabled"`
		Mode         string          `json:"mode"`
		Regex        string          `json:"regex"`
		Point        []interface{}   `json:"point"`
		AttackType   string          `json:"attack_type"`
		Rules        []string        `json:"rules"`
		Counter      string          `json:"counter,omitempty"`
		VarType      string          `json:"var_type"`
		// Headers for the Set Response Headers Rule
		// are defined by these two parameters.
		Name   string        `json:"name"`
		Values []interface{} `json:"values"`
	}

	// ActionCreateResp is the response of just created Rule.
	ActionCreateResp struct {
		Status int         `json:"status"`
		Body   *ActionBody `json:"body"`
	}

	// HintReadResp is the response of filtered rules by Action ID.
	HintReadResp struct {
		Status int           `json:"status"`
		Body   *[]ActionBody `json:"body"`
	}

	// HintRead is used to define whether action of the rule exists.
	HintRead struct {
		Filter    *HintFilter `json:"filter"`
		OrderBy   string      `json:"order_by"`
		OrderDesc bool        `json:"order_desc"`
		Limit     int         `json:"limit"`
		Offset    int         `json:"offset"`
	}

	// HintFilter is used as a filter by Action ID.
	HintFilter struct {
		Clientid        []int    `json:"clientid,omitempty"`
		ActionID        []int    `json:"actionid,omitempty"`
		ID              []int    `json:"id,omitempty"`
		NotID           []int    `json:"!id,omitempty"`
		NotActionID     []int    `json:"!actionid,omitempty"`
		CreateUserid    []int    `json:"create_userid,omitempty"`
		NotCreateUserid []int    `json:"!create_userid,omitempty"`
		CreateTime      [][]int  `json:"create_time,omitempty"`
		NotCreateTime   [][]int  `json:"!create_time,omitempty"`
		System          bool     `json:"system,omitempty"`
		Type            []string `json:"type,omitempty"`
	}

	// HintDelete is used for removal of Rule by Hint ID.
	HintDelete struct {
		Filter *HintDeleteFilter `json:"filter"`
	}

	// HintDeleteFilter is used as a filter by Hint ID.
	HintDeleteFilter struct {
		Clientid []int `json:"clientid"`
		ID       int   `json:"id"`
	}
)

// HintRead reads the Rules defined by Action ID.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) HintRead(hintBody *HintRead) (*HintReadResp, error) {

	uri := "/v1/objects/hint"
	respBody, err := api.makeRequest("POST", uri, "hint", hintBody)
	if err != nil {
		return nil, err
	}
	var h HintReadResp
	if err = json.Unmarshal(respBody, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

// RuleRead reads the Rules defined by a filter.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) RuleRead(ruleBody *ActionRead) (*ActionFetch, error) {

	uri := "/v1/objects/action"
	respBody, err := api.makeRequest("POST", uri, "rule", ruleBody)
	if err != nil {
		return nil, err
	}
	var a ActionFetch
	if err = json.Unmarshal(respBody, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// HintCreate creates Rules in Wallarm Cloud.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) HintCreate(ruleBody *ActionCreate) (*ActionCreateResp, error) {

	uri := "/v1/objects/hint/create"
	respBody, err := api.makeRequest("POST", uri, "rule", ruleBody)
	if err != nil {
		return nil, err
	}
	var a ActionCreateResp
	if err = json.Unmarshal(respBody, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// RuleDelete deletes the Rule defined by unique ID.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) RuleDelete(actionID int) error {

	uri := fmt.Sprintf("/v2/action/%d", actionID)
	_, err := api.makeRequest("DELETE", uri, "rule", nil)
	if err != nil {
		return err
	}
	return nil
}

// HintDelete deletes the Rule defined by the unique Hint ID.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) HintDelete(hintbody *HintDelete) error {

	uri := "/v1/objects/hint/delete"
	_, err := api.makeRequest("POST", uri, "hint", hintbody)
	if err != nil {
		return err
	}
	return nil
}
