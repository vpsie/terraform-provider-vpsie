package manageddb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccManagedDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.TestAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedDatabaseConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_managed_database.test", "name", "terraform-acc-db"),
					resource.TestCheckResourceAttr("vpsie_managed_database.test", "node_count", "1"),
					resource.TestCheckResourceAttrSet("vpsie_managed_database.test", "identifier"),
				),
			},
		},
	})
}

const testAccManagedDatabaseConfig = `
resource "vpsie_managed_database" "test" {
  name                  = "terraform-acc-db"
  db_type               = "mysql"
  datacenter_identifier = "replace-with-datacenter-identifier"
  plan_id               = 1
  node_count            = 1
}
`
