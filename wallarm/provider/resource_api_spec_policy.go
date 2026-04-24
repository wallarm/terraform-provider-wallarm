package wallarm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	wallarm "github.com/wallarm/wallarm-go"
)

var violationModes = []string{"block", "monitor", "ignore"}

func resourceWallarmAPISpecPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmAPISpecPolicyPut,
		ReadContext:   resourceWallarmAPISpecPolicyRead,
		UpdateContext: resourceWallarmAPISpecPolicyPut,
		DeleteContext: resourceWallarmAPISpecPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmAPISpecPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The Wallarm client ID owning the parent API spec.",
			},
			"api_spec_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the API spec this policy applies to.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the policy is actively enforced. Setting to false (or destroying the resource) soft-disables enforcement while preserving all other settings on the spec.",
			},
			"condition": resourcerule.ScopeActionSchemaMutable(),

			"undefined_endpoint_mode":      violationModeSchema("Action when a request hits an endpoint not defined in the spec."),
			"undefined_parameter_mode":     violationModeSchema("Action when a request carries a parameter not defined in the spec."),
			"missing_parameter_mode":       violationModeSchema("Action when a required parameter is missing."),
			"invalid_parameter_value_mode": violationModeSchema("Action when a parameter value does not match its declared type/format."),
			"missing_auth_mode":            violationModeSchema("Action when the request lacks the authentication declared by the spec."),
			"invalid_request_mode":         violationModeSchema("Action when the request body does not match the declared schema."),

			"timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Max spec-processing time per request in milliseconds. Managed by Wallarm; requires elevated permissions to modify (not settable through this resource).",
			},
			"timeout_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Reaction when the timeout is exceeded. Managed by Wallarm; requires elevated permissions to modify.",
			},
			"max_request_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Max inspected request size in KB. Managed by Wallarm; requires elevated permissions to modify.",
			},
			"max_request_size_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Reaction when the request-size limit is exceeded. Managed by Wallarm; requires elevated permissions to modify.",
			},
		},
	}
}

func violationModeSchema(desc string) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "monitor",
		ValidateFunc: validation.StringInSlice(violationModes, false),
		Description:  desc + " One of: block, monitor, ignore.",
	}
}

func resourceWallarmAPISpecPolicyPut(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	conditions, err := expandPolicyConditions(d.Get("condition").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	body := &wallarm.APISpecPolicy{
		Enabled:                   d.Get("enabled").(bool),
		Conditions:                conditions,
		UndefinedEndpointMode:     d.Get("undefined_endpoint_mode").(string),
		UndefinedParameterMode:    d.Get("undefined_parameter_mode").(string),
		MissingParameterMode:      d.Get("missing_parameter_mode").(string),
		InvalidParameterValueMode: d.Get("invalid_parameter_value_mode").(string),
		MissingAuthMode:           d.Get("missing_auth_mode").(string),
		InvalidRequestMode:        d.Get("invalid_request_mode").(string),
		// Timeout/TimeoutMode/MaxRequestSize/MaxRequestSizeMode deliberately omitted —
		// those fields are admin-only on the Wallarm API side and ignored for regular
		// users. Exposed as Computed-only in the schema so state still reflects API values.
	}

	resp, err := client.APISpecPolicyPut(clientID, apiSpecID, body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d/%d/policy", clientID, apiSpecID))

	// Populate state from the PUT response — it echoes all 11 fields, so the
	// extra APISpecReadByID roundtrip is unnecessary.
	if resp.Body != nil {
		return setPolicyToState(d, resp.Body)
	}
	return nil
}

func resourceWallarmAPISpecPolicyRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	spec, err := client.APISpecReadByID(clientID, apiSpecID)
	if err != nil {
		if errors.Is(err, wallarm.ErrNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if spec.Policy == nil {
		d.SetId("")
		return nil
	}

	return setPolicyToState(d, spec.Policy)
}

// setPolicyToState writes every policy field from the API response into state.
// Shared between Read and Put so the PUT response can populate state without
// an extra APISpecReadByID call.
func setPolicyToState(d *schema.ResourceData, p *wallarm.APISpecPolicy) diag.Diagnostics {
	d.Set("enabled", p.Enabled)
	d.Set("undefined_endpoint_mode", p.UndefinedEndpointMode)
	d.Set("undefined_parameter_mode", p.UndefinedParameterMode)
	d.Set("missing_parameter_mode", p.MissingParameterMode)
	d.Set("invalid_parameter_value_mode", p.InvalidParameterValueMode)
	d.Set("missing_auth_mode", p.MissingAuthMode)
	d.Set("invalid_request_mode", p.InvalidRequestMode)
	d.Set("timeout_mode", p.TimeoutMode)
	d.Set("max_request_size_mode", p.MaxRequestSizeMode)
	d.Set("timeout", p.Timeout)
	d.Set("max_request_size", p.MaxRequestSize)
	if err := d.Set("condition", flattenPolicyConditions(p.Conditions)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting condition: %w", err))
	}
	return nil
}

func resourceWallarmAPISpecPolicyDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	// Soft-delete: read current policy, then PUT back with Enabled=false so all
	// other settings are preserved. Lets the user re-enable later without
	// reconfiguring.
	spec, err := client.APISpecReadByID(clientID, apiSpecID)
	if err != nil {
		// Parent spec gone → policy gone with it.
		if errors.Is(err, wallarm.ErrNotFound) {
			return nil
		}
		return diag.FromErr(err)
	}
	if spec.Policy == nil {
		// No policy on this spec → nothing to disable.
		return nil
	}

	body := *spec.Policy
	body.Enabled = false
	// Zero out the 4 admin-only threshold fields so they drop out of the JSON
	// via omitempty — regular users cannot PUT them back, even to the same
	// values the API just returned, without tripping the admin-only check.
	body.Timeout = 0
	body.TimeoutMode = ""
	body.MaxRequestSize = 0
	body.MaxRequestSizeMode = ""
	if _, err := client.APISpecPolicyPut(clientID, apiSpecID, &body); err != nil {
		if errors.Is(err, wallarm.ErrNotFound) {
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}

func resourceWallarmAPISpecPolicyImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 || parts[2] != "policy" {
		return nil, fmt.Errorf("invalid id %q, expected {client_id}/{api_spec_id}/policy", d.Id())
	}
	clientID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid client_id: %w", err)
	}
	apiSpecID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid api_spec_id: %w", err)
	}
	d.Set("client_id", clientID)
	d.Set("api_spec_id", apiSpecID)
	d.SetId(fmt.Sprintf("%d/%d/policy", clientID, apiSpecID))
	return []*schema.ResourceData{d}, nil
}

// expandPolicyConditions converts the HCL condition TypeSet into the
// []APISpecPolicyCondition wire shape. It reuses the rule scope expansion
// helper (same TypeSet hash + schema) and then copies fields across.
func expandPolicyConditions(set *schema.Set) ([]wallarm.APISpecPolicyCondition, error) {
	raw, err := resourcerule.ExpandSetToActionDetailsList(set)
	if err != nil {
		return nil, err
	}
	out := make([]wallarm.APISpecPolicyCondition, 0, len(raw))
	for _, r := range raw {
		out = append(out, wallarm.APISpecPolicyCondition{
			Type:  r.Type,
			Value: r.Value,
			Point: r.Point,
		})
	}
	return out, nil
}

// flattenPolicyConditions converts []APISpecPolicyCondition back into the
// schema-compatible form used by the condition TypeSet. It constructs an
// intermediate wallarm.ActionDetails so the existing ActionDetailsToMap +
// TransformAPIActionToSchema helpers (which drive rule reads) can be reused —
// this guarantees the same hash used by ScopeActionSchema's Set function.
func flattenPolicyConditions(conds []wallarm.APISpecPolicyCondition) *schema.Set {
	set := &schema.Set{F: resourcerule.HashActionDetails}
	for _, c := range conds {
		ad := wallarm.ActionDetails{Type: c.Type, Value: c.Value, Point: c.Point}
		m, err := resourcerule.ActionDetailsToMap(ad)
		if err != nil {
			log.Printf("[WARN] wallarm_api_spec_policy: dropping unrepresentable condition (type=%q point=%v value=%v): %s",
				c.Type, c.Point, c.Value, err)
			continue
		}
		resourcerule.TransformAPIActionToSchema(m)
		set.Add(m)
	}
	return set
}
