package common

import (
	"testing"
)

func TestConvertToStringSlice_Basic(t *testing.T) {
	input := []interface{}{"a", "b", "c"}
	got := ConvertToStringSlice(input)
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("expected [a b c], got %v", got)
	}
}

func TestConvertToStringSlice_SkipsNil(t *testing.T) {
	input := []interface{}{"a", nil, "c"}
	got := ConvertToStringSlice(input)
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("expected [a c], got %v", got)
	}
}

func TestConvertToStringSlice_NonStringTypes(t *testing.T) {
	input := []interface{}{42, true, 3.14}
	got := ConvertToStringSlice(input)
	if len(got) != 3 || got[0] != "42" || got[1] != "true" || got[2] != "3.14" {
		t.Errorf("expected [42 true 3.14], got %v", got)
	}
}

func TestConvertToStringSlice_Empty(t *testing.T) {
	got := ConvertToStringSlice([]interface{}{})
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}
