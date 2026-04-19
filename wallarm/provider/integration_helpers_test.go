package wallarm

import (
	"fmt"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"
)

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
