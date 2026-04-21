package wallarm

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testIntegrationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"client_id":      {Type: schema.TypeInt, Optional: true},
			"integration_id": {Type: schema.TypeInt, Optional: true},
		},
	}
}

func TestImportIntegration_Valid(t *testing.T) {
	d := testIntegrationSchema().TestResourceData()
	d.SetId("1111/slack/2222")

	result, err := importIntegration("slack")(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Id() != "1111/slack/2222" {
		t.Errorf("id: got %q", result[0].Id())
	}
	if result[0].Get("client_id").(int) != 1111 {
		t.Errorf("client_id: got %d", result[0].Get("client_id").(int))
	}
	if result[0].Get("integration_id").(int) != 2222 {
		t.Errorf("integration_id: got %d", result[0].Get("integration_id").(int))
	}
}

func TestImportIntegration_WrongType(t *testing.T) {
	d := testIntegrationSchema().TestResourceData()
	d.SetId("1111/email/2222")
	_, err := importIntegration("slack")(context.Background(), d, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid type segment") {
		t.Errorf("expected type-mismatch error, got %v", err)
	}
}

func TestImportIntegration_WrongPartCount(t *testing.T) {
	for _, id := range []string{"", "1111", "1111/slack", "1111/slack/2222/extra"} {
		d := testIntegrationSchema().TestResourceData()
		d.SetId(id)
		_, err := importIntegration("slack")(context.Background(), d, nil)
		if err == nil || !strings.Contains(err.Error(), "invalid id") {
			t.Errorf("id=%q: expected invalid-id error, got %v", id, err)
		}
	}
}

func TestImportIntegration_InvalidClientID(t *testing.T) {
	d := testIntegrationSchema().TestResourceData()
	d.SetId("abc/slack/2222")
	_, err := importIntegration("slack")(context.Background(), d, nil)
	if err == nil || !strings.Contains(err.Error(), "client_id") {
		t.Errorf("expected client_id error, got %v", err)
	}
}

func TestImportIntegration_InvalidIntegrationID(t *testing.T) {
	d := testIntegrationSchema().TestResourceData()
	d.SetId("1111/slack/xyz")
	_, err := importIntegration("slack")(context.Background(), d, nil)
	if err == nil || !strings.Contains(err.Error(), "integration_id") {
		t.Errorf("expected integration_id error, got %v", err)
	}
}
