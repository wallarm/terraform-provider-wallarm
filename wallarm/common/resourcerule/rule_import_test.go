package resourcerule

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testImportSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action_id": {Type: schema.TypeInt, Optional: true},
			"rule_id":   {Type: schema.TypeInt, Optional: true},
			"rule_type": {Type: schema.TypeString, Optional: true},
		},
	}
}

func TestImport_Valid(t *testing.T) {
	d := testImportSchema().TestResourceData()
	d.SetId("42/100/200")

	result, err := Import("bola")(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Id() != "42/100/200" {
		t.Errorf("expected id 42/100/200, got %q", result[0].Id())
	}
	if result[0].Get("action_id").(int) != 100 {
		t.Errorf("expected action_id 100, got %d", result[0].Get("action_id").(int))
	}
	if result[0].Get("rule_id").(int) != 200 {
		t.Errorf("expected rule_id 200, got %d", result[0].Get("rule_id").(int))
	}
	if result[0].Get("rule_type").(string) != "bola" {
		t.Errorf("expected rule_type bola, got %q", result[0].Get("rule_type").(string))
	}
}

func TestImport_WrongPartCount(t *testing.T) {
	cases := []string{"", "42", "42/100", "42/100/200/extra", "42/100/200/extra/more"}
	for _, id := range cases {
		d := testImportSchema().TestResourceData()
		d.SetId(id)
		_, err := Import("bola")(context.Background(), d, nil)
		if err == nil {
			t.Errorf("expected error for id %q, got nil", id)
			continue
		}
		if !strings.Contains(err.Error(), "invalid id") {
			t.Errorf("expected 'invalid id' in error for %q, got %q", id, err.Error())
		}
	}
}

func TestImport_EmptySegments(t *testing.T) {
	cases := []string{"/100/200", "42//200", "42/100/"}
	for _, id := range cases {
		d := testImportSchema().TestResourceData()
		d.SetId(id)
		_, err := Import("bola")(context.Background(), d, nil)
		if err == nil {
			t.Errorf("expected error for id %q, got nil", id)
		}
	}
}

func TestImport_InvalidClientID(t *testing.T) {
	d := testImportSchema().TestResourceData()
	d.SetId("abc/100/200")
	_, err := Import("bola")(context.Background(), d, nil)
	if err == nil || !strings.Contains(err.Error(), "client_id") {
		t.Errorf("expected client_id error, got %v", err)
	}
}

func TestImport_InvalidActionID(t *testing.T) {
	d := testImportSchema().TestResourceData()
	d.SetId("42/xyz/200")
	_, err := Import("bola")(context.Background(), d, nil)
	if err == nil || !strings.Contains(err.Error(), "action_id") {
		t.Errorf("expected action_id error, got %v", err)
	}
}

func TestImport_InvalidRuleID(t *testing.T) {
	d := testImportSchema().TestResourceData()
	d.SetId("42/100/zzz")
	_, err := Import("bola")(context.Background(), d, nil)
	if err == nil || !strings.Contains(err.Error(), "rule_id") {
		t.Errorf("expected rule_id error, got %v", err)
	}
}
