package wallarm

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() { // nolint:gochecknoinits
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"wallarm": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(_ *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("WALLARM_API_HOST"); v == "" {
		t.Fatal(`
		WALLARM_API_HOST must be set for acceptance tests
		Possible values:
		for EU cloud: https://api.wallarm.com
		for US cloud: https://us1.api.wallarm.com
		for RU cloud: https://api.wallarm.ru
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

// ResourceExistsError returns regexp to be used inside TestStep with ExpectError state.
func ResourceExistsError(regex, name string) *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("the resource with the ID "+
			`"%[1]s"`+" already exists - to be managed via Terraform this resource needs "+
			"to be imported into the State. Please see the resource documentation for "+
			`"%[2]s"`+" for more information.", regex, name))
}

// ArgumentMustBePresented returns regexp to be used inside TestStep with ExpectError state.
// When some arguments must be presented while others are specified.
func ArgumentMustBePresented(attribute, templateID string) *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf(`"%[1]s" must be presented with the "%[2]s" template`, attribute, templateID))
}
