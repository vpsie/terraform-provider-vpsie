package tag_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccTagResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_tag.test", "entity", "ssh_keys"),
					resource.TestCheckResourceAttr("vpsie_tag.test", "tags.#", "2"),
				),
			},
		},
	})
}

const testAccTagConfig = `
resource "vpsie_tag" "test" {
  entity              = "ssh_keys"
  resource_identifier = "replace-with-ssh-key-identifier"
  tags                = ["terraform-acc", "test"]
}
`
