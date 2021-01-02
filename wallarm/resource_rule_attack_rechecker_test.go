package wallarm

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccRuleAttackRecheckerCreate(t *testing.T) {
	if os.Getenv("WALLARM_EXTRA_PERMISSIONS") == "" {
		t.Skip("Skipping not test as it requires WALLARM_EXTRA_PERMISSIONS set")
	}

	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_attack_rechecker." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAttackRecheckerCreate(rnd, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "enabled", "false"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleAttackRecheckerCreate(resourceID, enabled string) string {
	return fmt.Sprintf(`
resource "wallarm_application" "%[1]s" {
	name = "tf-testacc-app-rechecker"
	app_id = 97
}

resource "wallarm_rule_attack_rechecker" "%[1]s" {
	enabled = %[2]s
	action {
		point = {
			instance = wallarm_application.%[1]s.app_id
		}
	}
}`, resourceID, enabled)
}
