package resourcerule

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	wallarm "github.com/wallarm/wallarm-go"
)

type mockUpdateAPI struct {
	wallarm.API
	gotRuleID int
	gotParams *wallarm.HintUpdateV3Params
	err       error
}

func (m *mockUpdateAPI) HintUpdateV3(ruleID int, params *wallarm.HintUpdateV3Params) (*wallarm.ActionCreateResp, error) {
	m.gotRuleID = ruleID
	m.gotParams = params
	return nil, m.err
}

func testUpdateSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"rule_id":              {Type: schema.TypeInt, Optional: true},
			"variativity_disabled": {Type: schema.TypeBool, Optional: true},
			"comment":              {Type: schema.TypeString, Optional: true},
		},
	}
}

func TestUpdate_Success(t *testing.T) {
	mock := &mockUpdateAPI{}
	cp := func(_ interface{}) wallarm.API { return mock }

	d := testUpdateSchema().TestResourceData()
	d.Set("rule_id", 42)
	d.Set("variativity_disabled", true)
	d.Set("comment", "hello")

	diags := Update(cp)(context.Background(), d, nil)
	if diags.HasError() {
		t.Fatalf("expected no error, got %v", diags)
	}
	if mock.gotRuleID != 42 {
		t.Errorf("expected ruleID=42, got %d", mock.gotRuleID)
	}
	if mock.gotParams == nil || mock.gotParams.Comment == nil || *mock.gotParams.Comment != "hello" {
		t.Errorf("expected comment=hello, got %+v", mock.gotParams)
	}
	if mock.gotParams.VariativityDisabled == nil || !*mock.gotParams.VariativityDisabled {
		t.Errorf("expected variativityDisabled=true, got %+v", mock.gotParams.VariativityDisabled)
	}
}

func TestUpdate_APIError(t *testing.T) {
	mock := &mockUpdateAPI{err: errors.New("boom")}
	cp := func(_ interface{}) wallarm.API { return mock }

	d := testUpdateSchema().TestResourceData()
	d.Set("rule_id", 1)

	diags := Update(cp)(context.Background(), d, nil)
	if !diags.HasError() {
		t.Fatal("expected error diagnostic, got none")
	}
}
