package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccLoadBalancerResourceConfig(name, region string) string {
	return fmt.Sprintf(`
resource "epilayer_private_network" "test_network" {
  name    = "lb-test-net"
  region  = %[2]q
  cidr_v4 = "10.0.0.0/24"
}

resource "epilayer_load_balancer" "test" {
  name    = %[1]q
  region  = %[2]q
  network = epilayer_private_network.test_network.id
  mode    = "tcp"

  ports = [
    {
      port        = 80
      target_port = 8080
      targets     = ["10.0.0.1", "10.0.0.2"]
    }
  ]

  health_check = {
    protocol = "tcp"
    port     = 8080
  }
}
`, name, region)
}

func TestAccLoadBalancerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccLoadBalancerResourceConfig("lb-one", "NORD-NO-KRS-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("epilayer_load_balancer.test", "name", "lb-one"),
					resource.TestCheckResourceAttr("epilayer_load_balancer.test", "region", "NORD-NO-KRS-1"),
					resource.TestCheckResourceAttr("epilayer_load_balancer.test", "mode", "tcp"),
					resource.TestCheckResourceAttr("epilayer_load_balancer.test", "ports.0.port", "80"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "epilayer_load_balancer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"health_check"},
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccLoadBalancerResourceConfig("lb-two", "NORD-NO-KRS-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("epilayer_load_balancer.test", "name", "lb-two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
