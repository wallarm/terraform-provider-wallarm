package wallarm

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	wallarm "github.com/wallarm/wallarm-go"
)

// testAccWallarmAPISpecPolicyConfig renders an HCL config combining a throwaway
// parent wallarm_api_spec with a wallarm_api_spec_policy attached to it.
// The caller supplies policyBody (the attributes inside the policy resource)
// so each test can vary enabled / mode / threshold fields independently.
func testAccWallarmAPISpecPolicyConfig(rnd, clientID, policyBody string) string {
	return fmt.Sprintf(`
resource "wallarm_api_spec" %[1]q {
  client_id           = %[2]s
  title               = "tf-test-policy-%[1]s"
  file_remote_url     = "https://raw.githubusercontent.com/concentrator/petstore/refs/heads/main/spec.yaml"
  regular_file_update = false
  api_detection       = false
  domains             = ["petstore-tf-test-%[1]s.example.com"]
  instances           = [1]
}

resource "wallarm_api_spec_policy" %[1]q {
  client_id   = %[2]s
  api_spec_id = wallarm_api_spec.%[1]s.api_spec_id
%[3]s
}`, rnd, clientID, policyBody)
}

// testAccCheckWallarmAPISpecPolicyExists — standard presence check on the state.
func testAccCheckWallarmAPISpecPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for %s", resourceName)
		}
		return nil
	}
}

// testAccCheckWallarmAPISpecPolicyDestroy verifies that after destroy either
// the parent spec is gone, or the policy record still exists on the spec but
// is effectively disabled (enabled == false). The policy record itself is a
// persisted field on the spec — the resource cannot truly delete it, only
// flip enabled back to false. If the spec survives with an enabled policy,
// the destroy failed.
func testAccCheckWallarmAPISpecPolicyDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_api_spec_policy" {
			continue
		}
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return fmt.Errorf("invalid client_id for %s: %w", rs.Primary.ID, err)
		}
		apiSpecID, err := strconv.Atoi(rs.Primary.Attributes["api_spec_id"])
		if err != nil {
			return fmt.Errorf("invalid api_spec_id for %s: %w", rs.Primary.ID, err)
		}
		spec, err := api.APISpecReadByID(clientID, apiSpecID)
		if err != nil {
			if errors.Is(err, wallarm.ErrNotFound) {
				continue // parent spec already gone — policy is gone too
			}
			return fmt.Errorf("checking api_spec %d on policy destroy: %w", apiSpecID, err)
		}
		if spec.Policy != nil && spec.Policy.Enabled {
			return fmt.Errorf("wallarm_api_spec_policy %s still enabled on spec %d", rs.Primary.ID, apiSpecID)
		}
	}
	return nil
}

// TestAPISpecPolicyValidation — pure unit test for the StringInSlice validators
// on the violation mode fields. Does not boot the plugintest harness, so it
// runs without Terraform installed. The threshold fields (timeout/timeout_mode/
// max_request_size/max_request_size_mode) are Computed-only (admin-managed),
// so they have no ValidateFunc.
func TestAPISpecPolicyValidation(t *testing.T) {
	t.Parallel()

	violationModeFields := []string{
		"undefined_endpoint_mode",
		"undefined_parameter_mode",
		"missing_parameter_mode",
		"invalid_parameter_value_mode",
		"missing_auth_mode",
		"invalid_request_mode",
	}

	// Violation modes — block | monitor | ignore
	for _, field := range violationModeFields {
		sch := resourceWallarmAPISpecPolicy().Schema[field]
		if sch == nil || sch.ValidateFunc == nil {
			t.Fatalf("violation mode field %q has no ValidateFunc", field)
		}
		for _, good := range []string{"block", "monitor", "ignore"} {
			if _, errs := sch.ValidateFunc(good, field); len(errs) != 0 {
				t.Errorf("violation %s: valid value %q rejected: %v", field, good, errs)
			}
		}
		for _, bad := range []string{"off", ""} {
			_, errs := sch.ValidateFunc(bad, field)
			if len(errs) == 0 {
				t.Errorf("violation %s: expected error for invalid value %q, got none", field, bad)
				continue
			}
			want := fmt.Sprintf(`expected %s to be one of`, field)
			if !strings.Contains(errs[0].Error(), want) {
				t.Errorf("violation %s: want error containing %q, got %q", field, want, errs[0].Error())
			}
		}
	}
}

// TestAccWallarmAPISpecPolicy_Basic creates a policy with all six violation
// modes + both threshold modes set to "monitor", enabled=true, and verifies
// state + ID format {client_id}/{api_spec_id}.
func TestAccWallarmAPISpecPolicy_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec_policy." + rnd
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	body := `
  enabled                      = true
  undefined_endpoint_mode      = "monitor"
  undefined_parameter_mode     = "monitor"
  missing_parameter_mode       = "monitor"
  invalid_parameter_value_mode = "monitor"
  missing_auth_mode            = "monitor"
  invalid_request_mode         = "monitor"
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecPolicyConfig(rnd, clientID, body),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmAPISpecPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "undefined_endpoint_mode", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "undefined_parameter_mode", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "missing_parameter_mode", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "invalid_parameter_value_mode", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "missing_auth_mode", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "invalid_request_mode", "monitor"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(fmt.Sprintf(`^%s/[0-9]+$`, clientID))),
				),
			},
		},
	})
}

// TestAccWallarmAPISpecPolicy_Update flips three violation modes from
// "monitor" to "block" in a second step and verifies the resource ID is
// stable (the change is an in-place update, not a replace).
func TestAccWallarmAPISpecPolicy_Update(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec_policy." + rnd
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	bodyMonitor := `
  enabled                      = true
  undefined_endpoint_mode      = "monitor"
  undefined_parameter_mode     = "monitor"
  missing_parameter_mode       = "monitor"
  invalid_parameter_value_mode = "monitor"
  missing_auth_mode            = "monitor"
  invalid_request_mode         = "monitor"
`

	bodyMixed := `
  enabled                      = true
  undefined_endpoint_mode      = "block"
  undefined_parameter_mode     = "block"
  missing_parameter_mode       = "block"
  invalid_parameter_value_mode = "monitor"
  missing_auth_mode            = "monitor"
  invalid_request_mode         = "monitor"
`

	var firstID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecPolicyConfig(rnd, clientID, bodyMonitor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmAPISpecPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "undefined_endpoint_mode", "monitor"),
					func(s *terraform.State) error {
						firstID = s.RootModule().Resources[resourceName].Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccWallarmAPISpecPolicyConfig(rnd, clientID, bodyMixed),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmAPISpecPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "undefined_endpoint_mode", "block"),
					resource.TestCheckResourceAttr(resourceName, "undefined_parameter_mode", "block"),
					resource.TestCheckResourceAttr(resourceName, "missing_parameter_mode", "block"),
					resource.TestCheckResourceAttr(resourceName, "invalid_parameter_value_mode", "monitor"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[resourceName].Primary.ID
						if newID != firstID {
							return fmt.Errorf("expected resource ID to remain %q across Update, got %q (replace instead of in-place)", firstID, newID)
						}
						return nil
					},
				),
			},
		},
	})
}

// TestAccWallarmAPISpecPolicy_Disable flips enabled=true → false in a second
// step and verifies that the non-default mode set in step 1 persists on the
// spec (soft-delete semantics: the policy record stays, enabled goes false).
func TestAccWallarmAPISpecPolicy_Disable(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec_policy." + rnd
	specResourceName := "wallarm_api_spec." + rnd
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	bodyEnabled := `
  enabled                 = true
  undefined_endpoint_mode = "block"
`
	bodyDisabled := `
  enabled                 = false
  undefined_endpoint_mode = "block"
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecPolicyConfig(rnd, clientID, bodyEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmAPISpecPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "undefined_endpoint_mode", "block"),
				),
			},
			{
				Config: testAccWallarmAPISpecPolicyConfig(rnd, clientID, bodyDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmAPISpecPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "undefined_endpoint_mode", "block"),
					testAccCheckAPISpecPolicyPersistedMode(specResourceName, "undefined_endpoint_mode", "block"),
				),
			},
		},
	})
}

// testAccCheckAPISpecPolicyPersistedMode reads the parent spec directly via
// the API and asserts that a specific policy mode field has the expected
// value — used to verify that Disable keeps configured settings on the spec
// (soft-delete semantics).
func testAccCheckAPISpecPolicyPersistedMode(specResource, field, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[specResource]
		if !ok {
			return fmt.Errorf("parent spec resource not found in state: %s", specResource)
		}
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return fmt.Errorf("invalid client_id: %w", err)
		}
		apiSpecID, err := strconv.Atoi(rs.Primary.Attributes["api_spec_id"])
		if err != nil {
			return fmt.Errorf("invalid api_spec_id: %w", err)
		}
		api, err := testAccNewAPIClient()
		if err != nil {
			return err
		}
		spec, err := api.APISpecReadByID(clientID, apiSpecID)
		if err != nil {
			return fmt.Errorf("reading spec %d: %w", apiSpecID, err)
		}
		if spec.Policy == nil {
			return fmt.Errorf("spec %d has no policy record", apiSpecID)
		}
		var got string
		switch field {
		case "undefined_endpoint_mode":
			got = spec.Policy.UndefinedEndpointMode
		case "undefined_parameter_mode":
			got = spec.Policy.UndefinedParameterMode
		case "missing_parameter_mode":
			got = spec.Policy.MissingParameterMode
		case "invalid_parameter_value_mode":
			got = spec.Policy.InvalidParameterValueMode
		case "missing_auth_mode":
			got = spec.Policy.MissingAuthMode
		case "invalid_request_mode":
			got = spec.Policy.InvalidRequestMode
		case "timeout_mode":
			got = spec.Policy.TimeoutMode
		case "max_request_size_mode":
			got = spec.Policy.MaxRequestSizeMode
		default:
			return fmt.Errorf("unsupported field %q", field)
		}
		if got != want {
			return fmt.Errorf("policy %s: want %q, got %q (expected soft-delete to preserve configured value)", field, want, got)
		}
		return nil
	}
}

// TestAccWallarmAPISpecPolicy_Import round-trips the resource via terraform
// import and verifies every field matches. Policy has no signed-URL analog
// or other sensitive/volatile fields, so ImportStateVerifyIgnore is empty.
func TestAccWallarmAPISpecPolicy_Import(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec_policy." + rnd
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	body := `
  enabled                      = true
  undefined_endpoint_mode      = "block"
  undefined_parameter_mode     = "monitor"
  missing_parameter_mode       = "monitor"
  invalid_parameter_value_mode = "ignore"
  missing_auth_mode            = "monitor"
  invalid_request_mode         = "monitor"
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecPolicyConfig(rnd, clientID, body),
				Check:  testAccCheckWallarmAPISpecPolicyExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
