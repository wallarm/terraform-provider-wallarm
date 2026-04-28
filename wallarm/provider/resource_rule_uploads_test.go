package wallarm

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleUploadsCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_uploads." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleUploadsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleUploadsBasicConfig(rnd, "docs", "iequal", "uploads.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "file_type", "docs"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "form_urlencoded"),
					resource.TestCheckResourceAttr(name, "point.1.1", "query"),
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleUploadsCreate_IncorrectFileType(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmRuleUploadsBasicConfig(rnd, "incorrect", "iequal", "uploads.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				ExpectError: regexp.MustCompile(`expected file_type to be one of \["docs" "html" "images" "music" "video"\], got incorrect`),
			},
		},
	})
}

func TestAccRuleUploadsCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_uploads." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleUploadsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleUploadsCreateRecreate(rnd, "html"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "file_type", "html"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
			{
				Config: testAccRuleUploadsCreateRecreate(rnd, "html"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "file_type", "html"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
		},
	})
}

func TestAccRuleUploadsCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_uploads." + rnd
	point := `["header","HOST"],["pollution"]`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleUploadsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleUploadsDefaultBranchConfig(rnd, "music", point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "file_type", "music"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckResourceAttr(name, "action.#", "0"),
				),
			},
		},
	})
}

func testWallarmRuleUploadsBasicConfig(resourceID, fileType, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_uploads" "%[1]s" {
  action {
    type = "%[2]s"
    value = "%[3]s"
    point = {
      header = "%[4]s"
    }
  }
  point = [%[5]s]
  file_type = "%[6]s"
}`, resourceID, actionType, actionValue, actionPoint, point, fileType)
}

func testWallarmRuleUploadsDefaultBranchConfig(resourceID, fileType, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_uploads" "%[1]s" {
  point = [%[2]s]
  file_type = "%[3]s"
}`, resourceID, point, fileType)
}

func testAccRuleUploadsCreateRecreate(resourceID, fileType string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_uploads" "%[1]s" {
  point = [["header", "X-FOOBAR"]]
  file_type = "%[2]s"
}`, resourceID, fileType)
}

func TestAccRuleUploadsUpdateInPlaceComment(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_uploads." + rnd
	var firstRuleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleUploadsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleUploadsUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleUploadsUpdateCommentConfig(rnd, "second comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "second comment"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						if newID != firstRuleID {
							return fmt.Errorf("expected rule_id to stay stable on in-place update, was %s now %s", firstRuleID, newID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccRuleUploadsUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_uploads" %[1]q {
  comment   = %[2]q
  file_type = "docs"
  action {
    type  = "iequal"
    value = "uploads_comment_update.example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"], ["form_urlencoded", "query"]]
}`, resourceID, comment)
}

func testAccCheckWallarmRuleUploadsDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_uploads" {
			continue
		}

		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return err
		}
		actionID, err := strconv.Atoi(rs.Primary.Attributes["action_id"])
		if err != nil {
			return err
		}

		hint := &wallarm.HintRead{
			Limit:     APIListLimit,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ActionID: []int{actionID},
				Type:     []string{"uploads"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Allow Certain File Type rule still exists")
		}
	}

	return nil
}
