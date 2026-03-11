package wallarm

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func TestAccWallarmUser_RequiredFieldsAnalytic(t *testing.T) {
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
				Config: testWallarmUserWithRequiredFields(rnd, email, "analytic", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "analytic"),
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

func TestAccWallarmUser_RequiredFieldsAuditor(t *testing.T) {
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
				Config: testWallarmUserWithRequiredFields(rnd, email, "auditor", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "auditor"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsPartnerAdmin(t *testing.T) {
	if os.Getenv("WALLARM_GLOBAL_ADMIN") != "" {
		t.Skip("Skipping test as it requires 'WALLARM_GLOBAL_ADMIN' not set")
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
				Config: testWallarmUserWithRequiredFields(rnd, email, "partner_admin", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "partner_admin"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
				ExpectError: regexp.MustCompile(`HTTP Status: 403`),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsPartnerAnalytic(t *testing.T) {
	if os.Getenv("WALLARM_GLOBAL_ADMIN") != "" {
		t.Skip("Skipping test as it requires 'WALLARM_GLOBAL_ADMIN' not set")
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
				Config: testWallarmUserWithRequiredFields(rnd, email, "partner_analytic", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "partner_analytic"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
				ExpectError: regexp.MustCompile(`HTTP Status: 403`),
			},
		},
	})
}

func TestAccWallarmUser_RequiredFieldsPartnerAuditor(t *testing.T) {
	if os.Getenv("WALLARM_GLOBAL_ADMIN") != "" {
		t.Skip("Skipping test as it requires 'WALLARM_GLOBAL_ADMIN' not set")
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
				Config: testWallarmUserWithRequiredFields(rnd, email, "partner_auditor", emailName+" "+domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "email", email),
					resource.TestCheckResourceAttr(name, "permissions", "partner_auditor"),
					resource.TestCheckResourceAttr(name, "realname", emailName+" "+domain),
				),
				ExpectError: regexp.MustCompile(`HTTP Status: 403`),
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
