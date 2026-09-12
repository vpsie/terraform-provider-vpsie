package manageddb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccManagedDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedDatabaseConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_managed_database.test", "cluster_name", "terraform-acc-db"),
					resource.TestCheckResourceAttr("vpsie_managed_database.test", "node_count", "1"),
					resource.TestCheckResourceAttrSet("vpsie_managed_database.test", "identifier"),
				),
			},
		},
	})
}

const testAccManagedDatabaseConfig = `
resource "vpsie_managed_database" "test" {
  cluster_name          = "terraform-acc-db"
  datacenter_identifier = "replace-with-datacenter-identifier"
  resource_identifier   = "replace-with-offer-identifier"
  vpc_id                = 1
  project_identifier    = "replace-with-project-identifier"
  node_count            = 1
}
`
