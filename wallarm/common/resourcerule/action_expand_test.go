package resourcerule

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

// newActionSet creates a schema.Set with HashActionDetails for testing.
func newActionSet(items ...map[string]interface{}) *schema.Set {
	s := schema.NewSet(HashActionDetails, nil)
	for _, item := range items {
		s.Add(item)
	}
	return s
}

func TestExpandSetToActionDetailsList_Empty(t *testing.T) {
	result, err := ExpandSetToActionDetailsList(newActionSet())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestExpandSetToActionDetailsList_HeaderIequal(t *testing.T) {
	input := newActionSet(map[string]interface{}{
		"type":  "iequal",
		"value": "Example.Com",
		"point": map[string]interface{}{"header": "HOST"},
	})
	result, err := ExpandSetToActionDetailsList(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	r := result[0]
	if r.Type != "iequal" {
		t.Errorf("type: expected iequal, got %s", r.Type)
	}
	if r.Value != "example.com" {
		t.Errorf("value: expected example.com (lowercased), got %s", r.Value)
	}
	if len(r.Point) != 2 || r.Point[0] != "header" || r.Point[1] != "HOST" {
		t.Errorf("point: expected [header HOST], got %v", r.Point)
	}
}

func TestExpandSetToActionDetailsList_PathEqual(t *testing.T) {
	input := newActionSet(map[string]interface{}{
		"type":  "equal",
		"value": "",
		"point": map[string]interface{}{Path: "0"},
	})
	result, err := ExpandSetToActionDetailsList(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	r := result[0]
	if r.Type != "equal" {
		t.Errorf("type: expected equal, got %s", r.Type)
	}
	if len(r.Point) != 2 || r.Point[0] != Path {
		t.Errorf("point: expected [path <float>], got %v", r.Point)
	}
	// Path index is parsed as float64
	if idx, ok := r.Point[1].(float64); !ok || idx != 0.0 {
		t.Errorf("point index: expected 0.0, got %v", r.Point[1])
	}
}

func TestExpandSetToActionDetailsList_Instance(t *testing.T) {
	input := newActionSet(map[string]interface{}{
		"type":  "",
		"value": "",
		"point": map[string]interface{}{"instance": "42"},
	})
	result, err := ExpandSetToActionDetailsList(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	r := result[0]
	if r.Type != "equal" {
		t.Errorf("type: expected equal (default for instance), got %q", r.Type)
	}
	if r.Value != "42" {
		t.Errorf("value: expected 42, got %v", r.Value)
	}
	if len(r.Point) != 1 || r.Point[0] != "instance" {
		t.Errorf("point: expected [instance], got %v", r.Point)
	}
}

func TestExpandSetToActionDetailsList_QueryIequal(t *testing.T) {
	input := newActionSet(map[string]interface{}{
		"type":  "iequal",
		"value": "SearchVal",
		"point": map[string]interface{}{"query": "Q"},
	})
	result, err := ExpandSetToActionDetailsList(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	r := result[0]
	// Query maps to "get" in API
	if len(r.Point) != 2 || r.Point[0] != "get" {
		t.Errorf("point: expected [get <lowercased>], got %v", r.Point)
	}
	if r.Point[1] != "q" {
		t.Errorf("point[1]: expected q (lowercased), got %v", r.Point[1])
	}
}

func TestExpandSetToActionDetailsList_ActionName(t *testing.T) {
	input := newActionSet(map[string]interface{}{
		"type":  "equal",
		"value": "",
		"point": map[string]interface{}{"action_name": "login"},
	})
	result, err := ExpandSetToActionDetailsList(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	r := result[0]
	if r.Value != "login" {
		t.Errorf("value: expected login (moved from point), got %v", r.Value)
	}
	if len(r.Point) != 1 || r.Point[0] != "action_name" {
		t.Errorf("point: expected [action_name], got %v", r.Point)
	}
}

func TestExpandSetToActionDetailsList_Absent(t *testing.T) {
	input := newActionSet(map[string]interface{}{
		"type":  "absent",
		"value": "",
		"point": map[string]interface{}{"action_ext": ""},
	})
	result, err := ExpandSetToActionDetailsList(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	r := result[0]
	if r.Type != "absent" {
		t.Errorf("type: expected absent, got %s", r.Type)
	}
	if r.Value != nil {
		t.Errorf("value: expected nil for absent, got %v", r.Value)
	}
}

func TestWrapPointElements_Paired(t *testing.T) {
	input := []interface{}{"header", "HOST"}
	result := WrapPointElements(input)
	expected := [][]string{{"header", "HOST"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWrapPointElements_Simple(t *testing.T) {
	input := []interface{}{"post"}
	result := WrapPointElements(input)
	expected := [][]string{{"post"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWrapPointElements_IntegerIndex(t *testing.T) {
	input := []interface{}{"path", 0}
	result := WrapPointElements(input)
	expected := [][]string{{"path", "0"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWrapPointElements_MultiLevelChain(t *testing.T) {
	input := []interface{}{"post", "form_urlencoded", "username"}
	result := WrapPointElements(input)
	expected := [][]string{{"post"}, {"form_urlencoded", "username"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWrapPointElements_Empty(t *testing.T) {
	result := WrapPointElements([]interface{}{})
	if len(result) != 0 {
		t.Errorf("expected nil or empty, got %v", result)
	}
}

func TestWrapPointElements_ComplexChain(t *testing.T) {
	// post > json_doc > hash "password"
	input := []interface{}{"post", "json_doc", "hash", "password"}
	result := WrapPointElements(input)
	expected := [][]string{{"post"}, {"json_doc"}, {"hash", "password"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestExpandPointsToTwoDimensionalArray_Empty(t *testing.T) {
	result, err := ExpandPointsToTwoDimensionalArray([]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestExpandPointsToTwoDimensionalArray_NumericPath(t *testing.T) {
	input := []interface{}{
		[]interface{}{"path", "0"},
	}
	result, err := ExpandPointsToTwoDimensionalArray(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	// "0" should be converted to float64(0)
	if result[0][1] != float64(0) {
		t.Errorf("expected float64(0), got %v (%T)", result[0][1], result[0][1])
	}
}

func TestExpandPointsToTwoDimensionalArray_NonNumeric(t *testing.T) {
	input := []interface{}{
		[]interface{}{"header", "HOST"},
	}
	result, err := ExpandPointsToTwoDimensionalArray(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0][0] != "header" || result[0][1] != "HOST" {
		t.Errorf("expected [header HOST], got %v", result[0])
	}
}

func TestConvertToStringSlice_Basic(t *testing.T) {
	result := ConvertToStringSlice([]interface{}{"a", "b", "c"})
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestConvertToStringSlice_SkipsNil(t *testing.T) {
	result := ConvertToStringSlice([]interface{}{"a", nil, "b"})
	expected := []string{"a", "b"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestConvertToStringSlice_NonString(t *testing.T) {
	result := ConvertToStringSlice([]interface{}{42, true})
	if len(result) != 2 || result[0] != "42" || result[1] != "true" {
		t.Errorf("expected [42 true], got %v", result)
	}
}

func TestConvertToStringSlice_Empty(t *testing.T) {
	result := ConvertToStringSlice([]interface{}{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestActionDetailsToMap_Basic(t *testing.T) {
	input := wallarm.ActionDetails{
		Type:  "equal",
		Value: "test",
		Point: []interface{}{"header", "HOST"},
	}
	result, err := ActionDetailsToMap(input)
	if err != nil {
		t.Fatal(err)
	}
	if result["type"] != "equal" {
		t.Errorf("type: expected equal, got %v", result["type"])
	}
	if result["value"] != "test" {
		t.Errorf("value: expected test, got %v", result["value"])
	}
}

func TestActionDetailsToMap_NilValue(t *testing.T) {
	input := wallarm.ActionDetails{
		Type:  "absent",
		Point: []interface{}{"action_ext"},
	}
	result, err := ActionDetailsToMap(input)
	if err != nil {
		t.Fatal(err)
	}
	// nil value should become ""
	if result["value"] != "" {
		t.Errorf("value: expected empty string for nil, got %v", result["value"])
	}
}
