package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// A release in which `region` is still Required and RequiresReplace.
const lastRegionRequiredVersion = "2.3.1"

const testAccRegionResourcesConfig = `
resource "epilayer_volume" "region_upgrade" {
  name   = "tf-acc-region-upgrade"
  size   = 1
  type   = "hdd"
  %[1]s
}

resource "epilayer_floating_ip" "region_upgrade" {
  name    = "tf-acc-region-upgrade"
  version = "ipv4"
  %[1]s
}

resource "epilayer_private_network" "region_upgrade" {
  name = "tf-acc-region-upgrade"
  %[1]s
}

resource "epilayer_security_group" "region_upgrade" {
  name = "tf-acc-region-upgrade"
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 22
      port_range_max = 22
    },
  ]
  %[1]s
}
`

// testAccRegionResources sets `region` on every resource, or leaves it out if
// region is empty.
func testAccRegionResources(region string) string {
	line := ""
	if region != "" {
		line = fmt.Sprintf("region = %q", region)
	}
	return providerConfig + fmt.Sprintf(testAccRegionResourcesConfig, line)
}

// TestAccRegionDeprecationUpgrade creates resources with a provider
// release that required `region`, then checks that this provider plans no
// changes, both with `region` still set and with it removed from the
// configuration.
func TestAccRegionDeprecationUpgrade(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"epilayer": {
						Source:            "registry.terraform.io/epilayer-public/epilayer",
						VersionConstraint: lastRegionRequiredVersion,
					},
				},
				Config: testAccRegionResources("NORD-NO-KRS-1"),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccRegionResources("NORD-NO-KRS-1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccRegionResources(""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("epilayer_volume.region_upgrade", "region", "NORD-NO-KRS-1"),
					resource.TestCheckResourceAttr("epilayer_floating_ip.region_upgrade", "region", "NORD-NO-KRS-1"),
					resource.TestCheckResourceAttr("epilayer_private_network.region_upgrade", "region", "NORD-NO-KRS-1"),
					resource.TestCheckResourceAttr("epilayer_security_group.region_upgrade", "region", "NORD-NO-KRS-1"),
				),
			},
		},
	})
}
