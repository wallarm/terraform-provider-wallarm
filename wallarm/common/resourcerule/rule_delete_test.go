package resourcerule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"
)

// stubDeleteAPI lets tests observe HintDelete calls and decide what to return.
type stubDeleteAPI struct {
	wallarm.API
	gotFilter *wallarm.HintDeleteFilter
	resp      *wallarm.HintDeleteResp
	err       error
}

func (s *stubDeleteAPI) HintDelete(body *wallarm.HintDelete) (*wallarm.HintDeleteResp, error) {
	if body != nil {
		s.gotFilter = body.Filter
	}
	return s.resp, s.err
}

func newDeleteResource(t *testing.T) *schema.ResourceData {
	t.Helper()
	res := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"client_id": {Type: schema.TypeInt, Optional: true},
			"rule_id":   {Type: schema.TypeInt, Optional: true},
		},
	}
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"client_id": 8649,
		"rule_id":   1234,
	})
	d.SetId("8649/0/1234")
	return d
}

func TestDelete_SuccessPathReturnsNoDiagnostics(t *testing.T) {
	stub := &stubDeleteAPI{
		resp: &wallarm.HintDeleteResp{
			Status: 200,
			Body:   []wallarm.ActionBody{{ID: 1234, ActionID: 5678}},
		},
	}
	del := Delete(func(_ interface{}) wallarm.API { return stub })

	diags := del(context.Background(), newDeleteResource(t), nil)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	assertDeleteFilter(t, stub.gotFilter, 8649, 1234)
}

func TestDelete_EmptyBodyDoesNotError(t *testing.T) {
	stub := &stubDeleteAPI{
		resp: &wallarm.HintDeleteResp{Status: 200, Body: []wallarm.ActionBody{}},
	}
	del := Delete(func(_ interface{}) wallarm.API { return stub })

	diags := del(context.Background(), newDeleteResource(t), nil)
	if diags.HasError() {
		t.Fatalf("empty-body no-op should still succeed; got %v", diags)
	}
	assertDeleteFilter(t, stub.gotFilter, 8649, 1234)
}

func TestDelete_APIErrorReturnsDiagnostic(t *testing.T) {
	stub := &stubDeleteAPI{err: context.Canceled}
	del := Delete(func(_ interface{}) wallarm.API { return stub })

	diags := del(context.Background(), newDeleteResource(t), nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostic on API error, got nil")
	}
	assertDeleteFilter(t, stub.gotFilter, 8649, 1234)
}

func assertDeleteFilter(t *testing.T, f *wallarm.HintDeleteFilter, wantClient, wantRule int) {
	t.Helper()
	if f == nil ||
		len(f.Clientid) != 1 || f.Clientid[0] != wantClient ||
		len(f.ID) != 1 || f.ID[0] != wantRule {
		t.Errorf("unexpected filter: %#v (want client=%d, rule=%d)", f, wantClient, wantRule)
	}
}
