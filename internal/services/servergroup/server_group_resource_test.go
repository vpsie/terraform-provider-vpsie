package servergroup_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccServerGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerGroupConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_server_group.test", "group_name", "terraform-acc-group"),
					resource.TestCheckResourceAttrSet("vpsie_server_group.test", "identifier"),
				),
			},
			{
				ResourceName:      "vpsie_server_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccServerGroupConfig = `
resource "vpsie_server_group" "test" {
  group_name        = "terraform-acc-group"
  group_description = "created by acceptance test"
  is_distributed    = false
}
`
