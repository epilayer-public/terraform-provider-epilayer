package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccSecurityGroupResourceConfig(name, rules string) string {
	return fmt.Sprintf(`
resource "epilayer_security_group" "test" {
  name  = %[1]q
  rules = [%[2]s]
}
`, name, rules)
}

const (
	testAccRuleSSH    = `{ direction = "ingress", protocol = "tcp", port_range_min = 22, port_range_max = 22 }`
	testAccRuleVRRP   = `{ direction = "ingress", protocol = "112" }`
	testAccRuleEgress = `{ direction = "egress", protocol = "all" }`
)

func TestAccSecurityGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing, with rules that leave the ports empty
			{
				Config: providerConfig + testAccSecurityGroupResourceConfig("one", testAccRuleSSH+", "+testAccRuleEgress),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("epilayer_security_group.test", "name", "one"),
					resource.TestCheckResourceAttr("epilayer_security_group.test", "rules.#", "2"),
					resource.TestCheckResourceAttr("epilayer_security_group.test", "rules.0.port_range_min", "22"),
					resource.TestCheckResourceAttr("epilayer_security_group.test", "rules.1.protocol", "all"),
					resource.TestCheckNoResourceAttr("epilayer_security_group.test", "rules.1.port_range_min"),
					resource.TestCheckNoResourceAttr("epilayer_security_group.test", "rules.1.port_range_max"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "epilayer_security_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing, adding a rule for an IP protocol number
			{
				Config: providerConfig + testAccSecurityGroupResourceConfig(
					"two", testAccRuleSSH+", "+testAccRuleVRRP+", "+testAccRuleEgress,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("epilayer_security_group.test", "name", "two"),
					resource.TestCheckResourceAttr("epilayer_security_group.test", "rules.#", "3"),
					resource.TestCheckResourceAttr("epilayer_security_group.test", "rules.1.protocol", "112"),
					resource.TestCheckNoResourceAttr("epilayer_security_group.test", "rules.1.port_range_min"),
					resource.TestCheckNoResourceAttr("epilayer_security_group.test", "rules.1.port_range_max"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
