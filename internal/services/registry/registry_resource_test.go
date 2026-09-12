package registry_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccRegistryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.TestAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_registry.test", "name", "terraform-acc-registry"),
					resource.TestCheckResourceAttrSet("vpsie_registry.test", "identifier"),
				),
			},
		},
	})
}

const testAccRegistryConfig = `
resource "vpsie_registry" "test" {
  name                  = "terraform-acc-registry"
  datacenter_identifier = "replace-with-datacenter-identifier"
  plan_identifier       = "replace-with-plan-identifier"
}
`
