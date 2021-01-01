package wallarm

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmUser_RequiredFieldsDeploy(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_user." + rnd
	emailName := generateRandomResourceName(8)
	domain := generateRandomResourceName(6)
	topDomain := generateRandomResourceName(3)
	email := emailName + "@" + domain + "." + topDomain

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmUserWithRequiredFields(rnd, email, "deploy", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "deploy"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsAnalyst(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_user." + rnd
	emailName := generateRandomResourceName(8)
	domain := generateRandomResourceName(6)
	topDomain := generateRandomResourceName(3)
	email := emailName + "@" + domain + "." + topDomain

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmUserWithRequiredFields(rnd, email, "analyst", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "analyst"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsAdmin(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_user." + rnd
	emailName := generateRandomResourceName(8)
	domain := generateRandomResourceName(6)
	topDomain := generateRandomResourceName(3)
	email := emailName + "@" + domain + "." + topDomain

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmUserWithRequiredFields(rnd, email, "admin", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "admin"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsReadOnly(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_user." + rnd
	emailName := generateRandomResourceName(8)
	domain := generateRandomResourceName(6)
	topDomain := generateRandomResourceName(3)
	email := emailName + "@" + domain + "." + topDomain

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmUserWithRequiredFields(rnd, email, "read_only", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "read_only"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsGlobalAdmin(t *testing.T) {
	if os.Getenv("WALLARM_GLOBAL_ADMIN") == "" {
		t.Skip("Skipping not finished test")
	}

	rnd := generateRandomResourceName(5)
	name := "wallarm_user." + rnd
	emailName := generateRandomResourceName(8)
	domain := generateRandomResourceName(6)
	topDomain := generateRandomResourceName(3)
	email := emailName + "@" + domain + "." + topDomain

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmUserWithRequiredFields(rnd, email, "global_admin", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "global_admin"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
				ExpectError: regexp.MustCompile(`HTTP Status: 403, Body: {"status":403,"body":{"permissions":{"error":"invalid","reason":"check access failed"}}}`),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsGlobalAnalyst(t *testing.T) {
	if os.Getenv("WALLARM_GLOBAL_ADMIN") == "" {
		t.Skip("Skipping not finished test")
	}

	rnd := generateRandomResourceName(5)
	name := "wallarm_user." + rnd
	emailName := generateRandomResourceName(8)
	domain := generateRandomResourceName(6)
	topDomain := generateRandomResourceName(3)
	email := emailName + "@" + domain + "." + topDomain

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmUserWithRequiredFields(rnd, email, "global_analyst", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "global_analyst"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
				ExpectError: regexp.MustCompile(`HTTP Status: 403, Body: {"status":403,"body":{"permissions":{"error":"invalid","reason":"check access failed"}}}`),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsGlobalReadOnly(t *testing.T) {
	if os.Getenv("WALLARM_GLOBAL_ADMIN") == "" {
		t.Skip("Skipping not finished test")
	}

	rnd := generateRandomResourceName(5)
	name := "wallarm_user." + rnd
	emailName := generateRandomResourceName(8)
	domain := generateRandomResourceName(6)
	topDomain := generateRandomResourceName(3)
	email := emailName + "@" + domain + "." + topDomain

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmUserWithRequiredFields(rnd, email, "global_read_only", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "global_read_only"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
				ExpectError: regexp.MustCompile(`HTTP Status: 403, Body: {"status":403,"body":{"permissions":{"error":"invalid","reason":"check access failed"}}}`),
			},
		},
	})
}

func testWallarmUserWithRequiredFields(resourceID, email, permissions, realname string) string {
	return fmt.Sprintf(`
resource "wallarm_user" "%[1]s" {
	email = "%[2]s"
	permissions = "%[3]s"
	realname = "%[4]s"
}`, resourceID, email, permissions, realname)
}
