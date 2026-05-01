// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wallarm

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	wallarm "github.com/wallarm/wallarm-go"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccProtoV5ProviderFactories map[string]func() (tfprotov5.ProviderServer, error)

func TestMain(m *testing.M) {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"wallarm": testAccProvider,
	}
	testAccProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
		"wallarm": func() (tfprotov5.ProviderServer, error) {
			return schema.NewGRPCProviderServer(Provider()), nil
		},
	}
	os.Exit(m.Run())
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(_ *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("WALLARM_API_HOST"); v == "" {
		t.Fatal(`
		WALLARM_API_HOST must be set for acceptance tests
		Possible values:
		for EU cloud: https://api.wallarm.com
		for US cloud: https://us1.api.wallarm.com
		`)
	}

	if v := os.Getenv("WALLARM_API_TOKEN"); v == "" {
		t.Fatal("WALLARM_API_TOKEN must be set for acceptance tests. The TOKEN is used to authenticate in the Cloud")
	}
}

func generateRandomResourceName(n int) string {
	return acctest.RandStringFromCharSet(n, acctest.CharSetAlpha)
}

func generateRandomNumber(n int) string {
	return strconv.Itoa(acctest.RandIntRange(1000, n))
}

func generateRandomUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// testAccNewAPIClient builds a wallarm.API from WALLARM_API_TOKEN and
// WALLARM_API_HOST. Use it in CheckDestroy when the test uses
// ProtoV5ProviderFactories (each factory call returns a fresh provider whose
// meta is not reachable via testAccProvider.Meta()). The client is independent
// of the provider under test, so parallel tests do not race on its Configure.
// Returns an uncached client — CheckDestroy wants ground truth from the API,
// not a potentially-stale entry from a different provider's cache.
func testAccNewAPIClient() (wallarm.API, error) {
	host := os.Getenv("WALLARM_API_HOST")
	token := os.Getenv("WALLARM_API_TOKEN")
	if host == "" || token == "" {
		return nil, fmt.Errorf("WALLARM_API_HOST and WALLARM_API_TOKEN must be set")
	}
	headers := make(http.Header)
	headers.Add("X-WallarmAPI-Token", token)
	api, err := wallarm.New(
		wallarm.UsingBaseURL(host),
		wallarm.Headers(headers),
	)
	if err != nil {
		return nil, fmt.Errorf("creating Wallarm client: %w", err)
	}
	return api, nil
}

// testAccCheckHintDestroyed verifies that every Terraform resource of the given
// type has been deleted server-side. Looks up each resource's hint by exact
// rule_id; fails if any still exists. Use as the body of a per-resource
// CheckDestroy:
//
//	func testAccCheckWallarmRuleXxxDestroy(s *terraform.State) error {
//	    return testAccCheckHintDestroyed(s, "wallarm_rule_xxx")
//	}
//
// Built independently of the provider under test (`testAccNewAPIClient`) so it
// is race-safe under proto-v5 factories where each test gets a fresh provider
// instance and `testAccProvider.Meta()` returns nil.
func testAccCheckHintDestroyed(s *terraform.State, resourceType string) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}
		ruleID, err := strconv.Atoi(rs.Primary.Attributes["rule_id"])
		if err != nil {
			return fmt.Errorf("invalid rule_id for %s: %w", rs.Primary.ID, err)
		}
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return fmt.Errorf("invalid client_id for %s: %w", rs.Primary.ID, err)
		}
		// OrderBy is required by the API — HintRead returns 400 without it.
		resp, err := api.HintRead(&wallarm.HintRead{
			Limit:   1,
			OrderBy: "updated_at",
			Filter:  &wallarm.HintFilter{Clientid: []int{clientID}, ID: []int{ruleID}},
		})
		if err != nil {
			return fmt.Errorf("checking hint %d still exists: %w", ruleID, err)
		}
		if resp.Body != nil && len(*resp.Body) > 0 {
			return fmt.Errorf("%s %s still exists", resourceType, rs.Primary.ID)
		}
	}
	return nil
}

// ResourceExistsError returns regexp to be used inside TestStep with ExpectError state.
func ResourceExistsError(regex, name string) *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("the resource with the ID "+
			`"%[1]s"`+" already exists - to be managed via Terraform this resource needs "+
			"to be imported into the State. Please see the resource documentation for "+
			`"%[2]s"`+" for more information", regex, name))
}

// ArgumentMustBePresented returns regexp to be used inside TestStep with ExpectError state.
// When some arguments must be presented while others are specified.
func ArgumentMustBePresented(attribute, templateID string) *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf(`"%[1]s" must be presented with the "%[2]s" template`, attribute, templateID))
}
