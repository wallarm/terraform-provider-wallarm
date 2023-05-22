package wallarm

import (
	"encoding/json"
	"net/url"
	"strconv"
)

type (
	IPListType string

	IPList interface {
		IPListRead(listType IPListType, clientID int) ([]IPRule, error)
		IPListCreate(clientID int, params IPRuleCreationParams) error
		IPListDelete(listType IPListType, clientID int, ids []int) error
	}

	IPRuleCreationParams struct {
		ExpiredAt int        `json:"expired_at"`
		List      IPListType `json:"list"`
		Pools     []int      `json:"pools"`
		Reason    string     `json:"reason"`
		RuleType  string     `json:"rule_type"`
		Subnet    string     `json:"subnet"`
	}

	IPRule struct {
		ID              int      `json:"id"`
		ClientID        int      `json:"clientid"`
		RuleType        string   `json:"rule_type"`
		List            string   `json:"list"`
		Author          string   `json:"author"`
		CreatedAt       int      `json:"created_at"`
		ExpiredAt       int      `json:"expired_at"`
		Pools           []int    `json:"pools"`
		Reason          string   `json:"reason"`
		AuthorTriggerID int      `json:"author_trigger_id"`
		AuthorUserID    int      `json:"author_user_id"`
		Subnet          string   `json:"subnet"`
		Country         string   `json:"country"`
		ProxyType       string   `json:"proxy_type"`
		Datacenter      string   `json:"datacenter"`
		SourceValues    []string `json:"source_values"`
	}
)

const (
	DenylistType  IPListType = "black"
	AllowlistType IPListType = "white"
	GraylistType  IPListType = "gray"
)

func (api *api) IPListRead(listType IPListType, clientID int) ([]IPRule, error) {
	uri := "/v4/ip_rules"

	q := url.Values{}
	q.Set("filter[clientid]", strconv.Itoa(clientID))
	q.Set("filter[list]", string(listType))
	q.Set("limit", "100")
	q.Set("offset", "0")

	var bulkIPRules struct {
		Body struct {
			Objects []IPRule `json:"objects"`
		} `json:"body"`
	}

	result := []IPRule{}
	offset := 0

	for {
		q.Set("offset", strconv.Itoa(offset))

		respBody, err := api.makeRequest("GET", uri, "", q.Encode())
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(respBody, &bulkIPRules); err != nil {
			return nil, err
		}

		result = append(result, bulkIPRules.Body.Objects...)

		if len(bulkIPRules.Body.Objects) < 100 {
			break
		}

		offset += 100
	}

	return result, nil
}

func (api *api) IPListCreate(clientID int, params IPRuleCreationParams) error {
	uri := "/v4/ip_rules"
	reqBody := struct {
		ClientID int                  `json:"clientid"`
		Force    bool                 `json:"force"`
		IPRule   IPRuleCreationParams `json:"ip_rule"`
	}{ClientID: clientID, Force: false, IPRule: params}

	_, err := api.makeRequest("POST", uri, "", &reqBody)

	return err
}

func (api *api) IPListDelete(listType IPListType, clientID int, ids []int) error {
	uri := "/v4/ip_rules"
	reqBody := struct {
		Filter struct {
			ID       []int      `json:"id"`
			ClientID int        `json:"clientid"`
			List     IPListType `json:"list"`
		} `json:"filter"`
	}{}
	reqBody.Filter.ID = ids
	reqBody.Filter.ClientID = clientID
	reqBody.Filter.List = listType

	_, err := api.makeRequest("DELETE", uri, "ip_rules", &reqBody)

	return err
}
