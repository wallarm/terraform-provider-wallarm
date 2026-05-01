package wallarm

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleDisableStampCreate_Basic(t *testing.T) {
	if os.Getenv("WALLARM_EXTRA_PERMISSIONS") == "" {
		t.Skip("Skipping not test as it requires WALLARM_EXTRA_PERMISSIONS set")
	}
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_stamp." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableStampDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleDisableStampBasicConfig(rnd, 1234, "iequal", "stamp.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "1234"),
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

func TestAccRuleDisableStampCreateRecreate(t *testing.T) {
	if os.Getenv("WALLARM_EXTRA_PERMISSIONS") == "" {
		t.Skip("Skipping not test as it requires WALLARM_EXTRA_PERMISSIONS set")
	}
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_stamp." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableStampDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDisableStampCreateRecreate(rnd, 5678),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "5678"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
			{
				Config: testAccRuleDisableStampCreateRecreate(rnd, 5678),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "5678"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
		},
	})
}

func TestAccRuleDisableStampCreate_DefaultBranch(t *testing.T) {
	if os.Getenv("WALLARM_EXTRA_PERMISSIONS") == "" {
		t.Skip("Skipping not test as it requires WALLARM_EXTRA_PERMISSIONS set")
	}
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_stamp." + rnd
	point := `["header","HOST"],["pollution"]`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableStampDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleDisableStampDefaultBranchConfig(rnd, 9012, point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "9012"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckResourceAttr(name, "action.#", "0"),
				),
			},
		},
	})
}

func testWallarmRuleDisableStampBasicConfig(resourceID string, stamp int, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_stamp" %[1]q {
  action {
    type = %[2]q
    value = %[3]q
    point = {
      header = %[4]q
    }
  }
  point = [%[5]s]
  stamp = %[6]d
}`, resourceID, actionType, actionValue, actionPoint, point, stamp)
}

func testWallarmRuleDisableStampDefaultBranchConfig(resourceID string, stamp int, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_stamp" %[1]q {
  point = [%[2]s]
  stamp = %[3]d
}`, resourceID, point, stamp)
}

func testAccRuleDisableStampCreateRecreate(resourceID string, stamp int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_stamp" %[1]q {
  point = [["header", "X-FOOBAR"]]
  stamp = %[2]d
}`, resourceID, stamp)
}

func TestAccRuleDisableStampUpdateInPlaceComment(t *testing.T) {
	if os.Getenv("WALLARM_EXTRA_PERMISSIONS") == "" {
		t.Skip("Skipping not test as it requires WALLARM_EXTRA_PERMISSIONS set")
	}
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_stamp." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableStampDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDisableStampUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleDisableStampUpdateCommentConfig(rnd, "second comment"),
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

func testAccRuleDisableStampUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_stamp" %[1]q {
  comment = %[2]q
  stamp   = 1234
  action {
    type  = "iequal"
    value = "disable_stamp_comment_update.example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"], ["form_urlencoded", "query"]]
}`, resourceID, comment)
}

func testAccCheckWallarmRuleDisableStampDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_disable_stamp")
}
