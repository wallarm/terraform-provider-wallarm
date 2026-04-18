package resourcerule

import (
	"bytes"
	"encoding/json"
	"fmt"

	wallarm "github.com/wallarm/wallarm-go"
)

// HashResponseActionDetails is the hash function for the action TypeSet.
// It transforms API point arrays into point maps as a side effect (e.g.,
// ["header","HOST"] → {header: "HOST"}, ["get","key"] → {query: "key"}).
func HashResponseActionDetails(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	var p []interface{}
	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))
	if val, ok := m["point"]; ok {
		p = val.([]interface{})
		switch p[0].(string) {
		case "action_name":
			pointMap := make(map[string]string)
			pointMap["action_name"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "action_ext":
			pointMap := make(map[string]string)
			pointMap["action_ext"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "scheme":
			pointMap := make(map[string]string)
			pointMap["scheme"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "uri":
			pointMap := make(map[string]string)
			pointMap["uri"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "proto":
			pointMap := make(map[string]string)
			pointMap["proto"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "method":
			pointMap := make(map[string]string)
			pointMap["method"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case Path:
			pointMap := make(map[string]string)
			pointMap[Path] = fmt.Sprintf("%d", int(p[1].(float64)))
			m["point"] = pointMap
		case "instance":
			pointMap := make(map[string]string)
			pointMap["instance"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
			m["type"] = ""
		case pointKeyHeader:
			pointMap := make(map[string]string)
			pointMap[pointKeyHeader] = p[1].(string)
			m["point"] = pointMap
		case pointKeyGet:
			pointMap := make(map[string]string)
			pointMap["query"] = p[1].(string)
			m["point"] = pointMap
		}

		buf.WriteString(fmt.Sprintf("%v-", m["point"]))
	}
	return HashString(buf.String())
}

// HashActionDetails is a pure hash function for the action TypeSet that works
// on both API-format (pre-transform) and config-format (post-transform) data.
// Unlike HashResponseActionDetails, it has NO side effects on the input map.
// For instance conditions, type "" and "equal" are normalized to the same value
// so that config (type="" / omitted) and state (type="equal") produce the same
// hash. Other types like "regex" are preserved to detect actual type changes.
func HashActionDetails(v interface{}) int {
	m := v.(map[string]interface{})
	var buf bytes.Buffer

	condType := m["type"].(string)
	value := m["value"].(string)

	// Detect point format: []interface{} (API) or map (config/transformed).
	var pointStr string
	if val, ok := m["point"]; ok {
		switch p := val.(type) {
		case []interface{}:
			// API format — compute what the transformed point map would be.
			pointMap := make(map[string]string)
			key := p[0].(string)
			switch key {
			case "action_name", "action_ext", "scheme", "uri", "proto", "method":
				pointMap[key] = value
			case Path:
				pointMap[Path] = fmt.Sprintf("%d", int(p[1].(float64)))
			case "instance":
				pointMap["instance"] = value
				condType = normalizeInstanceType(condType)
			case pointKeyHeader:
				pointMap[pointKeyHeader] = p[1].(string)
			case pointKeyGet:
				pointMap["query"] = p[1].(string)
			}
			pointStr = fmt.Sprintf("%v", pointMap)
		case map[string]string:
			if _, isInstance := p["instance"]; isInstance {
				condType = normalizeInstanceType(condType)
			}
			pointStr = fmt.Sprintf("%v", p)
		case map[string]interface{}:
			// Config format from Terraform SDK.
			if _, isInstance := p["instance"]; isInstance {
				condType = normalizeInstanceType(condType)
			}
			pointStr = fmt.Sprintf("%v", p)
		}
	}

	buf.WriteString(fmt.Sprintf("%s-", condType))
	buf.WriteString(fmt.Sprintf("%s-", value))
	if pointStr != "" {
		buf.WriteString(fmt.Sprintf("%s-", pointStr))
	}
	return HashString(buf.String())
}

// normalizeInstanceType normalizes the default instance types ("" and "equal")
// to "" for hashing, so omitted type and API-returned "equal" produce the same
// hash. Non-default types like "regex" are preserved to detect actual changes.
func normalizeInstanceType(t string) string {
	if t == "" || t == "equal" {
		return ""
	}
	return t
}

// TransformAPIActionToSchema transforms an API-format action map (point as
// []interface{}) into the Terraform config format (point as map[string]string).
// Unlike HashResponseActionDetails, this is a pure transform with no hash
// computation. It performs the same mutations that HashResponseActionDetails
// applies as side effects.
func TransformAPIActionToSchema(m map[string]interface{}) {
	val, ok := m["point"]
	if !ok {
		return
	}
	p, ok := val.([]interface{})
	if !ok {
		return // Already transformed (map format) — nothing to do.
	}
	pointMap := make(map[string]string)
	switch p[0].(string) {
	case "action_name":
		pointMap["action_name"] = m["value"].(string)
		m["value"] = ""
	case "action_ext":
		pointMap["action_ext"] = m["value"].(string)
		m["value"] = ""
	case "scheme":
		pointMap["scheme"] = m["value"].(string)
		m["value"] = ""
	case "uri":
		pointMap["uri"] = m["value"].(string)
		m["value"] = ""
	case "proto":
		pointMap["proto"] = m["value"].(string)
		m["value"] = ""
	case "method":
		pointMap["method"] = m["value"].(string)
		m["value"] = ""
	case Path:
		pointMap[Path] = fmt.Sprintf("%d", int(p[1].(float64)))
	case "instance":
		pointMap["instance"] = m["value"].(string)
		m["value"] = ""
	case pointKeyHeader:
		pointMap[pointKeyHeader] = p[1].(string)
	case pointKeyGet:
		pointMap["query"] = p[1].(string)
	}
	m["point"] = pointMap
}

// TODO: add unit test — valid ActionDetails, nil value handling, value always present in output
// ActionDetailsToMap converts an API ActionDetails struct to a Terraform-compatible map
// via JSON marshal/unmarshal. Ensures "value" key is always present.
func ActionDetailsToMap(actionDetails wallarm.ActionDetails) (map[string]interface{}, error) {
	jsonActions, err := json.Marshal(actionDetails)
	if err != nil {
		return nil, err
	}
	var mapActions map[string]interface{}
	if err = json.Unmarshal(jsonActions, &mapActions); err != nil {
		return nil, err
	}
	if v, ok := mapActions["value"]; !ok || v == nil {
		mapActions["value"] = ""
	}
	return mapActions, nil
}
