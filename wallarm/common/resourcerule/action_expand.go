package resourcerule

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

// TODO: add unit test — header/iequal, path/equal, instance, query, action_name, absent, empty set
// nolint
func ExpandSetToActionDetailsList(action *schema.Set) ([]wallarm.ActionDetails, error) {
	var as []wallarm.ActionDetails
	for _, actionMap := range action.List() {
		// Derive maps consecutively from a Set List
		actionMap := actionMap.(map[string]interface{})

		// Make keys of map sorted to
		// then iterate over a map in order
		keys := make([]string, 0, len(actionMap))
		for k := range actionMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		a := wallarm.ActionDetails{}
		for _, k := range keys {
			switch k {
			case "point":
				point := actionMap[k].(map[string]interface{})
				for pointKey, pointValue := range point {
					switch pointKey {
					case Path:
						// Marshalling of the number leads to float64 even though it was int initially
						// Therefore, we parse string into float64 to compare structs properly afterwards
						pointValue, err := strconv.ParseFloat(pointValue.(string), 64)
						if err != nil {
							return nil, err
						}
						a.Point = []interface{}{pointKey, pointValue}
					case "action_name", "action_ext", "method",
						"proto", "scheme", "uri":
						a.Point = []interface{}{pointKey}
						// This is required by the API when case is insensitive
						switch {
						case actionMap["type"] == Iequal:
							a.Value = strings.ToLower(pointValue.(string))
						case actionMap["type"] == "absent":
							a.Value = nil
						default:
							a.Value = pointValue.(string)
						}
					case "instance":
						a.Point = []interface{}{pointKey}
						a.Value = pointValue.(string)
						a.Type = "equal" // default; overridden by the "type" key if explicitly set
					case Header:
						// This is required by the API when a header field is specified
						a.Point = []interface{}{pointKey, strings.ToUpper(pointValue.(string))}
					case "query":
						// This is required by the API when case is insensitive
						if actionMap["type"] == Iequal {
							a.Point = []interface{}{"get", strings.ToLower(pointValue.(string))}
						} else {
							a.Point = []interface{}{"get", pointValue.(string)}
						}
					default:
						// This is required by the API when case is insensitive
						if actionMap["type"] == "iequal" {
							a.Point = []interface{}{pointKey, strings.ToLower(pointValue.(string))}
						} else {
							a.Point = []interface{}{pointKey, pointValue.(string)}
						}
					}
				}

			case "type":
				// Fill out only when it is presented
				// Then default values will be omitted in the JSON request body
				// Otherwise, the API returns 4xx back due to the incorrect schema
				if actionMap[k].(string) != "" {
					a.Type = actionMap[k].(string)
				}
			case "value":
				if actionMap[k].(string) != "" {
					if actionMap["type"] == "iequal" {
						a.Value = strings.ToLower(actionMap[k].(string))
					} else {
						a.Value = actionMap[k].(string)
					}
				}
			}
		}

		// Check if there is anything to append, ensure it's not a default branch
		if a.Type != "" {
			as = append(as, a)
		}
	}
	// Check if this is for a default branch
	if len(as) == 0 {
		as = []wallarm.ActionDetails{}
	}
	return as, nil
}

// TODO: add unit test — paired element, simple element, integer index, multi-level chain, empty
// WrapPointElements converts a flat API point array into a 2D string slice
// for the Terraform point schema. 2-part elements (hash, header, get, form_urlencoded, etc.)
// consume the next element as their value; 1-part elements (post, json_doc, uri, etc.) stand alone.
func WrapPointElements(input []interface{}) [][]string {
	var result [][]string // This will store the final result as a 2D slice of strings
	i := 0

	for i < len(input) {
		switch input[i] {
		// Paired point types — consume the next element as key/index.
		// Keep in sync with TYPES_INFO in proton/types.rb (all except simple:true).
		case
			// Core
			"hash", "array", "json", "json_obj", "json_array",
			// HTTP
			pointKeyHeader, "cookie", pointKeyGet, "path", "multipart",
			"form_urlencoded", "content_disp", "response_header",
			// XML
			"xml_pi", "xml_dtd_entity", "xml_tag_array", "xml_tag",
			"xml_attr", "xml_comment",
			// JWT / Protobuf / gRPC
			"jwt", "grpc", "protobuf",
			// ViewState
			"viewstate_array", "viewstate_pair", "viewstate_triplet",
			"viewstate_dict", "viewstate_sparse_array",
			// GraphQL
			"gql_query", "gql_mutation", "gql_subscription", "gql_fragment",
			"gql_dir", "gql_spread", "gql_type", "gql_var":
			// Check if there is a next element to include
			if i+1 < len(input) {
				// Convert both elements to strings and wrap them in a slice of strings
				result = append(result, []string{
					fmt.Sprintf("%v", input[i]),
					fmt.Sprintf("%v", input[i+1]),
				})
				i++ // Skip the next element as it's already included
			} else {
				// If no next element, still wrap the special case string alone
				result = append(result, []string{fmt.Sprintf("%v", input[i])})
			}
		default:
			// For regular elements, convert to string and wrap it in a slice of strings
			result = append(result, []string{fmt.Sprintf("%v", input[i])})
		}
		i++ // Move to the next element
	}

	return result
}

// TODO: add unit test — numeric types to float64, non-numeric passthrough, empty
// ExpandPointsToTwoDimensionalArray converts the Terraform point schema (list of lists of strings)
// to the API TwoDimensionalSlice format. Numeric-value point types (path, array, etc.) are
// converted from string to float64.
func ExpandPointsToTwoDimensionalArray(ps []interface{}) (wallarm.TwoDimensionalSlice, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	points := make(wallarm.TwoDimensionalSlice, len(ps))
	for i, point := range ps {
		pointSlice := point.([]interface{})
		switch pointSlice[0] {
		case "path", "array", "grpc", "json_array", "xml_comment",
			"xml_dtd_entity", "xml_pi", "xml_tag_array":
			// Align to the []string{} schema, float is used since marshalling considers numbers as float64
			if len(pointSlice) > 1 {
				number, err := strconv.ParseFloat(pointSlice[1].(string), 64)
				if err != nil {
					return nil, err
				}
				pointSlice[1] = number
				points[i] = pointSlice
			}
		default:
			points[i] = pointSlice
		}
	}
	return points, nil
}
