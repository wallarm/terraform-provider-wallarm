package resourcerule

import (
	"encoding/json"

	wallarm "github.com/wallarm/wallarm-go"
)

// ActionMeta contains full action metadata for the .action_meta file
// written inside each action directory.
type ActionMeta struct {
	ActionID         *int              `json:"action_id"`
	ClientID         int               `json:"client_id"`
	Name             *string           `json:"name"`
	Conditions       []ActionCondition `json:"conditions"`
	ConditionsHash   string            `json:"conditions_hash"`
	DirName          string            `json:"dir_name"`
	EndpointPath     *string           `json:"endpoint_path"`
	EndpointDomain   *string           `json:"endpoint_domain"`
	EndpointInstance *string           `json:"endpoint_instance"`
	UpdatedAt        *int              `json:"updated_at"`
}

// ActionCondition is a simplified condition representation for the meta file.
type ActionCondition struct {
	Type  string      `json:"type"`
	Point interface{} `json:"point"`
	Value interface{} `json:"value"`
}

// NewActionMeta builds an ActionMeta from API action data and computed fields.
func NewActionMeta(
	actionID *int,
	clientID int,
	name *string,
	conditions []wallarm.ActionDetails,
	endpointPath, endpointDomain, endpointInstance *string,
	updatedAt *int,
) ActionMeta {
	conds := make([]ActionCondition, len(conditions))
	for i, c := range conditions {
		conds[i] = ActionCondition{
			Type:  c.Type,
			Point: c.Point,
			Value: c.Value,
		}
	}

	return ActionMeta{
		ActionID:         actionID,
		ClientID:         clientID,
		Name:             name,
		Conditions:       conds,
		ConditionsHash:   ConditionsHash(conditions),
		DirName:          ActionDirName(conditions),
		EndpointPath:     endpointPath,
		EndpointDomain:   endpointDomain,
		EndpointInstance: endpointInstance,
		UpdatedAt:        updatedAt,
	}
}

// FormatActionMeta serializes an ActionMeta to pretty-printed JSON.
func FormatActionMeta(meta ActionMeta) ([]byte, error) {
	return json.MarshalIndent(meta, "", "  ")
}
