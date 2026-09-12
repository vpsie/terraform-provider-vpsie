package tag_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccTagResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.TestAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_tag.test", "name", "terraform-acc"),
					resource.TestCheckResourceAttr("vpsie_tag.test", "color", "#00ff00"),
					resource.TestCheckResourceAttrSet("vpsie_tag.test", "identifier"),
				),
			},
			{
				ResourceName:      "vpsie_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccTagConfig = `
resource "vpsie_tag" "test" {
  name  = "terraform-acc"
  color = "#00ff00"
}
`
