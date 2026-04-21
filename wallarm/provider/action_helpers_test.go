package wallarm

import (
	"testing"
)

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
