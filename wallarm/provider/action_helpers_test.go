package wallarm

import (
	"fmt"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

func TestEqualWithoutOrder_SameOrder(t *testing.T) {
	a := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []interface{}{"path", 0.0}, Value: nil},
	}
	b := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []interface{}{"path", 0.0}, Value: nil},
	}
	if !equalWithoutOrder(a, b) {
		t.Error("expected true for identical conditions")
	}
}

func TestEqualWithoutOrder_DifferentOrder(t *testing.T) {
	a := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"path", 0.0}, Value: nil},
		{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
	}
	b := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "example.com"},
		{Type: "equal", Point: []interface{}{"path", 0.0}, Value: nil},
	}
	if !equalWithoutOrder(a, b) {
		t.Error("expected true for same conditions in different order")
	}
}

func TestEqualWithoutOrder_Different(t *testing.T) {
	a := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "a.com"},
	}
	b := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "b.com"},
	}
	if equalWithoutOrder(a, b) {
		t.Error("expected false for different values")
	}
}

func TestEqualWithoutOrder_Empty(t *testing.T) {
	if !equalWithoutOrder([]wallarm.ActionDetails{}, []wallarm.ActionDetails{}) {
		t.Error("expected true for empty slices")
	}
}

func TestEqualWithoutOrder_DifferentLengths(t *testing.T) {
	a := []wallarm.ActionDetails{
		{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "a.com"},
	}
	if equalWithoutOrder(a, []wallarm.ActionDetails{}) {
		t.Error("expected false for different lengths")
	}
}

func TestCompareActionDetails_Match(t *testing.T) {
	a := wallarm.ActionDetails{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "x"}
	b := wallarm.ActionDetails{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "x"}
	if !compareActionDetails(a, b) {
		t.Error("expected true for matching conditions")
	}
}

func TestCompareActionDetails_DifferentType(t *testing.T) {
	a := wallarm.ActionDetails{Type: "equal", Point: []interface{}{"header", "HOST"}, Value: "x"}
	b := wallarm.ActionDetails{Type: "iequal", Point: []interface{}{"header", "HOST"}, Value: "x"}
	if compareActionDetails(a, b) {
		t.Error("expected false for different types")
	}
}

func TestActionPointsEqual_Same(t *testing.T) {
	a := []interface{}{"header", "HOST"}
	b := []interface{}{"header", "HOST"}
	if !actionPointsEqual(a, b) {
		t.Error("expected true")
	}
}

func TestActionPointsEqual_DifferentLength(t *testing.T) {
	a := []interface{}{"header", "HOST"}
	b := []interface{}{"header"}
	if actionPointsEqual(a, b) {
		t.Error("expected false for different lengths")
	}
}

func TestImportAsExistsError(t *testing.T) {
	err := ImportAsExistsError("wallarm_rule_mode", "1/2/3")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestIsNotFoundError_404(t *testing.T) {
	err := &wallarm.APIError{StatusCode: 404, Body: "not found"}
	if !isNotFoundError(err) {
		t.Error("expected true for 404")
	}
}

func TestIsNotFoundError_500(t *testing.T) {
	err := &wallarm.APIError{StatusCode: 500, Body: "server error"}
	if isNotFoundError(err) {
		t.Error("expected false for 500")
	}
}

func TestIsNotFoundError_NonAPI(t *testing.T) {
	err := fmt.Errorf("some other error")
	if isNotFoundError(err) {
		t.Error("expected false for non-API error")
	}
}
